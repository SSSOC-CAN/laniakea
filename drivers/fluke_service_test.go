package drivers

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/utils"
)

// TestGetAllTags tests if we can connect to Fluke DAQ OPC DA server and get a list of all tags
func TestGetAllTags(t *testing.T) {
	_, err := GetAllTags()
	if err != nil {
		t.Errorf("Coukd not grab all tags from Fluke DAQ OPC DA server: %v", err)
	}
	//TODO:SSSOCPaulCote - Create a hardcoded list of all tags we expect and make sure we are getting them
}

// TestNewFlukeService tests if we can initialize the FlukeService struct and properly connect to the Fluke DAQ OPC DA server
func TestNewFlukeService(t *testing.T) {
	tmp_dir, err := ioutil.TempDir(utils.AppDataDir("fmtd", false), "fluke_test")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	stateStore := state.CreateStore(FlukeInitialState, FlukeReducer)
	flukeService, err := NewFlukeService(&log, tmp_dir, stateStore)
	if err != nil {
		t.Errorf("Could not instantiate FlukeService struct: %v", err)
	}
	FlukeServiceStart(t, flukeService)
	FlukeRecording(t, flukeService)
}

// FlukeServiceStart tests wether we can successfully start the fluke service
func FlukeServiceStart(t *testing.T, s *FlukeService) {
	err := s.Start()
	if err != nil {
		t.Errorf("Could not start Fluke Service: %v", err)
	}
}

// Recording tests whether a recording can be successfully started and stopped
func FlukeRecording(t *testing.T, s *FlukeService) {
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
		t.Errorf("Could not stop Fluke Service: %v", err)
	}
}