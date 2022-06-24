// +build !demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"context"
	"sync/atomic"

	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOCPaulCote/gux"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type ControllerService struct {
	BaseControllerService
}

// Compile time check to ensure ControllerService implements api.RestProxyService
var _ api.RestProxyService = (*ControllerService)(nil)
// A compile time check to make sure that ControllerService fully implements the data.Service interface
var _ data.Service = (*ControllerService) (nil)

// NewControllerService creates an instance of the ControllerService struct
func NewControllerService(logger *zerolog.Logger, rtdStore *gux.Store, ctrlStore *gux.Store, _ drivers.DriverConnection) *ControllerService {
	return &ControllerService{
		BaseControllerService: BaseControllerService{
			rtdStateStore: rtdStore,
			ctrlStateStore: ctrlStore,
			Logger: logger,
			name: ControllerName,
		},
	}
}

// Start starts the Controller service. It does NOT start the data recording process
func (s *ControllerService) Start() error {
	s.Logger.Info().Msg("Starting controller service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return errors.ErrServiceAlreadyStarted
	}
	s.Logger.Info().Msg("No Connection to a controller is currently active.`")
	s.Logger.Info().Msg("Controller Service successfully started.")
	return nil
}

// Stop stops the Controller service.
func (s *ControllerService) Stop() error {
	s.Logger.Info().Msg("Stopping controller service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return errors.ErrServiceAlreadyStopped
	}
	s.Logger.Info().Msg("Controller service successfully stopped.")
	return nil
}

// Name satisfies the data.Service interface
func (s *ControllerService) Name() string {
	return s.name
}

// RegisterWithGrpcServer registers the gRPC server to the ControllerService
func (s *ControllerService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	return nil
}

// RegisterWithRestProxy registers the ControllerService with the REST proxy
func(s *ControllerService) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	return nil
}