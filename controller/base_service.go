/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"sync"

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

type controllerState uint8

const (
	waitingToStart controllerState = iota
	inUse
)

type BaseControllerService struct {
	Running				int32 // used atomically
	ctrlState			controllerState
	name 				string
	rtdStateStore		*state.Store
	ctrlStateStore	*state.Store
	Logger				*zerolog.Logger
	sync.RWMutex
}