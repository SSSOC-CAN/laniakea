// +build demo
package controller

import (
	"sync/atomic"

	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc/demorpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/rs/zerolog"
)

type ControllerService {
	demorpc.UnimplementedControllerServer
	BaseControllerService
}

// Compile time check to ensure ControllerService implements api.RestProxyService
var _ api.RestProxyService = (*ControllerService)(nil)
// A compile time check to make sure that ControllerService fully implements the data.Service interface
var _ data.Service = (*ControllerService) (nil)

// NewControllerService creates an instance of the ControllerService struct
func NewControllerService(logger *zerolog.Logger, store *state.Store, _ drivers.DriverConnection) *ControllerService {
	return &ControllerService{
		BaseControllerService{
			stateStore: store,
			Logger: logger,
			name: ControllerName,
		},
	}
}

// Start starts the Controller service. It does NOT start the data recording process
func (s *ControllerService) Start() error {
	s.Logger.Info().Msg("Starting controller service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start controller service. Service already started.")
	}
	s.Logger.Info().Msg("No Connection to a controller is currently active.`")
	s.Logger.Info().Msg("Controller Service successfully started.")
	return nil
}

// Stop stops the Controller service.
func (s *ControllerService) Stop() error {
	s.Logger.Info().Msg("Stopping controller service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop controller service. Service already stopped.")
	}
	if atomic.LoadInt32(&s.Recording) == 1 {
		err := s.stopRecording()
		if err != nil {
			return fmt.Errorf("Could not stop controller service: %v", err)
		}
	}
	s.Logger.Info().Msg("Controller service successfully stopped.")
	return nil
}

// SetTemperature takes a gRPC request and changes the temperature set point accordingly
func (s *ControllerService) SetTemperature(req *fmtrpc.SubscribeDataRequest, updateStream fmtrpc.DataCollector_SubscribeDataStreamServer) error {

}