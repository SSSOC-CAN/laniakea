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
	ErrNegativeRate = bg.Error("rates cannot be negative")
	ErrTelNotRecoring = bg.Error("telemetry service not yet recording")
	ErrCtrllerInUse = bg.Error("controller service currently in use")
	ErrNegativePressure = bg.Error("pressures cannot be negative")
	ErrServiceAlreadyStarted = bg.Error("service already started")
	ErrServiceAlreadyStopped = bg.Error("service already stopped")
	ErrEmptyMac = bg.Error("macaroon data is empty")
	ErrInvalidMacEncrypt = bg.Error("invalid encrypted macaroon. Format expected: 'snacl:<key_base64>:<encrypted_macaroon_base64>'")
	ErrFloatLargerThanOne = bg.Error("float is larger than or equal to 1")
	ErrNoError = bg.Error("expected an error and got none")
	ErrMacNotExpired = bg.Error("macaroon didn't expire")
)