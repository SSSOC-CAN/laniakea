// +build !demo

package rga

import (
	"io/ioutil"
	"os"
	"testing"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/utils"
)

// TestNewMessage tests if we can connect to MKSRGA server and create a new instance of the RGA Service struct
func TestNewMessage(t *testing.T) {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tmp_dir, err := ioutil.TempDir(utils.AppDataDir("fmtd", false), "rga_test")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	c, err := drivers.ConnectToRGA()
	if err != nil {
		t.Fatalf("Could not connect to RGA: %v", err)
	}
	rgaConn, ok := c.(*drivers.RGAConnection)
	if !ok {
		t.Fatalf("Could not connect to MKS RGA: %v", errors.ErrInvalidType)
	}
	defer rgaConn.Close()
	stateStore := state.CreateStore(RGAInitialState, RGAReducer)
	rgaService := NewRGAService(&log, tmp_dir, stateStore, rgaConn)
	if err != nil {
		t.Errorf("Could not instantiate RGA service: %v", err)
	}
	RGAServiceStart(t, rgaService)
	RGAStartRecord(t, rgaService)
	RGAStopRecord(t, rgaService)
	RGAServiceStop(t, rgaService)
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
	err := s.startRecording(drivers.RGAMinPollingInterval)
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