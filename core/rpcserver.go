/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20

Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package core

import (
	"context"
	"encoding/hex"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/SSSOC-CAN/laniakea/api"
	"github.com/SSSOC-CAN/laniakea/errors"
	"github.com/SSSOC-CAN/laniakea/health"
	"github.com/SSSOC-CAN/laniakea/intercept"
	"github.com/SSSOC-CAN/laniakea/lanirpc"
	"github.com/SSSOC-CAN/laniakea/macaroons"
	"github.com/SSSOC-CAN/laniakea/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	e "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

const (
	ErrGRPCMiddlewareNil    = bg.Error("gRPC middleware uninitialized")
	ErrEmptyPermissionsList = bg.Error("empty permissions list")
	ErrInvalidMacEntity     = bg.Error("invalid macaroon permission entity")
	ErrInvalidMacAction     = bg.Error("invalid macaroon permission action")
	ErrDeprecatedAction     = bg.Error("deprecated rpc command")
)

var (
	readPermissions = []bakery.Op{
		{
			Entity: "laniakea",
			Action: "read",
		},
		{
			Entity: "macaroon",
			Action: "read",
		},
		{
			Entity: "plugins",
			Action: "read",
		},
	}
	writePermissions = []bakery.Op{
		{
			Entity: "laniakea",
			Action: "write",
		},
		{
			Entity: "macaroon",
			Action: "generate",
		},
		{
			Entity: "macaroon",
			Action: "write",
		},
		{
			Entity: "plugins",
			Action: "write",
		},
	}
	validActions  = []string{"read", "write", "generate"}
	validEntities = []string{"laniakea", "macaroon", "plugins", macaroons.PermissionEntityCustomURI}
)

// StreamingPluginAPIPermissions returns a map of the command URI and it's assocaited permissions for the streaming Plugin API methods
func StreamingPluginAPIPermission() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/lanirpc.PluginAPI/Subscribe": {{
			Entity: "plugins",
			Action: "read",
		}},
		"/lanirpc.PluginAPI/Command": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/SubscribePluginState": {{
			Entity: "plugins",
			Action: "read",
		}},
	}
}

// MainGrpcServerPermissions returns a map of the command URI and it's associated permissions
func MainGrpcServerPermissions() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/lanirpc.Lani/StopDaemon": {{
			Entity: "laniakea",
			Action: "write",
		}},
		"/lanirpc.Lani/AdminTest": {{
			Entity: "laniakea",
			Action: "read",
		}},
		"/lanirpc.Lani/BakeMacaroon": {{
			Entity: "macaroon",
			Action: "generate",
		}},
		"/lanirpc.PluginAPI/StartRecord": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/StopRecord": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/Subscribe": {{
			Entity: "plugins",
			Action: "read",
		}},
		"/lanirpc.PluginAPI/StartPlugin": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/StopPlugin": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/Command": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/ListPlugins": {{
			Entity: "plugins",
			Action: "read",
		}},
		"/lanirpc.PluginAPI/AddPlugin": {{
			Entity: "plugins",
			Action: "write",
		}},
		"/lanirpc.PluginAPI/GetPlugin": {{
			Entity: "plugins",
			Action: "read",
		}},
		"/lanirpc.PluginAPI/SubscribePluginState": {{
			Entity: "plugins",
			Action: "read",
		}},
		"/lanirpc.Health/Check": {{
			Entity: "laniakea",
			Action: "read",
		}},
	}
}

// RpcServer is a child of the lanirpc.UnimplementedLaniServer struct. Meant to host all related attributes to the rpcserver
type RpcServer struct {
	Active int32
	lanirpc.UnimplementedLaniServer
	interceptor     *intercept.Interceptor
	grpcInterceptor *intercept.GrpcInterceptor
	cfg             *Config
	quit            chan struct{}
	SubLogger       *zerolog.Logger
	macSvc          *macaroons.Service
	Listener        net.Listener
}

// Compile time check to ensure RpcServer implements api.RestProxyService
var _ api.RestProxyService = (*RpcServer)(nil)

// NewRpcServer creates an instance of the GrpcServer struct
func NewRpcServer(interceptor *intercept.Interceptor, config *Config, log *zerolog.Logger) (*RpcServer, error) {
	logger := &NewSubLogger(log, "RPCS").SubLogger
	listener, err := net.Listen("tcp", ":"+strconv.FormatInt(config.GrpcPort, 10))
	if err != nil {
		logger.Error().Msgf("Couldn't open tcp listener on port %v: %v", config.GrpcPort, err)
		return nil, err
	}
	return &RpcServer{
		interceptor: interceptor,
		cfg:         config,
		quit:        make(chan struct{}, 1),
		SubLogger:   logger,
		Listener:    listener,
	}, nil
}

// RegisterWithGrpcServer registers the rpcServer with the root gRPC server.
func (s *RpcServer) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	lanirpc.RegisterLaniServer(grpcServer, s)
	return nil
}

// RegisterWithRestProxy registers the RPC Server with the REST proxy
func (s *RpcServer) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := lanirpc.RegisterLaniHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

// AddMacaroonService adds the macaroon service to the attributes of the RpcServer
func (s *RpcServer) AddMacaroonService(svc *macaroons.Service) {
	s.macSvc = svc
}

// AddGrpcInterceptor adds the grpc middleware to the RpcServer
func (s *RpcServer) AddGrpcInterceptor(i *intercept.GrpcInterceptor) {
	s.grpcInterceptor = i
}

// Start starts the RpcServer subserver
func (s *RpcServer) Start() error {
	s.SubLogger.Info().Msg("Starting RPC server...")
	if ok := atomic.CompareAndSwapInt32(&s.Active, 0, 1); !ok {
		return errors.ErrServiceAlreadyStarted
	}
	s.SubLogger.Info().Msg("RPC server started")
	return nil
}

// Stop stops the rpc sub-server
func (s *RpcServer) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&s.Active, 1, 0); !ok {
		return errors.ErrServiceAlreadyStopped
	}
	close(s.quit)
	err := s.Listener.Close()
	if err != nil {
		s.SubLogger.Error().Msgf("Could not stop listening at %v: %v", s.Listener.Addr(), s.Listener.Close())
		return e.Wrapf(err, "could not stop listening at %v", s.Listener.Addr())
	}
	return nil
}

// StopDaemon will send a shutdown request to the interrupt handler, triggering a graceful shutdown
func (s *RpcServer) StopDaemon(_ context.Context, _ *lanirpc.StopRequest) (*lanirpc.StopResponse, error) {
	s.interceptor.RequestShutdown()
	return &lanirpc.StopResponse{}, nil
}

// AdminTest will return a string only if the client has the admin macaroon
func (s *RpcServer) AdminTest(_ context.Context, _ *lanirpc.AdminTestRequest) (*lanirpc.AdminTestResponse, error) {
	return &lanirpc.AdminTestResponse{Msg: "This is an admin test"}, nil
}

// TestCommand will return a string for any macaroon
func (s *RpcServer) TestCommand(_ context.Context, _ *lanirpc.TestRequest) (*lanirpc.TestResponse, error) {
	return &lanirpc.TestResponse{Msg: "This is a regular test"}, nil
}

// BakeMacaroon bakes a new macaroon based on input permissions and constraints
func (s *RpcServer) BakeMacaroon(ctx context.Context, req *lanirpc.BakeMacaroonRequest) (*lanirpc.BakeMacaroonResponse, error) {
	if s.macSvc == nil {
		return nil, status.Error(codes.Aborted, errors.ErrMacSvcNil.Error())
	}
	if s.grpcInterceptor == nil {
		return nil, status.Error(codes.Aborted, ErrGRPCMiddlewareNil.Error())
	}
	if len(req.Permissions) == 0 {
		return nil, status.Error(codes.InvalidArgument, ErrEmptyPermissionsList.Error())
	}
	perms := make([]bakery.Op, len(req.Permissions))
	for i, op := range req.Permissions {
		if !utils.StrInStrSlice(validEntities, op.Entity) {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidMacEntity.Error())
		}
		if op.Entity == macaroons.PermissionEntityCustomURI {
			allPermissions := s.grpcInterceptor.Permissions()
			if _, ok := allPermissions[op.Action]; !ok {
				return nil, status.Error(codes.InvalidArgument, ErrInvalidMacAction.Error())
			}
		} else if !utils.StrInStrSlice(validActions, op.Action) {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidMacAction.Error())
		}
		perms[i] = bakery.Op{
			Entity: op.Entity,
			Action: op.Action,
		}
	}
	var timeoutSeconds int64
	if req.Timeout > 0 {
		switch req.TimeoutType {
		case lanirpc.TimeoutType_SECOND:
			timeoutSeconds = req.Timeout
		case lanirpc.TimeoutType_MINUTE:
			timeoutSeconds = req.Timeout * int64(60)
		case lanirpc.TimeoutType_HOUR:
			timeoutSeconds = req.Timeout * int64(60) * int64(60)
		case lanirpc.TimeoutType_DAY:
			timeoutSeconds = req.Timeout * int64(60) * int64(60) * int64(24)
		}
	}
	macBytes, err := bakeMacaroons(ctx, s.macSvc, perms, timeoutSeconds, req.Plugins)
	if err != nil {
		return nil, status.Error(codes.Internal, e.Wrap(err, "could not bake macaroon").Error())
	}
	return &lanirpc.BakeMacaroonResponse{
		Macaroon: hex.EncodeToString(macBytes),
	}, nil
}

// SetTemperature is a deprecated Controller API rpc command
func (s *RpcServer) SetTemperature(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// SetPressure is a deprecated Controller API rpc command
func (s *RpcServer) SetPressure(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// StartRecording is a deprecated DataCollector API rpc command
func (s *RpcServer) StartRecording(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// StopRecording is a deprecated DataCollector API rpc command
func (s *RpcServer) StopRecording(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// SubscribeDataStream is a deprecated DataCollector API rpc command
func (s *RpcServer) SubscribeDataStream(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// LoadTestPlan is a deprecated Executor API rpc command
func (s *RpcServer) LoadTestPlan(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// StartTestPlan is a deprecated Executor API rpc command
func (s *RpcServer) StartTestPlan(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// StopTestPlan is a deprecated Executor API rpc command
func (s *RpcServer) StopTestPlan(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

// InsertROIMarker is a deprecated Executor API rpc command
func (s *RpcServer) InsertROIMarker(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return nil, status.Error(codes.Unimplemented, ErrDeprecatedAction.Error())
}

var _ health.RegisteredHealthService = (*RpcServer)(nil)

// Ping implements the health package RegisteredHealthService interface
// TODO:SSSOCPaulCote - This should do more, it should actually probe the service to make sure everything is operating nominally
func (s *RpcServer) Ping(ctx context.Context) error {
	return nil
}
