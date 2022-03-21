package drivers

// DriverConnection interface defines a generic type with a Close() function
type DriverConnection interface {
	Close()
}