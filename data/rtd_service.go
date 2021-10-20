package data

import (
	"context"
	"fmt"
	//"reflect"
	"sync/atomic"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"google.golang.org/grpc"
)

var (
	FlukeName = "FLUKE"
	RgaName = "RGA"
	RtdName = "RTD"
	rpcEnumMap = map[fmtrpc.RecordService]string{
		fmtrpc.RecordService_FLUKE: FlukeName,
		fmtrpc.RecordService_RGA: RgaName,
	}
)

type StateType int32

const (
	BROADCASTING StateType = 0
	RECORDING StateType = 1
)

type StateChangeMsg struct {
	Type	StateType
	State 	bool
	ErrMsg	error
}

type RTDService struct {
	fmtrpc.UnimplementedDataCollectorServer
	Running				int32 // used atomically
	Listeners			int32 // used atomically
	Logger				*zerolog.Logger
	DataProviderChan	chan *fmtrpc.RealTimeData
	StateChangeChans	map[string]chan *StateChangeMsg
	ServiceRecStates	map[string]bool
	name				string
}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewRTDService(log *zerolog.Logger) *RTDService {
	return &RTDService{
		DataProviderChan: make(chan *fmtrpc.RealTimeData),
		ServiceRecStates: make(map[string]bool),
		StateChangeChans: make(map[string]chan *StateChangeMsg),
		Logger: log,
		name: RtdName,
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
	s.Logger.Info().Msg("RTD Service successfully started.")
	return nil
}

// Stop stops the RTD service. It closes all its channels, lifting the burdens from Data providers
func (s *RTDService) Stop() error {
	s.Logger.Info().Msg("Stopping RTD Service ...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop RTD service. Service already stopped.")
	}
	for _, channel := range s.StateChangeChans {
		close(channel)
	}
	s.Logger.Info().Msg("RTD Service stopped.")
	return nil
}

// Name satisfies the Service interface
func (s *RTDService) Name() string {
	return s.name
}

//RegisterDataProvider increments the counter by one
func (s *RTDService) RegisterDataProvider(serviceName string) {
	if _, ok := s.StateChangeChans[serviceName]; !ok {
		s.StateChangeChans[serviceName] = make(chan *StateChangeMsg)
	}
	if _, ok := s.ServiceRecStates[serviceName]; !ok {
		s.ServiceRecStates[serviceName] = false
	}
}

// StartRecording is called by gRPC client and CLI to begin the data recording process with Fluke
func (s *RTDService) StartRecording(ctx context.Context, req *fmtrpc.RecordRequest) (*fmtrpc.RecordResponse, error) {
	switch req.Type {
	case fmtrpc.RecordService_FLUKE:
		s.StateChangeChans[rpcEnumMap[req.Type]] <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil}
		resp := <-s.StateChangeChans[rpcEnumMap[req.Type]]
		if resp.ErrMsg != nil {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: %v", rpcEnumMap[req.Type], resp.ErrMsg),
			}, resp.ErrMsg
		}
		s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.State
	case fmtrpc.RecordService_RGA:
		// Leaving this until I can figure out how to make sure FLUKE is on and pressure is <= 0.00005 Torr
		if !s.ServiceRecStates[FlukeName] {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: Fluke Service not yet recording data.", rpcEnumMap[req.Type]),
			}, fmt.Errorf("Could not start %s data recording: Fluke Service not yet recording data.", rpcEnumMap[req.Type])
		}
		s.StateChangeChans[rpcEnumMap[req.Type]] <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil}
		resp := <-s.StateChangeChans[rpcEnumMap[req.Type]]
		if resp.ErrMsg != nil {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: %v", rpcEnumMap[req.Type], resp.ErrMsg),
			}, resp.ErrMsg
		}
	}
	return &fmtrpc.RecordResponse{
		Msg: "Data recording successfully started.",
	}, nil
}

// StopRecording is called by gRPC client and CLI to end data recording process
func (s *RTDService) StopRecording(ctx context.Context, req *fmtrpc.StopRecRequest) (*fmtrpc.StopRecResponse, error) {
	s.StateChangeChans[rpcEnumMap[req.Type]] <- &StateChangeMsg{Type: RECORDING, State: false, ErrMsg: nil}
	resp := <-s.StateChangeChans[rpcEnumMap[req.Type]]
	if resp.ErrMsg != nil {
		return &fmtrpc.StopRecResponse{
			Msg: "Could not stop data recording. Data recording already stopped.",
		}, resp.ErrMsg
	}
	s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.State
	return &fmtrpc.StopRecResponse{
		Msg: "Data recording successfully stopped.",
	}, nil
}

// SubscribeDataStream return a uni-directional stream (server -> client) to provide realtime data to the client
func (s *RTDService) SubscribeDataStream(req *fmtrpc.SubscribeDataRequest, updateStream fmtrpc.DataCollector_SubscribeDataStreamServer) error {
	s.Logger.Info().Msg("Have a new data listener.")
	_ = atomic.AddInt32(&s.Listeners, 1)
	for _, channel := range s.StateChangeChans {
		channel <- &StateChangeMsg{Type: BROADCASTING, State: true, ErrMsg: nil}
		resp := <- channel
		if resp.ErrMsg != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Could not change broadcast state: %v", resp.ErrMsg))
			return resp.ErrMsg
		}
	}
	for {
		select {
		case RTD := <-s.DataProviderChan:
			s.Logger.Info().Msg("Received data from data buffer. Sending out to server-client stream...")
			if err := updateStream.Send(RTD); err != nil {
				_ = atomic.AddInt32(&s.Listeners, -1)
				if atomic.LoadInt32(&s.Listeners) == 0 {
					for _, channel := range s.StateChangeChans {
						channel <- &StateChangeMsg{Type: BROADCASTING, State:false, ErrMsg: nil}
						resp := <- channel
						if resp.ErrMsg != nil {
							s.Logger.Error().Msg(fmt.Sprintf("Could not change broadcast state: %v", resp.ErrMsg))
							return resp.ErrMsg
						}
					}
				}
				return err
			}
		case <-updateStream.Context().Done():
			_ = atomic.AddInt32(&s.Listeners, -1)
			if atomic.LoadInt32(&s.Listeners) == 0 {
				for _, channel := range s.StateChangeChans {
					channel <- &StateChangeMsg{Type: BROADCASTING, State:false, ErrMsg: nil}
					resp := <- channel
					if resp.ErrMsg != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not change broadcast state: %v", resp.ErrMsg))
						return resp.ErrMsg
					}
				}
			}
			return updateStream.Context().Err()
		}
	}
}