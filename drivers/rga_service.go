package drivers

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync/atomic"
	"time"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
)

var (
	minimumPressure float64 = 0.00005
	minRgaPollingInterval int64 = 2 // Arbitrary not sure what this value should be
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
	c, err := connectToRGA()
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to RGA: %v", err)
	}
	return &RGAService{
		Logger: logger,
		QuitChan: make(chan struct{}),
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
	if resp.Fields["State"].Value.(RGASensorState) != Rga_SENSOR_STATE_READY {
		return fmt.Errorf("Sensor not ready: %v", resp.Fields["State"])
	}
	_, err = s.connection.AddBarchart("Bar1", 1, 200, Rga_PeakCenter, 5, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("Could not add Barchart: %v", err)
	}
	for i := 1; i < 6; i++ {
		_, err = s.connection.MeasurementAddMass(i)
		if err != nil {
			return fmt.Errorf("Could not add measurement mass: %v", err)
		}
	}
	_, err = s.connection.ScanAdd("Bar1")
	if err != nil {
		return fmt.Errorf("Could not add measurement to scan: %v", err)
	}
	_, err = s.connection.AnalogInputInterval(1, int(pol_int)*100000)
	if err != nil {
		return fmt.Errorf("Could not set input interval: %v", err)
	}
	_, err = s.connection.AnalogInputAverageCount(1, 10) // 1 reading every 0.3 seconds averaged every 3 seconds (10 readings averaged)
	if err != nil {
		return fmt.Errorf("Could not set input average count: %v", err)
	}
	_, err = s.connection.ScanStart(1)
	if err != nil {
		return fmt.Errorf("Could not start scan: %v", err)
	}
	_, err = s.connection.ReadResponse()
	_, err = s.connection.ReadResponse()
	// Header data for csv file
	headerData := []string{"Timestamp"}
	time.Sleep(time.Duration(pol_int)*time.Second)
	resp, err = s.connection.ReadMass()
	if err != nil {
		return err
	}
	for massPos, _ := range resp.Fields {
		headerData = append(headerData, fmt.Sprintf("Mass %d", massPos))
	}
	err = writer.Write(headerData)
	if err != nil {
		return err
	}
	// Now we begin recording
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// the actual data
	s.Logger.Info().Msg("Starting data recording...")
	go func() {
		for {
			select {
			case <-ticker.C:
				err = s.record(writer)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not write to %s: %v", file_name, err))
				}
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
func (s *RGAService) record(writer *csv.Writer) error {
	current_time := time.Now()
	current_time_str := fmt.Sprintf("%02d:%02d:%02d", current_time.Hour(), current_time.Minute(), current_time.Second())
	dataString := []string{current_time_str}
	dataField := make(map[int64]*fmtrpc.DataField)
	resp, err := s.connection.ReadMass()
	if err != nil {
		return err
	}
	i := 0
	for massPos, value := range resp.Fields {
		if atomic.LoadInt32(&s.Broadcasting) == 1 {
			dataField[int64(i)]= &fmtrpc.DataField{
				Name: fmt.Sprintf("Mass %d", massPos),
				Value: value.Value.(float64),
			}
		}
		dataString = append(dataString, fmt.Sprintf("%g", value.Value))
		i++
	}
	err = writer.Write(dataString)
	if err != nil {
		return err
	}
	if atomic.LoadInt32(&s.Broadcasting) == 1 {
		s.Logger.Info().Msg("Sending raw data to data buffer")
		dataFrame := &fmtrpc.RealTimeData{
			Source: s.name,
			IsScanning: true,
			Timestamp: current_time_str,
			Data: dataField,
		}
		s.OutputChan <- dataFrame // may need to go into a goroutine
		s.Logger.Info().Msg("After pushing into channel")
	}
	return nil
}

// stopRecording stops the data recording process
func (s *RGAService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return fmt.Errorf("Could not stop data recording. Data recording already stopped.")
	}
	s.QuitChan <- struct{}{}
	_, err := s.connection.Release()
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