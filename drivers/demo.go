// +build demo

package drivers

// RGAConnection
type RGAConnection struct {
	BlankConnection
}

type DAQConnection struct {
	BlankConnection
}

// ConnectToRGA returns a blank connection. Used for demo version of FMT
func ConnectToRGA() (BlankConnection, error) {
	return BlankConnection{}, nil
}

// ConnectToDAQ returns a blank connection. Used for demo version of FMT
func ConnectToDAQ() (BlankConnection, error) {
	return BlankConnection{}, nil
}