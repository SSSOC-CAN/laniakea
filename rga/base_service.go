/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package rga

import (
	"sync"

	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOCPaulCote/gux"
)

var (
	influxRGABucketName = "rga"
)

type BaseRGAService struct {
	Running				int32 //atomically
	Recording			int32 //atomically
	rtdStateStore		*gux.Store
	ctrlStateStore		*gux.Store
	Logger				*zerolog.Logger
	PressureChan		chan float64 // RGAs can't operate above certain pressures and so must get pressure readings from outside sources to prevent damaging them
	StateChangeChan		chan *data.StateChangeMsg
	QuitChan			chan struct{}
	CancelChan			chan struct{}
	name 				string
	currentPressure		float64
	wgListen			sync.WaitGroup
	wgRecord			sync.WaitGroup
	idb 				influx.Client
}

// Name satisfies the data.Service interface
func (s *BaseRGAService) Name() string {
	return s.name
}