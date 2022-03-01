package drivers

import (
	"io/ioutil"
	"os"
	"testing"
	"github.com/rs/zerolog"
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
	stateStore := state.CreateStore(RGAInitialState, RGAReducer)
	rga, err := NewRGAService(&log, tmp_dir, stateStore)
	if err != nil {
		t.Errorf("Could not instantiate RGA service: %v", err)
	}
	RGAServiceStart(t, rga)
	RGATestConnection(t, rga)
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

// RGATestConnection tests a few info commands on the RGA
func RGATestConnection(t *testing.T, s *RGAService) {
	err := s.connection.InitMsg()
	if err != nil {
		t.Errorf("Unable to communicate with RGA: %v", err)
	}
	resp, err := s.connection.SensorState()
	if err != nil {
		t.Errorf("Unable to communicate with RGA: %v", err)
	}
	t.Logf("SensoreState msg: %v", resp)
	resp, err = s.connection.FilamentInfo()
	if err != nil {
		t.Errorf("Unable to communicate with RGA: %v", err)
	}
	t.Logf("FilamentInfo msg: %v", resp)
}