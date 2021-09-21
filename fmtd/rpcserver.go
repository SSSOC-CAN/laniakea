package fmtd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// RpcServer is a child of the fmtrpc.UnimplementedFmtServer struct. Meant to host all related attributes to the rpcserver
type RpcServer struct {
	started int32
	shutdown int32
	fmtrpc.UnimplementedFmtServer
	interceptor *intercept.Interceptor
	GrpcServer	*grpc.Server
	cfg *Config
	quit chan struct{}
	SubLogger *zerolog.Logger
}

// NewRpcServer creates an instance of the GrpcServer struct
func NewRpcServer(interceptor *intercept.Interceptor, config *Config, log *zerolog.Logger) (RpcServer, error) {
	return RpcServer{
		interceptor: interceptor,
		cfg: config,
		quit: make(chan struct{}, 1),
		SubLogger: &NewSubLogger(log, "RPCS").SubLogger,
	}, nil
}

// AddGrpcServer adds a gRPC server to the attributes of the RpcServer struct
func (r *RpcServer) AddGrpcServer(server *grpc.Server) {
	r.GrpcServer = server
}

// RegisterWithGrpcServer registers the rpcServer with the root gRPC server.
func (s *RpcServer) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterFmtServer(grpcServer, s)
	return nil
}

// Start starts the RpcServer subserver
func (s *RpcServer) Start() (error) {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return nil
	}
	listener, err := net.Listen("tcp", ":"+strconv.FormatInt(s.cfg.GrpcPort, 10))
	if err != nil {
		s.SubLogger.Error().Msg(fmt.Sprintf("Couldn't open tcp listener on port %v: %v", s.cfg.GrpcPort, err))
		return err
	}
	err = s.RegisterWithGrpcServer(s.GrpcServer)
	if err != nil {
		s.SubLogger.Error().Msg(fmt.Sprintf("Couldn't register with gRPC server: %v", err))
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(lis *net.Listener) {
		wg.Done()
		_ = s.GrpcServer.Serve(listener)
	}(&listener)
	wg.Wait()
	s.SubLogger.Info().Msg(fmt.Sprintf("gRPC listening on port %v", s.cfg.GrpcPort))
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