// +build demo

package rga

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"reflect"
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
	minRgaPollingInterval int64 = 15 // After manual testing, value will be 15 seconds
	minimumPressure float64 = 0.00005
	bucketName string = "test"
)

type RGAService struct {
	BaseRGAService
}

// A compile time check to make sure that RGAService fully implements the data.Service interface
var _ data.Service = (*RGAService) (nil)

// NewRGAService creates an instance of the RGAService struct. It also establishes a connection to the RGA device
func NewRGAService(
	logger *zerolog.Logger, 
	rtdStore *state.Store, 
	ctrlStore *state.Store,
	_ drivers.DriverConnectionErr,
	influxUrl string,
	influxToken string,
	influxOrg string,
) *RGAService {
	var (
		wgL sync.WaitGroup
		wgR sync.WaitGroup
	)
	client := influx.NewClientWithOptions(influxUrl, influxToken, influx.DefaultOptions().SetTLSConfig(&tls.Config{InsecureSkipVerify : true}))
	return &RGAService{
		BaseRGAService{
			rtdStateStore: 	  rtdStore,
			ctrlStateStore:   ctrlStore,
			Logger: 		  logger,
			QuitChan:		  make(chan struct{}),
			CancelChan: 	  make(chan struct{}),
			name: 			  data.RgaName,
			wgListen: 		  wgL,
			wgRecord: 		  wgR,
			influxOrgName:    influxOrg,
			idb:      		  client,
		},
	}
}

// Start starts the RGA service. It does NOT start the data recording process
func (s *RGAService) Start() error {
	s.Logger.Info().Msg("Starting RGA Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start RGA service. Service already started.")
	}
	s.Logger.Info().Msg("Connection to MKS RGA is not currently active. Please recompile fmtd as follows `$ go install -tags \"mks\"`")
	s.wgListen.Add(1)
	go s.ListenForRTDSignal()
	s.Logger.Info().Msg("RGA Service successfully started.")
	return nil
}

// Stop stops the RGA service.
func (s *RGAService) Stop() error {
	s.Logger.Info().Msg("Stopping RGA Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop RGA service. Service already stopped.")
	}
	if atomic.LoadInt32(&s.Recording) == 1 {
		err := s.stopRecording()
		if err != nil {
			return fmt.Errorf("Could not stop RGA service: %v", err)
		}
	}
	close(s.CancelChan)
	s.wgRecord.Wait()
	close(s.QuitChan)
	s.wgListen.Wait()
	s.Logger.Info().Msg("RGA Service successfully stopped.")
	return nil
}

// startRecording starts data recording from the RGA device
func (s *RGAService) startRecording(pol_int int64) error {
	if s.currentPressure == 0 || s.currentPressure > minimumPressure {
		return fmt.Errorf("Current chamber pressure is too high. Current: %.6f Torr\tMinimum: %.5f Torr", s.currentPressure, minimumPressure)
	}
	if pol_int < minRgaPollingInterval && pol_int != 0 {
		return fmt.Errorf("Inputted polling interval smaller than minimum value: %v", minRgaPollingInterval)
	} else if pol_int == 0 { //No polling interval provided
		pol_int = minRgaPollingInterval
	}
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
		return ErrAlreadyRecording
	}
	writeAPI := s.idb.WriteAPI(s.influxOrgName, bucketName)
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// Write polling interval to state
	err := s.rtdStateStore.Dispatch(
		state.Action{
			Type: 	 "rga/polling_interval/update",
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

//record writes data from the RGA to a csv file and will pass it along it's Output channel
func (s *RGAService) record(writer api.WriteAPI) error {
	current_time := time.Now()
	dataField := make(map[int64]*fmtrpc.DataField)
	// figure out noise amplitude and offset
	// get current pressure set point
	currentState := s.ctrlStateStore.GetState()
	cState, ok := currentState.(data.InitialCtrlState)
	if !ok {
		return state.ErrInvalidStateType
	}
	for i := 0; i < 200; i++ {
		v := (rand.Float64()*0.001)+cState.PressureSetPoint
		dataField[int64(i)]= &fmtrpc.DataField{
			Name: fmt.Sprintf("Mass %v", i+1),
			Value: v,
		}
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
	err := s.rtdStateStore.Dispatch(
		state.Action{
			Type: 	 "rga/update",
			Payload: fmtrpc.RealTimeData{
				Source: s.name,
				IsScanning: false,
				Timestamp: current_time.UnixMilli(),
				Data: dataField,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("Could not update state: %v", err)
	}
	return nil
}

// stopRecording stops the data recording process
func (s *RGAService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return ErrAlreadyStoppedRecording
	}
	s.CancelChan <- struct{}{}
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *RGAService) ListenForRTDSignal() {
	defer s.wgListen.Done()
	signalChan, unsub := s.rtdStateStore.Subscribe(s.name)
	cleanUp := func() {
		unsub(s.rtdStateStore, s.name)
	}
	defer cleanUp()
	for {
		select {
		case msg := <- s.StateChangeChan:
			switch msg.Type {
			case data.RECORDING:
				if msg.State {
					err := s.startRecording(0)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not start recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
					} else {
						s.Logger.Info().Msg("Started recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: nil}
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
		case <- signalChan:
			if atomic.LoadInt32(&s.Recording) != 1 {
				continue
			}
			currentState := s.rtdStateStore.GetState()
			cState, ok := currentState.(data.InitialRtdState)
			if !ok {
				s.Logger.Error().Msg(fmt.Sprintf("Invalid type %v expected %v\nStopping recording...", reflect.TypeOf(currentState), reflect.TypeOf(data.InitialRtdState{})))
				err := s.stopRecording()
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
				}
			}
			s.currentPressure = cState.RealTimeData.Data[drivers.TelemetryPressureChannel].Value
			if s.currentPressure >= 0.00005 && atomic.LoadInt32(&s.Recording) == 1 {
				err := s.stopRecording()
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
				}
			}
		case <-s.QuitChan:
			return
		}
	}
}

// RegisterWithRTDService adds the RTD Service channels to the RGA Service Struct and incrememnts the number of registered data providers on the RTD
func (s *RGAService) RegisterWithRTDService(rtd *data.RTDService) {
	rtd.RegisterDataProvider(s.name)
	s.StateChangeChan = rtd.StateChangeChans[s.name]
}