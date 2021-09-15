package fmtd

import (
	"context"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"google.golang.org/grpc"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
)

// RpcServer is a child of the fmtrpc.UnimplementedFmtServer struct. Meant to host all related attributes to the rpcserver
type RpcServer struct {
	started int32
	shutdown int32
	fmtrpc.UnimplementedFmtServer
	interceptor intercept.Interceptor
	Grpc_server	*grpc.Server
	cfg *Config
	quit chan struct{}
}

// NewRpcServer creates an instance of the GrpcServer struct
func NewRpcServer(interceptor intercept.Interceptor, config *Config) (RpcServer, error) {
	return RpcServer{
		interceptor: interceptor,
		Grpc_server: grpc.NewServer(),
		cfg: config,
		quit: make(chan struct{}, 1),
	}, nil
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
		return err
	}
	err = s.RegisterWithGrpcServer(s.Grpc_server)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(lis *net.Listener) {
		wg.Done()
		_ = s.Grpc_server.Serve(listener)
	}(&listener)
	wg.Wait()
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