/*
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
package fmtd

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	readPermissions = []bakery.Op{
		{
			Entity: "fmtd",
			Action: "read",
		},
		{
			Entity: "macaroon",
			Action: "read",
		},
		{
			Entity: "tpex",
			Action: "read",
		},
	}
	writePermissions = []bakery.Op{
		{
			Entity: "fmtd",
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
			Entity: "tpex",
			Action: "write",
		},
	}
	validActions = []string{"read", "write"}
	validEntities = []string{"fmtd", "macaroon", "tpex", macaroons.PermissionEntityCustomURI}
)

// MainGrpcServerPermissions returns a map of the command URI and it's associated permissions
func MainGrpcServerPermissions() map[string][]bakery.Op {
	return map[string][]bakery.Op{
		"/fmtrpc.Fmt/StopDaemon": {{
			Entity: "fmtd",
			Action:	"write",
		}},
		"/fmtrpc.Fmt/AdminTest": {{
			Entity: "fmtd",
			Action: "read",
		}},
		"/fmtrpc.DataCollector/StartRecording": {{
			Entity: "fmtd",
			Action: "write",
		}},
		"/fmtrpc.DataCollector/StopRecording": {{
			Entity: "fmtd",
			Action: "write",
		}},
		"/fmtrpc.DataCollector/SubscribeDataStream": {{
			Entity: "fmtd",
			Action: "read",
		}},
		"/fmtrpc.DataCollector/DownloadHistoricalData": {{
			Entity: "fmtd",
			Action: "read",
		}},
		"/fmtrpc.TestPlanExecutor/LoadTestPlan": {{
			Entity: "tpex",
			Action: "write",
		}},
		"/fmtrpc.TestPlanExecutor/StartTestPlan": {{
			Entity: "tpex",
			Action: "write",
		}},
		"/fmtrpc.TestPlanExecutor/StopTestPlan": {{
			Entity: "tpex",
			Action: "write",
		}},
		"/fmtrpc.TestPlanExecutor/InsertROIMarker": {{
			Entity: "tpex",
			Action:	"write",
		}},
		"/fmtrpc.Fmt/BakeMacaroon": {{
			Entity: "macaroon",
			Action: "write",
		}},
	}
}

// RpcServer is a child of the fmtrpc.UnimplementedFmtServer struct. Meant to host all related attributes to the rpcserver
type RpcServer struct {
	started 						int32
	shutdown 						int32
	fmtrpc.UnimplementedFmtServer
	interceptor 					*intercept.Interceptor
	cfg 							*Config
	quit 							chan struct{}
	SubLogger 						*zerolog.Logger
	macSvc							*macaroons.Service
	Listener						net.Listener
}

// NewRpcServer creates an instance of the GrpcServer struct
func NewRpcServer(interceptor *intercept.Interceptor, config *Config, log *zerolog.Logger) (*RpcServer, error) {
	logger := &NewSubLogger(log, "RPCS").SubLogger
	listener, err := net.Listen("tcp", ":"+strconv.FormatInt(config.GrpcPort, 10))
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Couldn't open tcp listener on port %v: %v", config.GrpcPort, err))
		return nil, err
	}
	return &RpcServer{
		interceptor: interceptor,
		cfg: config,
		quit: make(chan struct{}, 1),
		SubLogger: logger,
		Listener: listener,
	}, nil
}

// RegisterWithGrpcServer registers the rpcServer with the root gRPC server.
func (s *RpcServer) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterFmtServer(grpcServer, s)
	return nil
}

// RegisterWithRestProxy registers the RPC Server with the REST proxy
func(s *RpcServer) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := fmtrpc.RegisterFmtHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

// func AddMacaroonService adds the macaroon service to the attributes of the RpcServer
func (s *RpcServer) AddMacaroonService(svc *macaroons.Service) {
	s.macSvc = svc
}

// Start starts the RpcServer subserver
func (s *RpcServer) Start() (error) {
	s.SubLogger.Info().Msg("Starting RPC server...")
	if atomic.AddInt32(&s.started, 1) != 1 {
		return fmt.Errorf("Could not start RPC server: server already started")
	}
	s.SubLogger.Info().Msg("RPC server started")
	return nil
}

// Stop stops the rpc sub-server
func (s *RpcServer) Stop() (error) {
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		return nil
	}
	close(s.quit)
	return nil
}

// StopDaemon will send a shutdown request to the interrupt handler, triggering a graceful shutdown
func (s *RpcServer) StopDaemon(_ context.Context, _*fmtrpc.StopRequest) (*fmtrpc.StopResponse, error) {
	s.interceptor.RequestShutdown()
	return &fmtrpc.StopResponse{}, nil
}

// AdminTest will return a string only if the client has the admin macaroon
func (s *RpcServer) AdminTest(_ context.Context, _*fmtrpc.AdminTestRequest) (*fmtrpc.AdminTestResponse, error) {
	return &fmtrpc.AdminTestResponse{Msg: "This is an admin test"}, nil
}

// TestCommand will return a string for any macaroon
func (s *RpcServer) TestCommand(_ context.Context, _*fmtrpc.TestRequest) (*fmtrpc.TestResponse, error) {
	return &fmtrpc.TestResponse{Msg: "This is a regular test"}, nil
}

// BakeMacaroon bakes a new macaroon based on input permissions and constraints
func (s *RpcServer) BakeMacaroon(ctx context.Context, req *fmtrpc.BakeMacaroonRequest) (*fmtrpc.BakeMacaroonResponse, error) {
	if s.macSvc == nil {
		return nil, fmt.Errorf("Could not bake macaroon: macaroon service not initialized")
	}
	if len(req.Permissions) == 0 {
		return nil, fmt.Errorf("Could not bake macaroon: empty permissions list")
	}
	perms := make([]bakery.Op, len(req.Permissions))
	for i, op := range req.Permissions {
		if !stringInSlice(op.Entity, validEntities) {
			return nil, fmt.Errorf("Could not bake macaroon: invalid permission entity")
		}
		if op.Entity == macaroons.PermissionEntityCustomURI {
			allPermissions := MainGrpcServerPermissions()
			if _, ok := allPermissions[op.Action]; !ok {
				return nil, fmt.Errorf("Could not bake macaroon: %s is not a valid action", op.Action)
			}
		} else if !stringInSlice(op.Action, validActions) {
			return nil, fmt.Errorf("Could not bake macaroon: invalid permission action")
		}
		perms[i] = bakery.Op{
			Entity: op.Entity,
			Action: op.Action,
		}
	}
	noTimeout := true
	var timeoutSeconds int64
	if req.Timeout != 0 {
		noTimeout = false
		switch req.TimeoutType {
		case fmtrpc.TimeoutType_SECOND:
			timeoutSeconds = req.Timeout
		case fmtrpc.TimeoutType_MINUTE:
			timeoutSeconds = req.Timeout * int64(60)
		case fmtrpc.TimeoutType_HOUR:
			timeoutSeconds = req.Timeout * int64(60) * int64(60)
		case fmtrpc.TimeoutType_DAY:
			timeoutSeconds = req.Timeout * int64(60) * int64(60) * int64(24)
		}
	}
	macBytes, err := bakeMacaroons(ctx, s.macSvc, perms, noTimeout, timeoutSeconds)
	if err != nil {
		return nil, fmt.Errorf("Could not bake macaroon: %v", err)
	}
	return &fmtrpc.BakeMacaroonResponse{
		Macaroon: hex.EncodeToString(macBytes),
	}, nil
}

// stringInSlice checks if a string "a" is in the slice 
func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}