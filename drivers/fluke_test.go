// +build windows,386,fluke,!demo

package drivers

import (
	"testing"

	"github.com/SSSOC-CAN/fmtd/errors"
)

// TestConnectToDAQ tests if the connection to the DAQ OPC server can be made successfully
func TestConnectToDAQ(t *testing.T) {
	daqConn, err := ConnectToDAQ()
	if err != nil {
		t.Fatalf("Could not connect to Fluke DAQ OPC: %v", err)
	}
	daqConn.Close()
}

// TestGetAllTags tests if we can retrieve tags from the DAQ OPC server
func TestGetAllTags(t *testing.T) {
	_, err := GetAllTags()
	if err != nil {
		t.Fatalf("Could not get tags from Fluke DAQ OPC: %v", err)
	}
}

// TestStartStopScanning tests if we can start and stop the scanning process on the DAQ OPC server
func TestStartStopScanning(t *testing.T) {
	c, err := ConnectToDAQ()
	if err != nil {
		t.Fatalf("Could not connect to Fluke DAQ OPC: %v", err)
	}
	defer c.Close()
	daqConn, ok := c.(*DAQConnection)
	if !ok {
		t.Fatalf("Could not connect to Fluke DAQ OPC: %v", errors.ErrInvalidType)
	}
	err = daqConn.StartScanning()
	if err != nil {
		t.Errorf("Could not start scanning on Fluke DAQ OPC: %v", err)
	}
	err = daqConn.StopScanning()
	if err != nil {
		t.Errorf("Could not stop scanning on Fluke DAQ OPC: %v", err)
	}
}