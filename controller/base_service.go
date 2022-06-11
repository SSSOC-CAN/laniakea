/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"sync"

	"github.com/SSSOCPaulCote/gux"
	"github.com/rs/zerolog"
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
	rtdStateStore		*gux.Store
	ctrlStateStore		*gux.Store
	Logger				*zerolog.Logger
	sync.RWMutex
}