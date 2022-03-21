// +build !demo

package telemetry

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/utils"
)

// initTelemetryService initializes a new telemetry service
func initTelemetryService(t *testing.T) *TelemetryService {
	tmp_dir, err := ioutil.TempDir(utils.AppDataDir("fmtd", false), "telemetry_test")
	if err != nil {
		t.Fatalf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	stateStore := state.CreateStore(TelemetryInitialState, TelemetryReducer)
	daqConn, err := drivers.ConnectToDAQ()
	if err != nil {
		t.Fatalf("Could not connect to telemetry DAQ: %v", err)
	}
	return NewTelemetryService(&log, tmp_dir, stateStore, drivers, daqConn)
}

// TestTelemetryService tests if we can initialize the TelemetryService struct and properly connect to the telemetry DAQ
func TestTelemetryService(t *testing.T) {
	telemetryService := initTelemetryService(t)
	TelemetryServiceStart(t, TelemetryService)
	TelemetryRecording(t, TelemetryService)
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