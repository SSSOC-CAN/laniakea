// +build demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package telemetry

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
)

var (
	ctrlInitialState = data.InitialCtrlState{
		PressureSetPoint: 760.0,
		TemperatureSetPoint: 25.0,
	}
	ctrlReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialCtrlState)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "setpoint/temperature/update":
			// assert type of payload
			newTemp, ok := a.Payload.(float64)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.TemperatureSetPoint = newTemp
			return oldState, nil
		case "setpoint/pressure/update":
			// assert type of payload
			newPres, ok := a.Payload.(float64)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.PressureSetPoint = newPres
			return oldState, nil
		default:
			return nil, state.ErrInvalidAction
		} 
	}
	rtdInitialState = data.InitialRtdState{}
	rtdReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialRtdState)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(data.InitialRtdState)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState.RealTimeData
			oldState.AverageTemperature = newState.AverageTemperature
			return oldState, nil
		case "rga/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState
			return oldState, nil
		case "telemetry/polling_interval/update":
			// assert type of payload
			newPol, ok := a.Payload.(int64)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.TelPollingInterval = newPol
			return oldState, nil
		default:
			return nil, state.ErrInvalidAction
		} 
	}
)

// initTelemetryService initializes a new telemetry service
func initTelemetryService(t *testing.T) (*TelemetryService, func()) {
	tmp_dir, err := ioutil.TempDir("", "telemetry_test-")
	if err != nil {
		t.Fatalf("Could not create a temporary directory: %v", err)
	}
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	stateStore := state.CreateStore(rtdInitialState, rtdReducer)
	ctrlStore := state.CreateStore(ctrlInitialState, ctrlReducer)
	return NewTelemetryService(&log, tmp_dir, stateStore, ctrlStore, drivers.BlankConnection{}), func(){os.RemoveAll(tmp_dir)}
}

// TestTelemetryService tests if we can initialize the TelemetryService struct and properly connect to the telemetry DAQ
func TestTelemetryService(t *testing.T) {
	telemetryService, cleanUp := initTelemetryService(t)
	defer cleanUp()
	TelemetryServiceStart(t, telemetryService)
	TelemetryRecording(t, telemetryService)
}

// TelemetryServiceStart tests wether we can successfully start the Telemetry service
func TelemetryServiceStart(t *testing.T, s *TelemetryService) {
	err := s.Start()
	if err != nil {
		t.Errorf("Could not start telemetry service: %v", err)
	}
}

// Recording tests whether a recording can be successfully started and stopped
func TelemetryRecording(t *testing.T, s *TelemetryService) {
	err := s.startRecording(DefaultPollingInterval)
	if err != nil {
		t.Errorf("Could not start recording: %v", err)
	}
	err = s.stopRecording()
	if err != nil {
		t.Errorf("Could not stop recording: %v", err)
	}
	time.Sleep(5*time.Second)
	err = s.Stop() // Only stop after since closing closes the channels
	if err != nil {
		t.Errorf("Could not stop telemetry service: %v", err)
	}
}