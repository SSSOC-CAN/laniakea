package drivers

// DriverConnection interface defines a generic type with a Close() function
type DriverConnection interface {
	Close()
}

// Compile time check to ensure BlankConnection satisfies the DriverConnection interface
var _ DriverConnection = (*BlankConnection) (nil)
var _ DriverConnection = BlankConnection{}

type BlankConnection struct {}

// Close implements the DriverConnection
func (b BlankConnection) Close() {
	return
}