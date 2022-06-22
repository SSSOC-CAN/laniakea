// +build demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"context"
	e "errors"
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
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/SSSOCPaulCote/gux"
	"google.golang.org/grpc"
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
func NewControllerService(logger *zerolog.Logger, rtdStore *gux.Store, ctrlStore *gux.Store, _ drivers.DriverConnection) *ControllerService {
	return &ControllerService{
		BaseControllerService: BaseControllerService{
			ctrlState: waitingToStart,
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

// setWaitingToStart sets the controller state to waitingToStart
func (s *ControllerService) setWaitingToStart() {
	s.Lock()
	defer s.Unlock()
	s.ctrlState = waitingToStart
}

// setInUse sets the controller state to inUse
func (s *ControllerService) setInUse() {
	s.Lock()
	defer s.Unlock()
	s.ctrlState = inUse
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
	if s.ctrlState != waitingToStart {
		return errors.ErrCtrllerInUse
	}
	if req.TempChangeRate < float64(0) {
		return errors.ErrNegativeRate
	}
	s.setInUse()
	defer s.setWaitingToStart()
	currentState := s.rtdStateStore.GetState()
	RTD, ok := currentState.(data.InitialRtdState)
	if !ok {
		return gux.ErrInvalidStateType
	}
	if RTD.TelPollingInterval == int64(0) {
		return errors.ErrTelNotRecoring
	}
	currentState = s.ctrlStateStore.GetState()
	ctrl, ok := currentState.(data.InitialCtrlState)
	if !ok {
		return gux.ErrInvalidStateType
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
		lastTempSetPoint = ctrl.TemperatureSetPoint
		setPtHelp := float64(1)
		// convert change rate from degree C/min to C/Polling Interval
		actualTempChangeRate = req.TempChangeRate * float64(RTD.TelPollingInterval)/float64(60)
		if req.TempSetPoint < RTD.AverageTemperature {
			actualTempChangeRate *= float64(-1)
			setPtHelp *= float64(-1)
		}
		lastFiveRates := []struct{
			Intervals int
			Rate	  float64
		}{}
		var (
			cumSum int
			rate   float64
			inter  int
		)
		for i := 0; i < 4; i++ {
			if i == 0 {
				rate = actualTempChangeRate/float64(2)
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/math.Abs(rate)),
					Rate: rate,
				})
			} else {
				rate = lastFiveRates[i-1].Rate/float64(2)
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/math.Abs(rate)),
					Rate: rate,
				})
			}
			inter = int(0.25/math.Abs(rate))
			cumSum += inter
		}
		for i := 3; i > -1; i-- {
			for j := 0; j < lastFiveRates[i].Intervals; j++ {
				changeRateSlice = append(changeRateSlice, struct{
					Rate float64
					SetPoint float64
				}{
					Rate: lastFiveRates[i].Rate,
					SetPoint: ctrl.TemperatureSetPoint+setPtHelp,
				})
			}
		}
		placeHolder := math.Abs(req.TempSetPoint-ctrl.TemperatureSetPoint)-float64(2)
		totalIntervals = int(placeHolder/math.Abs(actualTempChangeRate))
		totalIntervals = 2*cumSum+totalIntervals
		for i := 0; i < totalIntervals-(2*cumSum); i++ {
			changeRateSlice = append(changeRateSlice, struct{
				Rate float64
				SetPoint float64
			}{
				Rate: actualTempChangeRate,
				SetPoint: req.TempSetPoint-setPtHelp,
			})
		}
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
				return gux.ErrInvalidStateType
			}
			if RTD.RealTimeData.Source == "TEL" {
				if err := updateStream.Send(&demorpc.SetTempResponse{
					CurrentTempSetPoint: changeRateSlice[intervalCnt].SetPoint,
					CurrentAvgTemp: RTD.AverageTemperature,
					CurrentTempChangeRate: changeRateSlice[intervalCnt].Rate*float64(60)/float64(RTD.TelPollingInterval),
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
	if s.ctrlState != waitingToStart {
		return errors.ErrCtrllerInUse
	}
	// check if we have negative rate
	if req.PressureChangeRate < float64(0) {
		return errors.ErrNegativeRate
	}
	if req.PressureSetPoint < float64(0) {
		return errors.ErrNegativePressure
	}
	s.setInUse()
	defer s.setWaitingToStart()
	currentState := s.rtdStateStore.GetState()
	RTD, ok := currentState.(data.InitialRtdState)
	if !ok {
		return gux.ErrInvalidStateType
	}
	if RTD.TelPollingInterval == int64(0) {
		return errors.ErrTelNotRecoring
	}
	currentState = s.ctrlStateStore.GetState()
	ctrl, ok := currentState.(data.InitialCtrlState)
	if !ok {
		return gux.ErrInvalidStateType
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
		setPtHelp := float64(1)
		// convert change rate from degree Torr/min to Torr/Polling Interval
		actualPresChangeRate = req.PressureChangeRate * float64(RTD.TelPollingInterval)/float64(60)
		if req.PressureSetPoint < RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value {
			actualPresChangeRate *= float64(-1)
			setPtHelp *= float64(-1)
		}
		lastFiveRates := []struct{
			Intervals int
			Rate	  float64
		}{}
		var (
			cumSum int
			rate   float64
			inter  int
		)
		for i := 0; i < 4; i ++ {
			if i == 0 {
				rate = actualPresChangeRate/float64(2)
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/math.Abs(rate)),
					Rate: rate,
				})
			} else {
				rate = lastFiveRates[i-1].Rate/float64(2)
				lastFiveRates = append(lastFiveRates, struct{
					Intervals int
					Rate	  float64
				}{
					Intervals: int(0.25/math.Abs(rate)),
					Rate: rate,
				})
			}
			inter = int(0.25/math.Abs(rate))
			cumSum += inter
		}
		for i := 3; i > -1; i-- {
			for j := 0; j < lastFiveRates[i].Intervals; j++ {
				changeRateSlice = append(changeRateSlice, struct{
					Rate float64
					SetPoint float64
				}{
					Rate: lastFiveRates[i].Rate,
					SetPoint: ctrl.PressureSetPoint+setPtHelp,
				})
			}
		}
		placeHolder := math.Abs(req.PressureSetPoint-ctrl.PressureSetPoint)-float64(2)
		totalIntervals = int(placeHolder/math.Abs(actualPresChangeRate)) 
		totalIntervals = 2*cumSum+totalIntervals
		for i := 0; i < totalIntervals-(2*cumSum); i++ {
			changeRateSlice = append(changeRateSlice, struct{
				Rate float64
				SetPoint float64
			}{
				Rate: actualPresChangeRate,
				SetPoint: req.PressureSetPoint-setPtHelp,
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
				return gux.ErrInvalidStateType
			}
			if RTD.RealTimeData.Source == "TEL" {
				if err := updateStream.Send(&demorpc.SetPresResponse{
					CurrentPressureSetPoint: changeRateSlice[intervalCnt].SetPoint,
					CurrentAvgPressure: RTD.RealTimeData.Data[drivers.TelemetryPressureChannel].Value,
					CurrentPressureChangeRate: changeRateSlice[intervalCnt].Rate*float64(60)/float64(RTD.TelPollingInterval),
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