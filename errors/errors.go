/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package errors

type Error string

const (
	ErrInvalidType = Error("invalid type")
)

// Error implements the golang standard library error interface.
// This allows us to declare errors as constants
func (e Error) Error() string {
	return string(e)
}