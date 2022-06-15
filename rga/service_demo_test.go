// +build demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package rga

import (
	"io/ioutil"
	"os"
	"testing"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOCPaulCote/gux"
)

var (
	ctrlInitialState = data.InitialCtrlState{
		PressureSetPoint: 760.0,
	}
	ctrlReducer gux.Reducer = func(s interface{}, a gux.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialCtrlState)
		if !ok {
			return nil, errors.ErrInvalidType
		}
		// switch case action
		switch a.Type {
		case "setpoint/temperature/update":
			// assert type of payload
			newTemp, ok := a.Payload.(float64)
			if !ok {
				return nil, errors.ErrInvalidType
			}
			oldState.TemperatureSetPoint = newTemp
			return oldState, nil
		case "setpoint/pressure/update":
			// assert type of payload
			newPres, ok := a.Payload.(float64)
			if !ok {
				return nil, errors.ErrInvalidType
			}
			oldState.PressureSetPoint = newPres
			return oldState, nil
		default:
			return nil, errors.ErrInvalidAction
		} 
	}
)

// TestNewMessage tests if we can connect to MKSRGA server and create a new instance of the RGA Service struct
func TestNewMessage(t *testing.T) {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tmp_dir, err := ioutil.TempDir("", "rga_test-")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	stateStore := gux.CreateStore(RGAInitialState, RGAReducer)
	ctrlStore := gux.CreateStore(ctrlInitialState, ctrlReducer)
	rga := NewRGAService(&log, tmp_dir, stateStore, ctrlStore, drivers.BlankConnectionErr{})
	if err != nil {
		t.Errorf("Could not instantiate RGA service: %v", err)
	}
	RGAServiceStart(t, rga)
	RGAStartRecord(t, rga)
	RGAStopRecord(t, rga)
	RGAServiceStop(t, rga)
}

// RGAServiceStart tests whether we can successfully start the rga service
func RGAServiceStart(t *testing.T, s *RGAService) {
	err := s.Start()
	if err != nil {
		t.Errorf("Could not start RGA Service: %v", err)
	}
	
}

// RGAServiceStop tests whether we can successfully stop the rga service
func RGAServiceStop(t *testing.T, s *RGAService) {
	err := s.Stop()
	if err != nil {
		t.Errorf("Could not stop RGA Service: %v", err)
	}
}

// RGAStartRecord tests the startRecording method and expects an error
func RGAStartRecord(t *testing.T, s *RGAService) {
	err := s.startRecording(minRgaPollingInterval)
	if err == nil {
		t.Fatalf("Expected an error and none occured")
	}
}

// RGAStopRecord tests the stopRecording method and expects an error
func RGAStopRecord(t *testing.T, s *RGAService) {
	err := s.stopRecording()
	if err == nil {
		t.Fatalf("Expected an error and none occured")
	}
}