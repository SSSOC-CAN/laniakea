package rga

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/state"
)

var (
	ErrAlreadyRecording			 = fmt.Errorf("Could not start recording. Data recording already started")
	ErrAlreadyStoppedRecording   = fmt.Errorf("Could not stop data recording. Data recording already stopped.")
)

type BaseRGAService struct {
	Running				int32 //atomically
	Recording			int32 //atomically
	rtdStateStore		*state.Store
	ctrlStateStore		*state.Store
	Logger				*zerolog.Logger
	PressureChan		chan float64 // RGAs can't operate above certain pressures and so must get pressure readings from outside sources to prevent damaging them
	StateChangeChan		chan *data.StateChangeMsg
	QuitChan			chan struct{}
	CancelChan			chan struct{}
	outputDir  			string
	name 				string
	currentPressure		float64
	filepath			string
	wgListen			sync.WaitGroup
	wgRecord			sync.WaitGroup
}

// Name satisfies the data.Service interface
func (s *BaseRGAService) Name() string {
	return s.name
}