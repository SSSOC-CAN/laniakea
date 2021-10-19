package data

import (
	"sync/atomic"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"google.golang.org/grpc"
)

type RTDService struct {
	fmtrpc.UnimplementedDataCollectorServer
	Running				int32 // used atomically
	Listeners			int32 // used atomically
	Logger				*zerolog.Logger
	DataProviderChan	chan *fmtrpc.RealTimeData
	StartBrodcastChan	chan bool
	QuitChan			chan struct{}
	registeredProviders	int
}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewRTDService(log *zerolog.Logger) *RTDService {
	return &RTDService{
		QuitChan: make(chan struct{}),
		StartBrodcastChan: make(chan bool),
		Logger: log,
		wg: wg,
	}
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (s *RTDService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterDataCollectorServer(grpcServer, s)
	return nil
}

// Start starts the RTD service. It also creates a buffered channel to send data to the RTD service
func (s *RTDService) Start() error {
	s.Logger.Info().Msg("Starting RTD Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start RTD service. Service already started.")
	}
	s.DataProviderChan = make(chan *fmtrpc.RealTimeData, s.registeredProviders) //important to register all providers before starting the service
	s.Logger.Info().Msg("RTD Service successfully started.")
	return nil
}

//RegisterDataProvider increments the counter by one
func (s *RTDService) RegisterDataProvider() {
	s.registeredProviders += 1 // maybe this should be atomic
}

// StartRecording is called by gRPC client and CLI to begin the data recording process with Fluke
func (s *RTDService) StartRecording(ctx context.Context, req *fmtrpc.RecordRequest) (*fmtrpc.RecordResponse, error) {
	// TODO:SSSOCPaulCote - Send data using channel to start recording
	err := s.startRecording(req.PollingInterval)
	if err != nil {
		return &fmtrpc.RecordResponse{
			Msg: fmt.Sprintf("Could not start data recording: %v", err),
		}, err
	}
	return &fmtrpc.RecordResponse{
		Msg: "Data recording successfully started.",
	}, nil
}

// StopRecording is called by gRPC client and CLI to end data recording process
func (s *RTDService) StopRecording(ctx context.Context, req *fmtrpc.StopRecRequest) (*fmtrpc.StopRecResponse, error) {
	// TODO:SSSOCPaulCote - Send data using channel to stop recording
	err := s.stopRecording()
	if err != nil {
		return &fmtrpc.StopRecResponse{
			Msg: "Could not stop data recording. Data recording already stopped.",
		}, err
	}
	return &fmtrpc.StopRecResponse{
		Msg: "Data recording successfully stopped.",
	}, nil
}

// SubscribeDataStream return a uni-directional stream (server -> client) to provide realtime data to the client
func (s *RTDService) SubscribeDataStream(req *fmtrpc.SubscribeDataRequest, updateStream fmtrpc.DataCollector_SubscribeDataStreamServer) error {
	_ = atomic.AddInt32(&s.Listeners, 1)
	s.Logger.Info().Msg("Have a new data listener.")
	s.StartBrodcastChan <- true
	for {
		select {
		case RTD := <-s.DataProviderChan:
			s.Logger.Info().Msg("Received data from data buffer. Sending out to server-client stream...")
			if err := updateStream.Send(RTD); err != nil {
				_ = atomic.AddInt32(&s.Listeners, -1)
				if atomic.LoadInt32(&s.Listeners) == 0 {
					s.StartBrodcastChan <- false
				}
				return err
			}
		case <-updateStream.Context().Done():
			_ = atomic.AddInt32(&s.Listeners, -1)
			if atomic.LoadInt32(&s.Listeners) == 0 {
				s.StartBrodcastChan <- false
			}
			return updateStream.Context().Err()
		case <-s.QuitChan:
			atomic.StoreInt32(&s.Listeners, 0)
			s.StartBrodcastChan <- false
			return nil
		}
	}
}