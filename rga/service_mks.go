package rga

import (
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/state"
)

type RGAService struct {
	Running				int32 //atomically
	Recording			int32 //atomically
	stateStore			*state.Store
	Logger				*zerolog.Logger
	PressureChan		chan float64 // RGAs can't operate above certain pressures and so must get pressure readings from outside sources to prevent damaging them
	StateChangeChan		chan *data.StateChangeMsg
	QuitChan			chan struct{}
	CancelChan			chan struct{}
	outputDir  			string
	name 				string
	currentPressure		float64
	filepath			string
}