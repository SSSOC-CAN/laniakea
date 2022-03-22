// +build demo

package rga

import (
	"io/ioutil"
	"os"
	"testing"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/state"
)

// TestNewMessage tests if we can connect to MKSRGA server and create a new instance of the RGA Service struct
func TestNewMessage(t *testing.T) {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tmp_dir, err := ioutil.TempDir("", "rga_test-")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	stateStore := state.CreateStore(RGAInitialState, RGAReducer)
	rga := NewRGAService(&log, tmp_dir, stateStore, drivers.BlankConnection{})
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