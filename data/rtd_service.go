package data

import (
	"sync/atomic"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"google.golang.org/grpc"
)

var (
	FlukeName = "FLUKE"
	RgaName = "RGA"
	rpcEnumMap = map[fmtrpc.RecordService]string{
		fmtrpc.RecordService_FLUKE: FlukeName,
		fmtrpc.RecordService_RGA: RgaName,
	}
)

type RecordingState struct {
	RecordState 	bool
	ErrMsg			error
}

type RTDService struct {
	fmtrpc.UnimplementedDataCollectorServer
	Running				int32 // used atomically
	Listeners			int32 // used atomically
	Logger				*zerolog.Logger
	DataProviderChan	chan *fmtrpc.RealTimeData
	StartBrodcastChan	chan bool
	RecordingStateChans	map[string]chan *RecordingState
	ServiceRecStates	map[string]bool
	QuitChan			chan struct{}
	registeredProviders	int
}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewRTDService(log *zerolog.Logger) *RTDService {
	return &RTDService{
		QuitChan: make(chan struct{}),
		ServiceRecStates: make(map[string]bool),
		StartBrodcastChan: make(chan bool),
		RecordingStateChans: make(map[string]chan *RecordingState),
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
func (s *RTDService) RegisterDataProvider(serviceName string) {
	s.registeredProviders += 1 // maybe this should be atomic
	if _, ok := s.RecordingStateChans[serviceName]; !ok {
		s.RecordingStateChans[serviceName] = make(chan *RecordingState)
	}
	if _, ok := s.ServiceRecStates[serviceName]; !ok {
		s.ServiceRecStates[serviceName] = false
	}
}

// StartRecording is called by gRPC client and CLI to begin the data recording process with Fluke
func (s *RTDService) StartRecording(ctx context.Context, req *fmtrpc.RecordRequest) (*fmtrpc.RecordResponse, error) {
	switch req.Type {
	case fmtrpc.RecordService_FLUKE:
		s.RecordingStateChans[rpcEnumMap[req.Type]] <- &RecordingState{RecordState: true, ErrMsg: nil}
		resp := <-s.RecordingStateChans[rpcEnumMap[req.Type]]
		if resp.ErrMsg != nil {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start data recording: %v", err),
			}, err
		}
		s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.RecordState
	case fmtrpc.RecordService_RGA:
		// Leaving this until I can figure out how to make sure FLUKE is on and pressure is <= 0.00005 Torr
		return &fmtrpc.RecordResponse{
			Msg: "Could not start data recording: RGA unimplemented.",
		}, fmt.Errorf("Could not start data recording: RGA unimplemented.")
	}
	return &fmtrpc.RecordResponse{
		Msg: "Data recording successfully started.",
	}, nil
}

// StopRecording is called by gRPC client and CLI to end data recording process
func (s *RTDService) StopRecording(ctx context.Context, req *fmtrpc.StopRecRequest) (*fmtrpc.StopRecResponse, error) {
	s.RecordingStateChans[rpcEnumMap[req.Type]] <- &RecordingState{RecordState: false, ErrMsg: nil}
	resp := <-s.RecordingStateChans[rpcEnumMap[req.Type]]
	if resp.ErrMsg != nil {
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