// +build !fluke !windows !386

package drivers

// import (
// 	"encoding/csv"
// 	"fmt"
// 	"math/rand"
// 	"os"
// 	"strconv"
// 	"time"

// 	"github.com/SSSOC-CAN/fmtd/data"
// 	"github.com/SSSOC-CAN/fmtd/fmtrpc"
// 	"github.com/SSSOC-CAN/fmtd/state"
// 	"github.com/SSSOC-CAN/fmtd/utils"
// 	"github.com/rs/zerolog"
// )

// var (
// 	FlukeInitialState = fmtrpc.RealTimeData{}
// 	FlukeReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
// 		// assert type of s
// 		_, ok := s.(fmtrpc.RealTimeData)
// 		if !ok {
// 			return nil, state.ErrInvalidStateType
// 		}
// 		// switch case action
// 		switch a.Type {
// 		case "fluke/update":
// 			// assert type of payload
// 			newState, ok := a.Payload.(fmtrpc.RealTimeData)
// 			if !ok {
// 				return nil, state.ErrInvalidPayloadType
// 			}
// 			return newState, nil
// 		default:
// 			return nil, state.ErrInvalidAction
// 		} 
// 	}
// 	DefaultPollingInterval int64 = 5
// 	minPollingInterval     int64 = 5
// 	ErrAlreadyRecording			 = fmt.Errorf("Could not start recording. Data recording already started")
// 	ErrAlreadyStoppedRecording   = fmt.Errorf("Could not stop data recording. Data recording already stopped.")
// )

// // FlukeService is a struct for holding all relevant attributes to interfacing with the Fluke DAQ
// type FlukeService struct {
// 	Running     			int32 // atomic
// 	Recording				int32 // atomic
// 	stateStore				*state.Store
// 	Logger     				*zerolog.Logger
// 	name					string
// 	QuitChan   				chan struct{}
// 	CancelChan				chan struct{}
// 	outputDir  				string
// 	StateChangeChan			chan *data.StateChangeMsg
// 	filepath				string
// }

// // A compile time check to make sure that FlukeService fully implements the data.Service interface
// var _ data.Service = (*FlukeService) (nil)


// // NewFlukeService creates a new Fluke Service object which will use the drivers for the Fluke DAQ software
// func NewFlukeService(logger *zerolog.Logger, outputDir string, store *state.Store) (*FlukeService, error) {
// 	return &FlukeService{
// 		stateStore:   	  store,
// 		Logger:       	  logger,
// 		QuitChan:     	  make(chan struct{}),
// 		CancelChan:   	  make(chan struct{}),
// 		outputDir:   	  outputDir,
// 		name: 	      	  data.FlukeName,
// 	}, nil
// }

// // Start starts the service. Returns an error if any issues occur
// func (s *FlukeService) Start() error {
// 	s.Logger.Info().Msg("Starting Fluke Service...")
// 	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
// 		return fmt.Errorf("Could not start Fluke service: service already started.")
// 	}
// 	s.Logger.Info().Msg("Connection to Fluke DAQ is not currently active. Please recompile fmtd as follows `$ go install -tags \"fluke windows 386\"`")
// 	go s.ListenForRTDSignal()
// 	s.Logger.Info().Msg("Fluke Service started.")
// 	return nil
// }

// // Stop stops the service. Returns an error if any issues occur
// func (s *FlukeService) Stop() error {
// 	s.Logger.Info().Msg("Stopping Fluke Service...")
// 	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
// 		return fmt.Errorf("Could not stop Fluke service: service already stopped.")
// 	}
// 	if atomic.LoadInt32(&s.Recording) == 1 {
// 		err := s.stopRecording()
// 		if err != nil {
// 			return fmt.Errorf("Could not stop Fluke service: %v", err)
// 		}
// 	}
// 	close(s.CancelChan)
// 	close(s.QuitChan)
// 	s.Logger.Info().Msg("Fluke Service stopped.")
// 	return nil
// }

// // Name satisfies the fmtd.Service interface
// func (s *FlukeService) Name() string {
// 	return s.name
// }

// // StartRecording starts the recording process by creating a csv file and inserting the header row into the file and returns a quit channel and error message
// func (s *FlukeService) startRecording(pol_int int64) error {
// 	if pol_int < minPollingInterval && pol_int != 0 {
// 		return fmt.Errorf("Inputted polling interval smaller than minimum value: %v", minPollingInterval)
// 	} else if pol_int == 0 { //No polling interval provided
// 		pol_int = DefaultPollingInterval
// 	}
// 	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
// 		return ErrAlreadyRecording
// 	}
// 	current_time := time.Now()
// 	file_name := fmt.Sprintf("%s/%d-%02d-%02d-fluke.csv", s.outputDir, current_time.Year(), current_time.Month(), current_time.Day())
// 	file_name = utils.UniqueFileName(file_name)
// 	file, err := os.Create(file_name)
// 	if err != nil {
// 		return fmt.Errorf("Could not create file %v: %v", file, err)
// 	}
// 	s.filepath = file_name
// 	writer := csv.NewWriter(file)
// 	// headers
// 	headerData := []string{"Timestamp"}
// 	for i := 0; i < 136; i++ {
// 		headerData = append(headerData, fmt.Sprintf("Value #%v", i+1))
// 	}
// 	err = writer.Write(headerData)
// 	if err != nil {
// 		return err
// 	}
// 	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
// 	// the actual data
// 	s.Logger.Info().Msg("Starting data recording...")
// 	go func() {
// 		for {
// 			select {
// 			case <-ticker.C:
// 				err = s.record(writer)
// 				if err != nil {
// 					s.Logger.Error().Msg(fmt.Sprintf("Could not write to %s: %v", file_name, err))
// 				}
// 			case <-s.CancelChan:
// 				ticker.Stop()
// 				file.Close()
// 				writer.Flush()
// 				s.Logger.Info().Msg("Data recording stopped.")
// 				return
// 			case <-s.QuitChan:
// 				ticker.Stop()
// 				file.Close()
// 				writer.Flush()
// 				s.Logger.Info().Msg("Data recording stopped.")
// 				return
// 			}
// 		}
// 	}()
// 	return nil
// }

// // record records the live data from Fluke and inserts it into a csv file and passes it to the RTD service
// func (s *FlukeService) record(writer *csv.Writer) error {
// 	current_time := time.Now()
// 	current_time_str := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", current_time.Year(), current_time.Month(), current_time.Day(), current_time.Hour(), current_time.Minute(), current_time.Second())
// 	dataString := []string{current_time_str}
// 	dataField := make(map[int64]*fmtrpc.DataField)
// 	for i := 0; i < 136; i++ {
// 		v := (rand.Float64()*5)+20
// 		dataField[int64(i)]= &fmtrpc.DataField{
// 			Name: fmt.Sprintf("Value %v", i+1),
// 			Value: v,
// 		}
// 		dataString = append(dataString, fmt.Sprintf("%g", v))
// 	}
// 	//Write to csv in go routine
// 	errChan := make(chan error)
// 	go func(echan chan error) {
// 		err := writer.Write(dataString)
// 		if err != nil {
// 			echan<-err
// 		}
// 		echan<-nil
// 	}(errChan)
	
// 	err := s.stateStore.Dispatch(
// 		state.Action{
// 			Type: 	 "fluke/update",
// 			Payload: fmtrpc.RealTimeData{
// 				Source: s.name,
// 				IsScanning: false,
// 				Timestamp: current_time.UnixMilli(),
// 				Data: dataField,
// 			},
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("Could not update state: %v", err)
// 	}
// 	if err = <-errChan; err != nil {
// 		return err
// 	}
// 	return nil
// }

// // stopRecording sends an empty struct down the CancelChan to innitiate the stop recording process
// func (s *FlukeService) stopRecording() error {
// 	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
// 		return ErrAlreadyStoppedRecording
// 	}
// 	s.CancelChan<-struct{}{}
// 	return nil
// }

// //CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
// func (s *FlukeService) ListenForRTDSignal() {
// 	for {
// 		select {
// 		case msg := <-s.StateChangeChan:
// 			switch msg.Type {
// 			case data.RECORDING:
// 				if msg.State {
// 					var err error
// 					if n, err := strconv.ParseInt(msg.Msg, 10, 64); err != nil {
// 						err = s.startRecording(DefaultPollingInterval)
// 					} else {
// 						err = s.startRecording(n)
// 					}
// 					if err != nil {
// 						s.Logger.Error().Msg(fmt.Sprintf("Could not start recording: %v", err))
// 						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
// 					} else {
// 						s.Logger.Info().Msg("Started recording.")
// 						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: nil, Msg: s.filepath}
// 					}
// 				} else {
// 					s.Logger.Info().Msg("Stopping data recording...")
// 					err := s.stopRecording()
// 					if err != nil {
// 						s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
// 						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: fmt.Errorf("Could not stop recording: %v", err)}
// 					} else {
// 						s.Logger.Info().Msg("Stopped recording.")
// 						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: nil}
// 					}
// 				}
// 			}
// 		case <-s.QuitChan:
// 			return
// 		}
// 	}
// }

// // RegisterWithRTDService adds the RTD Service channels to the Fluke Service Struct and incrememnts the number of registered data providers on the RTD
// func (s *FlukeService) RegisterWithRTDService(rtd *data.RTDService) {
// 	rtd.RegisterDataProvider(s.name)
// 	s.StateChangeChan = rtd.StateChangeChans[s.name]
// }