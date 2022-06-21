// +build !demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOCPaulCote/gux"
)

var (
	makeInitialDataMap = func() map[int64]*fmtrpc.DataField {
		m := make(map[int64]*fmtrpc.DataField)
		for i := 0; i < 96; i++ {
			m[int64(i)] = &fmtrpc.DataField{
				Name: fmt.Sprintf("Value #%v", i+1),
				Value: (rand.Float64()*0.1)+25,
			}
		}
		return m
	}
	rtdInitialState = data.InitialRtdState{
		AverageTemperature: float64(25.0),
		TelPollingInterval: int64(5),
		RealTimeData: fmtrpc.RealTimeData{
			Source: "TEL",
			IsScanning: false,
			Timestamp: time.Now().UnixMilli(),
			Data: makeInitialDataMap(),
		},
	}
	rtdReducer gux.Reducer = func(s interface{}, a gux.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialRtdState)
		if !ok {
			return nil, errors.ErrInvalidType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(data.InitialRtdState)
			if !ok {
				return nil, errors.ErrInvalidType
			}
			oldState.RealTimeData = newState.RealTimeData
			oldState.AverageTemperature = newState.AverageTemperature
			return oldState, nil
		case "rga/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, errors.ErrInvalidType
			}
			oldState.RealTimeData = newState
			return oldState, nil
		case "telemetry/polling_interval/update":
			// assert type of payload
			newPol, ok := a.Payload.(int64)
			if !ok {
				return nil, errors.ErrInvalidType
			}
			oldState.TelPollingInterval = newPol
			return oldState, nil
		default:
			return nil, errors.ErrInvalidAction
		} 
	}
)

// initController initializes a controller service
func initController(t *testing.T) (*ControllerService, func(string) error, string) {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tmp_dir, err := ioutil.TempDir("", "controller_test-")
	if err != nil {
		t.Fatalf("Could not create a temporary directory: %v", err)
	}
	stateStore := gux.CreateStore(rtdInitialState, rtdReducer)
	ctrlStore := gux.CreateStore(InitialState, ControllerReducer)
	ctrlConn, _ := drivers.ConnectToController()
	return NewControllerService(
		&log,
		stateStore,
		ctrlStore,
		ctrlConn,
	), os.RemoveAll, tmp_dir
}

// TestStartStopControllerService tests the start/stop functions of the ControllerService
func TestStartStopControllerService(t *testing.T) {
	controllerService, cleanUp , tmpDir := initController(t)
	defer cleanUp(tmpDir)
	t.Run("Start Controller Service", func(t *testing.T) {
		err := controllerService.Start()
		if err != nil {
			t.Errorf("Could not start controller service: %v", err)
		}
	})
	t.Run("Start Controller Service Invalid", func(t *testing.T) {
		err := controllerService.Start()
		if err != errors.ErrServiceAlreadyStarted {
			t.Errorf("Unexpected error when starting controller service: %v", err)
		}
	})
	t.Run("Stop Controller Service", func(t *testing.T) {
		err := controllerService.Stop()
		if err != nil {
			t.Errorf("Could not stop controller service: %v", err)
		}
	})
	t.Run("Stop Controller Service Invalid", func(t *testing.T) {
		err := controllerService.Stop()
		if err != errors.ErrServiceAlreadyStopped {
			t.Errorf("Unexpected error when starting controller service: %v", err)
		}
	})
}