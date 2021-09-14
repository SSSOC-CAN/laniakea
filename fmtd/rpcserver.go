package fmtd

import (
	"context"
	"net"
	"google.golang.org/grpc"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
)

type GrpcServer struct {
	started int32
	shutdown int32
	fmtrpc.UnimplementedFmtServer
	interceptor intercept.Interceptor
}

// NewGrpcServer creates an instance of the GrpcServer struct
func NewGrpcServer(interceptor intercept.Interceptor) (GrpcServer, error) {
	return GrpcServer{interceptor: interceptor}, nil
}

// RegisterWithGrpcServer registers the rpcServer with the root gRPC server.
func (s *GrpcServer) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterFmtServer(grpcServer, s)
	return nil
}

// Start starts the GrpcServer subserver
func (s *GrpcServer) Start(port string) (error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	err = s.RegisterWithGrpcServer(grpcServer)
	if err != nil {
		return err
	}
	if err := grpcServer.Serve(listener); err != nil {
		return err
	}
	return nil
}

// StopDaemon will send a shutdown request to the interrupt handler, triggering a graceful shutdown
func (s *GrpcServer) StopDaemon(_ context.Context, _*fmtrpc.StopRequest) (*fmtrpc.StopResponse, error) {
	s.interceptor.RequestShutdown()
	return &fmtrpc.StopResponse{}, nil
}