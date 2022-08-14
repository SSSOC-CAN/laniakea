package health

import (
	"context"
	"time"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	bg "github.com/SSSOCPaulCote/blunderguard"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrHealthServiceAlreadyRegistered = bg.Error("health service already registered")
	ErrUnregisteredHealthService      = bg.Error("unregistered health service")
)

var (
	defaultCheckTimeout time.Duration = 5 * time.Second
)

type HealthService struct {
	fmtrpc.UnimplementedHealthServer
	registeredServices map[string]RegisteredHealthService
}

// NewHealthService instantiates a new HealthService
func NewHealthService() *HealthService {
	return &HealthService{
		registeredServices: make(map[string]RegisteredHealthService),
	}
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

// RegisterHealthService registers a given service that we want to perform health checks on
func (h *HealthService) RegisterHealthService(name string, s RegisteredHealthService) error {
	if _, ok := h.registeredServices[name]; ok {
		return ErrHealthServiceAlreadyRegistered
	}
	h.registeredServices[name] = s
	return nil
}

// Check is the gRPC command to perform the health check on given API and plugin services
func (h *HealthService) Check(ctx context.Context, req *fmtrpc.HealthRequest) (*fmtrpc.HealthResponse, error) {
	if req.Service == "" || req.Service == "all" {
		// perform health checks on all registered services
		statuses := []*fmtrpc.HealthUpdate{}
		for name, service := range h.registeredServices {
			newCtx, _ := context.WithTimeout(ctx, defaultCheckTimeout)
			errChan := make(chan error)
			go func(errChan chan error) {
				errChan <- service.Ping(newCtx)
			}(errChan)
			status := &fmtrpc.HealthUpdate{
				Name:  name,
				State: fmtrpc.HealthUpdate_SERVING,
			}
			select {
			case err := <-errChan:
				if err != nil {
					status.State = fmtrpc.HealthUpdate_NOT_SERVING
				}
			case <-newCtx.Done():
				status.State = fmtrpc.HealthUpdate_UNKNOWN
			}
			statuses = append(statuses, status)
		}
		return &fmtrpc.HealthResponse{
			Status: statuses,
		}, nil
	}
	if _, ok := h.registeredServices[req.Service]; !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregisteredHealthService.Error())
	}
	service := h.registeredServices[req.Service]
	newCtx, _ := context.WithTimeout(ctx, defaultCheckTimeout)
	errChan := make(chan error)
	go func(errChan chan error) {
		errChan <- service.Ping(newCtx)
	}(errChan)
	status := &fmtrpc.HealthUpdate{
		Name:  req.Service,
		State: fmtrpc.HealthUpdate_SERVING,
	}
	select {
	case err := <-errChan:
		if err != nil {
			status.State = fmtrpc.HealthUpdate_NOT_SERVING
		}
	case <-newCtx.Done():
		status.State = fmtrpc.HealthUpdate_UNKNOWN
	}
	return &fmtrpc.HealthResponse{Status: []*fmtrpc.HealthUpdate{status}}, nil
}
