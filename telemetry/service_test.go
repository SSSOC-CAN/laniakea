// +build !demo

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
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOCPaulCote/gux"
)

// initTelemetryService initializes a new telemetry service
func initTelemetryService(t *testing.T) (*TelemetryService, func()) {
	tmp_dir, err := ioutil.TempDir("", "telemetry_test-")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	stateStore := gux.CreateStore(TelemetryInitialState, TelemetryReducer)
	c, err := drivers.ConnectToDAQ()
	if err != nil {
		t.Fatalf("Could not connect to telemetry DAQ: %v", err)
	}
	daqConn, ok := c.(*drivers.DAQConnection)
	if !ok {
		t.Fatalf("Could not connect to telemetry DAQ: %v", errors.ErrInvalidType)
	}
	return NewTelemetryService(&log, stateStore, _, daqConn, "", ""), func(){
		daqConn.Close()
		os.RemoveAll(tmp_dir)
	}
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
	err := s.startRecording(drivers.TelemetryDefaultPollingInterval, "", "")
	if err == nil {
		t.Errorf("Expected an error and none occured")
	}
	err = s.stopRecording()
	if err == nil {
		t.Errorf("Expected an error and none occured")
	}
	time.Sleep(5*time.Second)
	err = s.Stop() // Only stop after since closing closes the channels
	if err != nil {
		t.Errorf("Could not stop telemetry service: %v", err)
	}
}