/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package telemetry

import (
	"fmt"
	"sync"
	
	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOCPaulCote/gux"
)

var (
	ErrAlreadyRecording			 = fmt.Errorf("Could not start recording. Data recording already started")
	ErrAlreadyStoppedRecording   = fmt.Errorf("Could not stop data recording. Data recording already stopped.")
)

type BaseTelemetryService struct {
	Running				int32 //atomically
	Recording			int32 //atomically
	rtdStateStore		*gux.Store
	ctrlStateStore		*gux.Store
	Logger				*zerolog.Logger
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
func (s *BaseTelemetryService) Name() string {
	return s.name
}