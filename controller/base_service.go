package controller

import (
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/rs/zerolog"
)

const (
	ErrServiceAlreadyStarted = errors.Error("service already started")
	ErrServiceAlreadyStopped = errors.Error("service already stopped")
)

var (
	ControllerName = "CTRL"
)

type BaseControllerService struct {
	Running				int32 // used atomically
	name 				string
	rtdStateStore		*state.Store
	ctrlStateStore	*state.Store
	Logger				*zerolog.Logger
}