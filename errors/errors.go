package errors

type Error string

// Error implements the golang standard library error interface.
// This allows us to declare errors as constants
func (e Error) Error() string {
	return string(e)
}