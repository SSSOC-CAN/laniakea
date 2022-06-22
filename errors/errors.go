/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/11
*/

package errors

import (
	bg "github.com/SSSOCPaulCote/blunderguard"
)

const (
	ErrInvalidType = bg.Error("invalid type")
	ErrServiceAlreadyStarted = bg.Error("service already started")
	ErrServiceAlreadyStopped = bg.Error("service already stopped")
	ErrNoError = bg.Error("expected an error and got none")
	ErrAlreadyRecording = bg.Error("already recording")
	ErrAlreadyStoppedRecording = bg.Error("already stopped recording")
)