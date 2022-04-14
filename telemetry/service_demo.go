// +build demo

package telemetry

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/rs/zerolog"
)

var (
	DefaultPollingInterval int64 = 5
	minPollingInterval     int64 = 5
	orgName 			   string = "sssoc"
	bucketName			   string = "ottawa"
)

// TelemetryService is a struct for holding all relevant attributes to interfacing with the DAQ
type TelemetryService struct {
	BaseTelemetryService
	idb influx.Client
}

// A compile time check to make sure that TelemetryService fully implements the data.Service interface
var _ data.Service = (*TelemetryService) (nil)


// NewTelemetryService creates a new Telemetry Service object which will use the drivers for the DAQ software
func NewTelemetryService(
	logger *zerolog.Logger,
	outputDir string,
	rtdStore *state.Store,
	ctrlStore *state.Store,
	_ drivers.DriverConnection,
) *TelemetryService {
	var (
		wgL sync.WaitGroup
		wgR sync.WaitGroup
	)
	client := influx.NewClientWithOptions("http://192.168.0.87:8086", "nQN6nFdpaXIG8EkgXq-BTP8E5uApWb0WawQaUUVZg2oiWIWBaP8C4-1bNZhjataYDmif_iLD4_rD07UhdpmEtw==", influx.DefaultOptions().SetBatchSize(12))
	return &TelemetryService{
		BaseTelemetryService: BaseTelemetryService{
			rtdStateStore:    rtdStore,
			ctrlStateStore:	  ctrlStore,
			Logger:       	  logger,
			QuitChan:     	  make(chan struct{}),
			CancelChan:   	  make(chan struct{}),
			outputDir:   	  outputDir,
			name: 	      	  data.TelemetryName,
			wgListen:		  wgL,
			wgRecord:		  wgR,
		},
		idb: client,
	}
}

// Start starts the service. Returns an error if any issues occur
func (s *TelemetryService) Start() error {
	s.Logger.Info().Msg("Starting telemetry service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start telemetry service: service already started.")
	}
	s.Logger.Info().Msg("Connection to DAQ is not currently active. Please recompile fmtd as follows `$ go install -tags \"fluke\"`")
	s.wgListen.Add(1)
	go s.ListenForRTDSignal()
	s.Logger.Info().Msg("Telemetry service started.")
	return nil
}

// Stop stops the service. Returns an error if any issues occur
func (s *TelemetryService) Stop() error {
	s.Logger.Info().Msg("Stopping telemetry service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop telemetry: service already stopped.")
	}
	if atomic.LoadInt32(&s.Recording) == 1 {
		err := s.stopRecording()
		if err != nil {
			return fmt.Errorf("Could not stop telemetry service: %v", err)
		}
	}
	close(s.CancelChan)
	s.wgRecord.Wait()
	close(s.QuitChan)
	s.wgListen.Wait()
	s.Logger.Info().Msg("Telemetry service stopped.")
	return nil
}

// Name satisfies the fmtd.Service interface
func (s *TelemetryService) Name() string {
	return s.name
}

// StartRecording starts the recording process by creating a csv file and inserting the header row into the file and returns a quit channel and error message
func (s *TelemetryService) startRecording(pol_int int64) error {
	if pol_int < minPollingInterval && pol_int != 0 {
		return fmt.Errorf("Inputted polling interval smaller than minimum value: %v", minPollingInterval)
	} else if pol_int == 0 { //No polling interval provided
		pol_int = DefaultPollingInterval
	}
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
		return ErrAlreadyRecording
	}
	writeAPI := s.idb.WriteAPI(orgName, bucketName)
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// Write polling interval to state
	err := s.rtdStateStore.Dispatch(
		state.Action{
			Type: 	 "telemetry/polling_interval/update",
			Payload: pol_int,
		},
	)
	if err != nil {
		return fmt.Errorf("Could not update state: %v", err)
	}
	// the actual data
	s.Logger.Info().Msg("Starting data recording...")
	s.wgRecord.Add(1)
	go func() {
		defer s.wgRecord.Done()
		for {
			select {
			case <-ticker.C:
				err = s.record(writeAPI)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not write to influxdb: %v", err))
				}
			case <-s.CancelChan:
				ticker.Stop()
				writeAPI.Flush()
				s.idb.Close()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			case <-s.QuitChan:
				ticker.Stop()
				writeAPI.Flush()
				s.idb.Close()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			}
		}
	}()
	return nil
}

// record records the live data from DAQ and inserts it into a csv file and passes it to the RTD service
func (s *TelemetryService) record(writer api.WriteAPI) error {
	current_time := time.Now()
	current_time_str := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", current_time.Year(), current_time.Month(), current_time.Day(), current_time.Hour(), current_time.Minute(), current_time.Second())
	dataString := []string{current_time_str}
	dataField := make(map[int64]*fmtrpc.DataField)
	var (
		cumSum float64
		cnt	   int64
	)
	// get current pressure set point
	currentState := s.ctrlStateStore.GetState()
	cState, ok := currentState.(data.InitialCtrlState)
	if !ok {
		return state.ErrInvalidStateType
	}
	for i := 0; i < 96; i++ {
		var (
			v float64
			n string
		)
		if i != int(drivers.TelemetryPressureChannel) {
			v = (rand.Float64()*0.1)+cState.TemperatureSetPoint
			n = fmt.Sprintf("Temperature %v", i+1)
			cumSum += v
			cnt += 1
			p := influx.NewPoint(
				"temperature",
				map[string]string{
					"id":       fmt.Sprintf("%v", i+1),
				},
				map[string]interface{}{
					"temperature": v,
				},
				current_time,
			)
			// write asynchronously
			writer.WritePoint(p)
		} else {
			v = (rand.Float64()*0.1)+cState.PressureSetPoint
			n = "Pressure"
			p := influx.NewPoint(
				"pressure",
				map[string]string{
					"id":       fmt.Sprintf("%v", i+1),
				},
				map[string]interface{}{
					"pressure": v,
				},
				current_time,
			)
			// write asynchronously
			writer.WritePoint(p)
		}
		
		dataField[int64(i)]= &fmtrpc.DataField{
			Name: n,
			Value: v,
		}
		dataString = append(dataString, fmt.Sprintf("%g", v))
	}
	
	err := s.rtdStateStore.Dispatch(
		state.Action{
			Type: 	 "telemetry/update",
			Payload: data.InitialRtdState{
				RealTimeData: fmtrpc.RealTimeData{
					Source: s.name,
					IsScanning: false,
					Timestamp: current_time.UnixMilli(),
					Data: dataField,
				},
				AverageTemperature: cumSum/float64(cnt),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("Could not update state: %v", err)
	}
	return nil
}

// stopRecording sends an empty struct down the CancelChan to innitiate the stop recording process
func (s *TelemetryService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return ErrAlreadyStoppedRecording
	}
	s.CancelChan<-struct{}{}
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *TelemetryService) ListenForRTDSignal() {
	defer s.wgListen.Done()
	for {
		select {
		case msg := <-s.StateChangeChan:
			switch msg.Type {
			case data.RECORDING:
				if msg.State {
					n, err := strconv.ParseInt(msg.Msg, 10, 64)
					if err != nil {
						n = DefaultPollingInterval
					}
					err = s.startRecording(n)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not start recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
					} else {
						s.Logger.Info().Msg("Started recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: nil, Msg: s.filepath}
					}
				} else {
					s.Logger.Info().Msg("Stopping data recording...")
					err := s.stopRecording()
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: fmt.Errorf("Could not stop recording: %v", err)}
					} else {
						s.Logger.Info().Msg("Stopped recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: nil}
					}
				}
			}
		case <-s.QuitChan:
			return
		}
	}
}

// RegisterWithRTDService adds the RTD Service channels to the Telemetry Service Struct and incrememnts the number of registered data providers on the RTD
func (s *TelemetryService) RegisterWithRTDService(rtd *data.RTDService) {
	rtd.RegisterDataProvider(s.name)
	s.StateChangeChan = rtd.StateChangeChans[s.name]
}