package drivers

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
)

var (
	minimumPressure float64 = 0.00005
	minRgaPollingInterval int64 = 15 // Arbitrary not sure what this value should be - After manual testing, value will be 15 seconds
)

type RGAService struct {
	Running				int32 //atomically
	Recording			int32 //atomically
	Broadcasting		int32 //atomically
	Logger				*zerolog.Logger
	OutputChan			chan *fmtrpc.RealTimeData
	PressureChan		chan float64
	StateChangeChan		chan *data.StateChangeMsg
	QuitChan			chan struct{}
	CancelChan			chan struct{}
	outputDir  			string
	connection			*RGAConnection
	name 				string
	currentPressure		float64
	// channel to send data to RealTimeDataService
}

// A compile time check to make sure that RGAService fully implements the data.Service interface
var _ data.Service = (*RGAService) (nil)

// NewRGAService creates an instance of the RGAService struct. It also establishes a connection to the RGA device
func NewRGAService(logger *zerolog.Logger, outputDir string) (*RGAService, error) {
	c, err := ConnectToRGA()
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to RGA: %v", err)
	}
	return &RGAService{
		Logger: logger,
		QuitChan: make(chan struct{}),
		CancelChan: make(chan struct{}),
		outputDir: outputDir,
		connection: &RGAConnection{c},
		name: data.RgaName,
	}, nil
}

// Start starts the RGA service. It does NOT start the data recording process
func (s *RGAService) Start() error {
	s.Logger.Info().Msg("Starting RGA Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start RGA service. Service already started.")
	}
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
	close(s.QuitChan)
	s.connection.Close()
	s.Logger.Info().Msg("RGA Service successfully stopped.")
	return nil
}

// Name satisfies the fmtd.Service interface
func (s *RGAService) Name() string {
	return s.name
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
		return fmt.Errorf("Could not start recording. Data recording already started")
	}
	// Create or Open csv
	current_time := time.Now()
	file_name := fmt.Sprintf("%s/%02d-%02d-%d-rga.csv", s.outputDir, current_time.Day(), current_time.Month(), current_time.Year())
	file_name = utils.UniqueFileName(file_name)
	file, err := os.Create(file_name)
	if err != nil {
		return fmt.Errorf("Could not create file %v: %v", file, err)
	}
	writer := csv.NewWriter(file)
	// InitMsg
	err = s.connection.InitMsg()
	if err != nil {
		return fmt.Errorf("Unable to communicate with RGA: %v", err)
	}
	// Setup Data
	_, err = s.connection.Control(utils.AppName, utils.AppVersion)
	if err != nil {
		return err
	}
	resp, err := s.connection.SensorState()
	if err != nil {
		return err
	}
	if resp.Fields["State"].Value.(string) != Rga_SENSOR_STATE_INUSE {
		return fmt.Errorf("Sensor not ready: %v", resp.Fields["State"])
	}
	_, err = s.connection.AddBarchart("Bar1", 1, 200, Rga_PeakCenter, 5, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("Could not add Barchart: %v", err)
	}
	_, err = s.connection.AddPeakJump("PeakJump1", Rga_PeakCenter, 5, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("Could not add PeakJump: %v", err)
	}
	for i := 1; i < 6; i++ { // TODO:SSSOCPaulCote - Change this for test next week
		_, err = s.connection.MeasurementAddMass(i)
		if err != nil {
			return fmt.Errorf("Could not add measurement mass: %v", err)
		}
	}
	_, err = s.connection.ScanAdd("Bar1")
	if err != nil {
		return fmt.Errorf("Could not add measurement to scan: %v", err)
	}
	// Now we begin recording
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// the actual data
	s.Logger.Info().Msg("Starting data recording...")
	go func() {
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
		// Start scan
		_, err := s.connection.ScanStart(1)
		if err != nil {
			return fmt.Errorf("Could not start scan: %v", err)
		}
		_, err = s.connection.FilamentControl("On")
		if err != nil {
			return fmt.Errorf("Could not turn on filament: %v", err)
		}
		for {
			resp, err := s.connection.ReadResponse()
			if err != nil {
				return fmt.Errorf("Could not read response: %v", err)
			}
			if resp.ErrMsg.CommandName == massReading {
				headerData = append(headerData, strconv.FormatInt(resp.Fields["MassPosition"].Value.(int64), 10))
				if resp.Fields["MassPosition"].Value.(int64) == int64(200) {
					break
				}
			}
		}
		err = writer.Write(headerData)
		if err != nil {
			return err
		}
		return nil
	}
	current_time := time.Now()
	current_time_str := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d", current_time.Day(), current_time.Month(), current_time.Year(), current_time.Hour(), current_time.Minute(), current_time.Second())
	dataString := []string{current_time_str}
	dataField := make(map[int64]*fmtrpc.DataField)
	// Start scan
	_, err := s.connection.ScanResume(1)
	if err != nil {
		return fmt.Errorf("Could not resume scan: %v", err)
	}
	for {
		resp, err := s.connection.ReadResponse()
		if err != nil {
			return fmt.Errorf("Could not read response: %v", err)
		}
		if resp.ErrMsg.CommandName == massReading {
			if atomic.LoadInt32(&s.Broadcasting) == 1 {
				massPos := resp.Fields["MassPosition"].Value.(int64)
				dataField[massPos]= &fmtrpc.DataField{
					Name: fmt.Sprintf("Mass %s", strconv.FormatInt(massPos, 10)),
					Value: resp.Fields["Value"].Value.(float64),
				}
			}
			dataString = append(dataString, fmt.Sprintf("%g", resp.Fields["Value"].Value.(float64)))
			if resp.Fields["MassPosition"].Value.(int64) == int64(200) {
				break
			}
		}
	}
	err = writer.Write(dataString)
	if err != nil {
		return err
	}
	if atomic.LoadInt32(&s.Broadcasting) == 1 {
		dataFrame := &fmtrpc.RealTimeData{
			Source: s.name,
			IsScanning: true,
			Timestamp: current_time.UnixMilli(),
			Data: dataField,
		}
		s.OutputChan <- dataFrame // may need to go into a goroutine
	}
	return nil
}

// stopRecording stops the data recording process
func (s *RGAService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return fmt.Errorf("Could not stop data recording. Data recording already stopped.")
	}
	s.CancelChan <- struct{}{}
	_, err := s.connection.FilamentControl("Off")
	if err != nil {
		return fmt.Errorf("Could nto safely turn off filament: %v", err)
	}
	_, err = s.connection.Release()
	if err != nil {
		return fmt.Errorf("Could not safely release control of RGA: %v", err)
	}
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *RGAService) ListenForRTDSignal() {
	for {
		select {
		case msg := <- s.StateChangeChan:
			switch msg.Type {
			case data.BROADCASTING:
				if msg.State {
					if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 0, 1); !ok {
						s.Logger.Warn().Msg("Could not start broadcasting to RTD Service: already broadcasting")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.BROADCASTING, State: false, ErrMsg: fmt.Errorf("Could not change broadcasting state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				} else {
					if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 1, 0); !ok {
						s.Logger.Warn().Msg("Could not stop broadcasting to RTD Service: already stopped broadcasting")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.BROADCASTING, State: true, ErrMsg: fmt.Errorf("Could not change broadcasting state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				}
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
		case p := <- s.PressureChan:
			s.currentPressure = p
			if p >= 0.00005 && atomic.LoadInt32(&s.Recording) == 1 {
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
	s.OutputChan = rtd.DataProviderChan
	s.StateChangeChan = rtd.StateChangeChans[s.name]
}

// RegisterWithFlukeService adds the Fluke Service pressure channel to the RGA service struct in order to have realtime pressure measurements.
func (s *RGAService) RegisterWithFlukeService(f *FlukeService) {
	s.PressureChan = f.PressureChan
	f.IsRGAReady = true
}