package controller

import (
	"github.com/SSSOC-CAN/fmtd/state"
)

var (
	ControllerName = "CTRL"
)

type BaseControllerService struct {
	Running				int32 // used atomically
	name 				string
	stateStore			*state.Store
}

// Name satisfies the data.Service interface
func (s *BaseRGAService) Name() string {
	return s.name
}