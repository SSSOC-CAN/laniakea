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
	ErrInvalidType             = bg.Error("invalid type")
	ErrServiceAlreadyStarted   = bg.Error("service already started")
	ErrServiceAlreadyStopped   = bg.Error("service already stopped")
	ErrNoError                 = bg.Error("expected an error and got none")
	ErrAlreadyRecording        = bg.Error("already recording")
	ErrAlreadyStoppedRecording = bg.Error("already stopped recording")
	ErrPollingIntervalTooSmall = bg.Error("inputted polling interval smaller than minimum required value")
	ErrMacSvcNil               = bg.Error("macaroon service uninitialized")
	ErrAlreadySubscribed       = bg.Error("subscriber with given name already subscribed")
	ErrInvalidPluginName       = bg.Error("invalid plugin name")
	ErrInvalidRequestType      = bg.Error("request is not a valid type")
	ErrDuplicateMacConstraints = bg.Error("duplicate macaroon constraints for given path")
	ErrUnknownPermission       = bg.Error("unknown permission for given method")
)
