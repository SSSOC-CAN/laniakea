// +build demo

package drivers

var (
	TelemetryPressureChannel        int64 = 81
)

// RGAConnection
type RGAConnection struct {
	BlankConnectionErr
}

type DAQConnection struct {
	BlankConnection
}

type ControllerConnection struct {
	BlankConnection
}

// ConnectToRGA returns a blank connection. Used for demo version of FMT
func ConnectToRGA() (DriverConnectionErr, error) {
	return &RGAConnection{BlankConnectionErr{}}, nil
}

// ConnectToDAQ returns a blank connection. Used for demo version of FMT
func ConnectToDAQ() (DriverConnection, error) {
	return &DAQConnection{BlankConnection{}}, nil
}