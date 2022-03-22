package drivers

// DriverConnection interface defines a generic type with a Close() function
type DriverConnection interface {
	Close()
}

type DriverConnectionErr interface {
	Close() error
}

// Compile time check to ensure BlankConnection satisfies the DriverConnection interface
var _ DriverConnection = (*BlankConnection) (nil)
var _ DriverConnection = BlankConnection{}

type BlankConnection struct {}

// Close implements the DriverConnection
func (b BlankConnection) Close() {
	return
}

var _ DriverConnectionErr = (*BlankConnectionErr) (nil)
var _ DriverConnectionErr = BlankConnectionErr{}

type BlankConnectionErr struct {}

// Close implements the DriverConnectionErr interface
func (b BlankConnectionErr) Close() error {
	return nil
}