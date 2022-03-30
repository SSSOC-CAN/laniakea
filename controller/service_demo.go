// +build demo

package controller

import (
	"context"
	e "errors"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc/demorpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/grpc"
)

const (
	ErrNegativeRate = errors.Error("rates cannot be negative")
	ErrTelNotRecoring = errors.Error("telemetry service not yet recording")
)

type ControllerService struct {
	demorpc.UnimplementedControllerServer
	BaseControllerService
}

// Compile time check to ensure ControllerService implements api.RestProxyService
var _ api.RestProxyService = (*ControllerService)(nil)
// A compile time check to make sure that ControllerService fully implements the data.Service interface
var _ data.Service = (*ControllerService) (nil)

// NewControllerService creates an instance of the ControllerService struct
func NewControllerService(logger *zerolog.Logger, rtdStore *state.Store, ctrlStore *state.Store, _ drivers.DriverConnection) *ControllerService {
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
	s.Logger.Info().Msg("Controller service successfully stopped.")
	return nil
}

// Name satisfies the data.Service interface
func (s *ControllerService) Name() string {
	return s.name
}

// RegisterWithGrpcServer registers the gRPC server to the ControllerService
func (s *ControllerService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	demorpc.RegisterControllerServer(grpcServer, s)
	return nil
}

// RegisterWithRestProxy registers the ControllerService with the REST proxy
func(s *ControllerService) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := demorpc.RegisterControllerHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

// SetTemperature takes a gRPC request and changes the temperature set point accordingly
func (s *ControllerService) SetTemperature(req *demorpc.SetTempRequest, updateStream demorpc.Controller_SetTemperatureServer) error {
	if req.TempSetPoint < float64(0) {
		return ErrNegativeRate
	}
	currentState := s.rtdStateStore.GetState()
	RTD, ok := currentState.(data.InitialRtdState)
	if !ok {
		return state.ErrInvalidStateType
	}
	if RTD.TelPollingInterval == int64(0) {
		return ErrTelNotRecoring
	}
	var (
		actualTempChangeRate float64
		intervalCnt			 int
		totalIntervals		 int
		changeRateSlice []struct{
			Rate 	 float64
			SetPoint float64
		}
		lastTempSetPoint float64
	)
	if req.TempChangeRate == float64(0) {
		err := s.ctrlStateStore.Dispatch(updateTempSetPointAction(req.TempSetPoint))
		if err != nil {
			return err
		}
		if err := updateStream.Send(&demorpc.SetTempResponse{
			CurrentTempSetPoint: req.TempSetPoint,
			CurrentAvgTemp: RTD.AverageTemperature,
		}); err != nil {
			return err
		}
		lastTempSetPoint = req.TempSetPoint
	} else {
		lastTempSetPoint = RTD.AverageTemperature
		// convert change rate from degree C/min to C/Polling Interval
		actualTempChangeRate = req.TempChangeRate * float64(RTD.TelPollingInterval)/float64(60)
		if req.TempSetPoint < RTD.AverageTemperature {
			actualTempChangeRate *= float64(-1)
		}
		lastFiveRates := []struct{
			Intervals int
			Rate	  float64
		}{}
		var cumSum int
		for i := 0; i < 4; i ++ {
			if i == 0 {
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/math.Abs(actualTempChangeRate)/float64(2)),
					Rate: actualTempChangeRate/float64(2),
				})
				cumSum += int(0.25/math.Abs(actualTempChangeRate)/float64(2))
			} else {
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/lastFiveRates[i-1].Rate/float64(2)),
					Rate: lastFiveRates[i-1].Rate/float64(2),
				})
				cumSum += int(0.25/math.Abs(lastFiveRates[i-1].Rate)/float64(2))
			}
		}
		for i := 3; i > -1; i-- {
			for j := 0; j < lastFiveRates[i].Intervals; j++ {
				changeRateSlice = append(changeRateSlice, struct{
					Rate float64
					SetPoint float64
				}{
					Rate: lastFiveRates[i].Rate,
					SetPoint: RTD.AverageTemperature+float64(1),
				})
			}
		} 
		totalIntervals = 2*cumSum+int((math.Abs(req.TempSetPoint-RTD.AverageTemperature)-float64(2))/math.Abs(actualTempChangeRate))
		s.Logger.Debug().Msg(fmt.Sprintf("intervalCnt: %v\nreq: %v\ntotalIntervals: %v\ncumSum: %v", intervalCnt, req, totalIntervals, cumSum))
		for i := 0; i < totalIntervals-(2*cumSum); i++ {
			changeRateSlice = append(changeRateSlice, struct{
				Rate float64
				SetPoint float64
			}{
				Rate: actualTempChangeRate,
				SetPoint: req.TempSetPoint-float64(1),
			})
		}
		// s.Logger.Debug().Msg(fmt.Sprintf("intervalCnt: %v\nreq: %v", intervalCnt, req))
		for _, r := range lastFiveRates {
			for j := 0; j < r.Intervals; j++ {
				changeRateSlice = append(changeRateSlice, struct{
					Rate float64
					SetPoint float64
				}{
					Rate: r.Rate,
					SetPoint: req.TempSetPoint,
				})
			}
		}
		// s.Logger.Debug().Msg(fmt.Sprintf("intervalCnt: %v\nreq: %v", intervalCnt, req))
	}
	changeRateSlice = append(changeRateSlice, struct{
		Rate float64
		SetPoint float64
	}{
		Rate: float64(0),
		SetPoint: req.TempSetPoint,
	})
	rand.Seed(time.Now().UnixNano())
	subscriberName := s.name+utils.RandSeq(10)
	updateChan, unsub := s.rtdStateStore.Subscribe(subscriberName)
	cleanUp := func() {
		unsub(s.rtdStateStore, subscriberName)
	}
	defer cleanUp()
	for {
		select {	
		case <-updateChan:
			currentState := s.rtdStateStore.GetState()
			RTD, ok := currentState.(data.InitialRtdState)
			if !ok {
				return state.ErrInvalidStateType
			}
			if RTD.RealTimeData.Source == "TEL" {
				if err := updateStream.Send(&demorpc.SetTempResponse{
					CurrentTempSetPoint: lastTempSetPoint,
					CurrentAvgTemp: RTD.AverageTemperature,
					CurrentTempChangeRate: changeRateSlice[intervalCnt].Rate,
				}); err != nil {
					return err
				}
				// Update set point
				lastTempSetPoint = lastTempSetPoint+changeRateSlice[intervalCnt].Rate
				err := s.ctrlStateStore.Dispatch(updateTempSetPointAction(lastTempSetPoint))
				if err != nil {
					return err
				}
				if intervalCnt == totalIntervals {
					return nil
				}
				intervalCnt += 1
			}
		case <-updateStream.Context().Done():
			if e.Is(updateStream.Context().Err(), context.Canceled) {
				return nil
			}
			return updateStream.Context().Err()
		}
	}
}

// SetPressure takes a gRPC request and changes the pressure set point accordingly
func (s *ControllerService) SetPressure(req *demorpc.SetPresRequest, updateStream demorpc.Controller_SetPressureServer) error {
	// check if we have negative rate
	if req.PressureSetPoint < float64(0) {
		return ErrNegativeRate
	}
	currentState := s.rtdStateStore.GetState()
	RTD, ok := currentState.(data.InitialRtdState)
	if !ok {
		return state.ErrInvalidStateType
	}
	if RTD.TelPollingInterval == int64(0) {
		return ErrTelNotRecoring
	}
	var (
		actualPresChangeRate float64
		intervalCnt			 int
		totalIntervals		 int
		changeRateSlice []struct{
			Rate 	 float64
			SetPoint float64
		}
		lastPresSetPoint float64
	)
	if req.PressureChangeRate == float64(0) {
		err := s.ctrlStateStore.Dispatch(updatePresSetPointAction(req.PressureSetPoint))
		if err != nil {
			return err
		}
		if err := updateStream.Send(&demorpc.SetPresResponse{
			CurrentPressureSetPoint: req.PressureSetPoint,
			CurrentAvgPressure: RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value,
		}); err != nil {
			return err
		}
		lastPresSetPoint = req.PressureSetPoint
	} else {
		lastPresSetPoint = RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value
		// convert change rate from degree Torr/min to Torr/Polling Interval
		actualPresChangeRate = req.PressureChangeRate * float64(RTD.TelPollingInterval)/float64(60)
		if req.PressureSetPoint < RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value {
			actualPresChangeRate *= float64(-1)
		}
		lastFiveRates := []struct{
			Intervals int
			Rate	  float64
		}{}
		var cumSum int
		for i := 0; i < 4; i ++ {
			if i == 0 {
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/actualPresChangeRate/float64(2)),
					Rate: actualPresChangeRate/float64(2),
				})
				cumSum += int(0.25/actualPresChangeRate/float64(2))
			} else {
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/lastFiveRates[i-1].Rate/float64(2)),
					Rate: lastFiveRates[i-1].Rate/float64(2),
				})
				cumSum += int(0.25/lastFiveRates[i-1].Rate/float64(2))
			}
		}
		for i := 3; i > -1; i-- {
			for j := 0; j < lastFiveRates[i].Intervals; j++ {
				changeRateSlice = append(changeRateSlice, struct{
					Rate float64
					SetPoint float64
				}{
					Rate: lastFiveRates[i].Rate,
					SetPoint: RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value+float64(1),
				})
			}
		} 
		totalIntervals = 2*cumSum+int((math.Abs(req.PressureSetPoint-RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value)-float64(2))/actualPresChangeRate)
		for i := 0; i < totalIntervals-(2*cumSum); i++ {
			changeRateSlice = append(changeRateSlice, struct{
				Rate float64
				SetPoint float64
			}{
				Rate: actualPresChangeRate,
				SetPoint: req.PressureSetPoint-float64(1),
			})
		}
		for _, r := range lastFiveRates {
			for j := 0; j < r.Intervals; j++ {
				changeRateSlice = append(changeRateSlice, struct{
					Rate float64
					SetPoint float64
				}{
					Rate: r.Rate,
					SetPoint: req.PressureSetPoint,
				})
			}
		}
	}
	changeRateSlice = append(changeRateSlice, struct{
		Rate float64
		SetPoint float64
	}{
		Rate: float64(0),
		SetPoint: req.PressureSetPoint,
	})
	rand.Seed(time.Now().UnixNano())
	subscriberName := s.name+utils.RandSeq(10)
	updateChan, unsub := s.rtdStateStore.Subscribe(subscriberName)
	cleanUp := func() {
		unsub(s.rtdStateStore, subscriberName)
	}
	defer cleanUp()
	for {
		select {	
		case <-updateChan:
			currentState := s.rtdStateStore.GetState()
			RTD, ok := currentState.(data.InitialRtdState)
			if !ok {
				return state.ErrInvalidStateType
			}
			if RTD.RealTimeData.Source == "TEL" {
				if err := updateStream.Send(&demorpc.SetPresResponse{
					CurrentPressureSetPoint: lastPresSetPoint,
					CurrentAvgPressure: RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value,
					CurrentPressureChangeRate: changeRateSlice[intervalCnt].Rate,
				}); err != nil {
					return err
				}
				// Update set point
				lastPresSetPoint = lastPresSetPoint+changeRateSlice[intervalCnt].Rate
				err := s.ctrlStateStore.Dispatch(updatePresSetPointAction(lastPresSetPoint))
				if err != nil {
					return err
				}
				if intervalCnt == totalIntervals {
					return nil
				}
				intervalCnt += 1
			}
		case <-updateStream.Context().Done():
			if e.Is(updateStream.Context().Err(), context.Canceled) {
				return nil
			}
			return updateStream.Context().Err()
		}
	}
}