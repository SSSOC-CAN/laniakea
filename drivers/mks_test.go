// +build mks,!demo

package drivers

import (
	"testing"
	
	"github.com/SSSOC-CAN/fmtd/errors"
)

// TestConnectToRGA tests if we can connect to the RGA
func TestConnectToRGA(t *testing.T) {
	c, err := ConnectToRGA()
	if err != nil {
		t.Fatalf("Could not connect to MKS RGA: %v", err)
	}
	c.Close()
}

// TestInitMsg tests if we can send the Init message to the MKS RGA
func TestInitMsg(t *testing.T) {
	c, err := ConnectToRGA()
	if err != nil {
		t.Fatalf("Could not connect to MKS RGA: %v", err)
	}
	defer c.Close()
	rgaConn, ok := c.(*RGAConnection)
	if !ok {
		t.Fatalf("Could not connect to MKS RGA: %v", errors.ErrInvalidType)
	}
	err = rgaConn.InitMsg()
	if err != nil {
		t.Errorf("Could not send Init Msg: %v", err)
	}
}

// TestSensorState tests if we can send the SensorState command to MKS RGA
func TestSensorState(t *testing.T) {
	c, err := ConnectToRGA()
	if err != nil {
		t.Fatalf("Could not connect to MKS RGA: %v", err)
	}
	defer c.Close()
	rgaConn, ok := c.(*RGAConnection)
	if !ok {
		t.Fatalf("Could not connect to MKS RGA: %v", errors.ErrInvalidType)
	}
	resp, err := rgaConn.SensorState()
	if err != nil {
		t.Errorf("Unable to send SensorState command: %v", err)
	}
	t.Logf("SensoreState msg: %v", resp)
}

// TestFilamentInfo tests if we can send the FilamentInfo command to MKS RGA
func TestFilamentInfo(t *testing.T) {
	c, err := ConnectToRGA()
	if err != nil {
		t.Fatalf("Could not connect to MKS RGA: %v", err)
	}
	defer c.Close()
	rgaConn, ok := c.(*RGAConnection)
	if !ok {
		t.Fatalf("Could not connect to MKS RGA: %v", errors.ErrInvalidType)
	}
	resp, err := rgaConn.FilamentInfo()
	if err != nil {
		t.Errorf("Unable to communicate with RGA: %v", err)
	}
	t.Logf("FilamentInfo msg: %v", resp)
}