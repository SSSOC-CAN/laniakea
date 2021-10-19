package drivers

import (
	"fmt"
	"net"
	"sync/atomic"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
)

type RGAService struct {
	Running				int32 //atomically
	Recording			int32 //atomically
	Broadcasting		int32 //atomically
	Logger				*zerolog.Logger
	OutputChan			chan *fmtrpc.RealTimeData
	StartBrodcastChan	chan bool
	QuitChan			chan struct{}
	outputDir  			string
	connection			*net.Conn
	// channel to send data to RealTimeDataService
}

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
		conncetion: c,
	}, nil
}

// Start starts the RGA service. It does NOT start the data recording process
func (s *RGAService) Start() error {
	s.Logger.Info().Msg("Starting RGA Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start RGA service. Service already started.")
	}
	go s.CheckIfBroadcasting()
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
	close(s.StartBrodcastChan)
	close(s.OutputChan)
	s.conncetion.Close()
	s.Logger.Info().Msg("RGA Service successfully stopped.")
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *RGAService) CheckIfBroadcasting() {
	for {
		select {
		case state := <-s.StartBrodcastChan:
			if state {
				if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 0, 1); !ok {
					s.Logger.Warn().Msg("Could not start broadcasting to RTD Service.")
				}
			} else {
				if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 1, 0); !ok {
					s.Logger.Warn().Msg("Could not stop broadcasting to RTD Service.")
				}
			}
		case <-s.QuitChan:
			return
		}
	}
}

// RegisterWithRTDService adds the RTD Service channels to the RGA Service Struct and incrememnts the number of registered data providers on the RTD
func (s *RGAService) RegisterWithRTDService(rtd *data.RTDService) {
	s.OutputChan = rtd.DataProviderChan
	s.StartBrodcastChan = rtd.StartBrodcastChan
	rtd.RegisterDataProvider()
}