package health

import (
	"context"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type HealthService struct {
	fmtrpc.UnimplementedHealthServer
}

// NewHealthService instantiates a new HealthService
func NewHealthService() *HealthService {
	return &HealthService{}
}

// RegisterWithGrpcServer registers the health service with the gRPC server
func (h *HealthService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterHealthServer(grpcServer, h)
	return nil
}

// RegisterWithRestProxy registers the health service with the REST proxy server
func (h *HealthService) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	return fmtrpc.RegisterHealthHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
}

// Check is the gRPC command to perform the health check on given API and plugin services
func (h *HealthService) Check(ctx context.Context, req *fmtrpc.HealthRequest) (*fmtrpc.HealthResponse, error) {
	return nil, nil
}
