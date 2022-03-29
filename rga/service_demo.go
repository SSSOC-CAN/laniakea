// +build demo

package rga

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SSSOC-CAN/fmtd/controller"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/rs/zerolog"
)

var (
	minRgaPollingInterval int64 = 15 // After manual testing, value will be 15 seconds
	minimumPressure float64 = 0.00005
)

type RGAService struct {
	BaseRGAService
}

// A compile time check to make sure that RGAService fully implements the data.Service interface
var _ data.Service = (*RGAService) (nil)

// NewRGAService creates an instance of the RGAService struct. It also establishes a connection to the RGA device
func NewRGAService(logger *zerolog.Logger, outputDir string, store *state.Store, _ drivers.DriverConnectionErr) *RGAService {
	var (
		wgL sync.WaitGroup
		wgR sync.WaitGroup
	)
	return &RGAService{
		BaseRGAService{
			stateStore: store,
			Logger: logger,
			QuitChan: make(chan struct{}),
			CancelChan: make(chan struct{}),
			outputDir: outputDir,
			name: data.RgaName,
			wgListen: wgL,
			wgRecord: wgR,
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
	// Create or Open csv
	current_time := time.Now()
	file_name := fmt.Sprintf("%s/%d-%02d-%02d-rga.csv", s.outputDir, current_time.Year(), current_time.Month(), current_time.Day())
	file_name = utils.UniqueFileName(file_name)
	file, err := os.Create(file_name)
	if err != nil {
		return fmt.Errorf("Could not create file %v: %v", file, err)
	}
	s.filepath = file_name
	writer := csv.NewWriter(file)
	// Now we begin recording
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// the actual data
	s.Logger.Info().Msg("Starting data recording...")
	s.wgRecord.Add(1)
	go func() {
		defer s.wgRecord.Done()
		ticks := 0
		for {
			select {
			case <-ticker.C:
				err = s.record(writer, ticks)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not write to %s: %v", file_name, err))
				}
				ticks++
			case <-s.CancelChan:
				ticker.Stop()
				file.Close()
				writer.Flush()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			case <-s.QuitChan:
				ticker.Stop()
				file.Close()
				writer.Flush()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			}
		}
	}()
	return nil
}

//record writes data from the RGA to a csv file and will pass it along it's Output channel
func (s *RGAService) record(writer *csv.Writer, ticks int) error {
	if ticks == 0 {
		// Header data for csv file
		headerData := []string{"Timestamp"}
		for i := 0; i < 200; i++ {
			headerData = append(headerData, fmt.Sprintf("%v AMU", i+1))
		}
		err := writer.Write(headerData)
		if err != nil {
			return err
		}
		return nil
	}
	current_time := time.Now()
	current_time_str := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", current_time.Year(), current_time.Month(), current_time.Day(), current_time.Hour(), current_time.Minute(), current_time.Second())
	dataString := []string{current_time_str}
	dataField := make(map[int64]*fmtrpc.DataField)
	// get current pressure set point
	currentState := s.stateStore.GetState()
	cState, ok := currentState.(controller.ControllerInitialState)
	if !ok {
		return state.ErrInvalidStateType
	}
	// figure out noise amplitude and offset
	for i := 0; i < 200; i++ {
		v := (rand.Float64()*5)+20
		dataField[int64(i)]= &fmtrpc.DataField{
			Name: fmt.Sprintf("Mass %v", i+1),
			Value: v,
		}
		dataString = append(dataString, fmt.Sprintf("%g", v))
	}
	//Write to csv in go routine
	errChan := make(chan error)
	go func(echan chan error) {
		err := writer.Write(dataString)
		if err != nil {
			echan<-err
		}
		echan<-nil
	}(errChan)
	
	err := s.stateStore.Dispatch(
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
	if err = <-errChan; err != nil {
		return err
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
	signalChan, unsub := s.stateStore.Subscribe(s.name)
	cleanUp := func() {
		unsub(s.stateStore, s.name)
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
		case <- signalChan:
			currentState := s.stateStore.GetState()
			cState, ok := currentState.(controller.ControllerInitialState)
			if !ok {
				s.Logger.Error().Msg(fmt.Sprintf("Invalid type %v expected %v\nStopping recording...", reflect.TypeOf(currentState), reflect.TypeOf(controller.ControllerInitialState{})))
				err := s.stopRecording()
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
				}
			}
			s.currentPressure = cState.Data[drivers.TelemetryPressureChannel].Value
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