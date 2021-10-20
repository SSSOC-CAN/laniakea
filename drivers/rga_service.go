package drivers

import (
	"fmt"
	"net"
	"sync/atomic"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
)

var (
	minimumPressure float64 = 0.00005
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
	connection			net.Conn
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
		connection: c,
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

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *RGAService) ListenForRTDSignal() {
	for {
		select {
		case msg := <- s.StateChangeChan:
			switch msg.Type {
			case data.BROADCASTING:
				if msg.State {
					if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 0, 1); !ok {
						s.Logger.Warn().Msg("Could not start broadcasting to RTD Service.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.BROADCASTING, State: false, ErrMsg: fmt.Errorf("Could not change broadcasting state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				} else {
					if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 1, 0); !ok {
						s.Logger.Warn().Msg("Could not stop broadcasting to RTD Service.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.BROADCASTING, State: true, ErrMsg: fmt.Errorf("Could not change broadcasting state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				}
			case data.RECORDING:
				if msg.State {
					if s.currentPressure == 0 || s.currentPressure > minimumPressure {
						s.Logger.Warn().Msg(fmt.Sprintf("Could not start RGA recording: current chamber pressure is too high. Current: %.6f Torr\tMinimum: %.5f Torr", s.currentPressure, minimumPressure))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start RGA recording: current chamber pressure is too high.\nCurrent: %.6f Torr\nMinimum: %.5f Torr", s.currentPressure, minimumPressure)}
					} else {
						if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
							s.Logger.Warn().Msg("Could not change recording state.")
							s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not change recording state.")}
						} else {
							s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
						}
					}
				} else {
					if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
						s.Logger.Warn().Msg("Could not change recording state.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: fmt.Errorf("Could not change recording state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				}
			}
		case p := <- s.PressureChan:
			s.currentPressure = p
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