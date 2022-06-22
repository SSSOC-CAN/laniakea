/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package data

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
	"sync/atomic"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	e "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/SSSOCPaulCote/gux"
	"google.golang.org/grpc"
)

var (
	TelemetryName = "TEL"
	RgaName = "RGA"
	RtdName = "RTD"
	rpcEnumMap = map[fmtrpc.RecordService]string{
		fmtrpc.RecordService_TELEMETRY: TelemetryName,
		fmtrpc.RecordService_RGA: RgaName,
	}
)

type StateType int32

const (
	BROADCASTING StateType = 0
	RECORDING StateType = 1
	ErrServiceNotRegistered = bg.Error("service not registered with RTD service")
	ErrServiceNotRecording = bg.Error("service not yet recording data")
)

type StateChangeMsg struct {
	Type	StateType
	State 	bool
	ErrMsg	error
	Msg		string
}

type RTDService struct {
	fmtrpc.UnimplementedDataCollectorServer
	Running				int32 // used atomically
	Logger				*zerolog.Logger
	StateChangeChans	map[string]chan *StateChangeMsg
	ServiceRecStates	map[string]bool
	name				string
	stateStore			*gux.Store
}

type InitialRtdState struct {
	AverageTemperature	float64
	RealTimeData		fmtrpc.RealTimeData
	TelPollingInterval	int64
}

type InitialCtrlState struct {
	PressureSetPoint	float64
	TemperatureSetPoint	float64
}

// Compile time check to ensure RTDService implements api.RestProxyService
var _ api.RestProxyService = (*RTDService)(nil)

//NewDataBuffer returns an instantiated DataBuffer struct
func NewRTDService(log *zerolog.Logger, s *gux.Store) *RTDService {
	return &RTDService{
		ServiceRecStates: 	make(map[string]bool),
		StateChangeChans: 	make(map[string]chan *StateChangeMsg),
		Logger: log,
		name: RtdName,
		stateStore: s,
	}
}

// RegisterWithGrpcServer registers the gRPC server to the RTDService
func (s *RTDService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterDataCollectorServer(grpcServer, s)
	return nil
}

// RegisterWithRestProxy registers the RTDService with the REST proxy
func(s *RTDService) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := fmtrpc.RegisterDataCollectorHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

// Start starts the RTD service. It also creates a buffered channel to send data to the RTD service
func (s *RTDService) Start() error {
	s.Logger.Info().Msg("Starting RTD Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return errors.ErrServiceAlreadyStarted
	}
	s.Logger.Info().Msg("RTD Service successfully started.")
	return nil
}

// Stop stops the RTD service. It closes all its channels, lifting the burdens from Data providers
func (s *RTDService) Stop() error {
	s.Logger.Info().Msg("Stopping RTD Service ...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return errors.ErrServiceAlreadyStopped
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

// StartRecording is called by gRPC client and CLI to begin the data recording process with Telemetry or RGA
func (s *RTDService) StartRecording(ctx context.Context, req *fmtrpc.RecordRequest) (*fmtrpc.RecordResponse, error) {
	if _, ok := s.StateChangeChans[rpcEnumMap[req.Type]]; !ok {
		return nil, e.Wrapf(ErrServiceNotRegistered, "Could not start %s data recording", rpcEnumMap[req.Type])
	}
	switch req.Type {
	case fmtrpc.RecordService_TELEMETRY:
		s.StateChangeChans[rpcEnumMap[req.Type]] <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil, Msg: fmt.Sprintf("%v:%v:%v", req.PollingInterval, req.OrgName, req.BucketName)}
		resp := <-s.StateChangeChans[rpcEnumMap[req.Type]]
		if resp.ErrMsg != nil {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: %v", rpcEnumMap[req.Type], resp.ErrMsg),
			}, resp.ErrMsg
		}
		s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.State
	case fmtrpc.RecordService_RGA:
		if !s.ServiceRecStates[TelemetryName] {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: telemetry service not yet recording data.", rpcEnumMap[req.Type]),
			}, e.Wrapf(ErrServiceNotRecording, "Could not start %s data recording", rpcEnumMap[req.Type])
		}
		s.StateChangeChans[rpcEnumMap[req.Type]] <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil, Msg: fmt.Sprintf("%v", req.OrgName)}
		resp := <-s.StateChangeChans[rpcEnumMap[req.Type]]
		if resp.ErrMsg != nil {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: %v", rpcEnumMap[req.Type], resp.ErrMsg),
			}, resp.ErrMsg
		}
		s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.State
	}
	return &fmtrpc.RecordResponse{
		Msg: "Data recording successfully started.",
	}, nil
}

// StopRecording is called by gRPC client and CLI to end data recording process
func (s *RTDService) StopRecording(ctx context.Context, req *fmtrpc.StopRecRequest) (*fmtrpc.StopRecResponse, error) {
	if _, ok := s.StateChangeChans[rpcEnumMap[req.Type]]; !ok {
		return nil, e.Wrapf(ErrServiceNotRegistered, "Could not start %s data recording", rpcEnumMap[req.Type])
	}
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
	lastSentRTDTimestamp := int64(0)
	rand.Seed(time.Now().UnixNano())
	subscriberName := s.name+utils.RandSeq(10)
	updateChan, unsub := s.stateStore.Subscribe(subscriberName)
	cleanUp := func() {
		unsub(s.stateStore, subscriberName)
	}
	defer cleanUp()
	for {
		select {	
		case <-updateChan:
			currentState := s.stateStore.GetState()
			RTD, ok := currentState.(InitialRtdState)
			if !ok {
				return errors.ErrInvalidType
			}
			if RTD.RealTimeData.Timestamp < lastSentRTDTimestamp {
				continue
			} 
			if err := updateStream.Send(&RTD.RealTimeData); err != nil {
				return err
			}
			lastSentRTDTimestamp = RTD.RealTimeData.Timestamp
		case <-updateStream.Context().Done():
			if errors.Is(updateStream.Context().Err(), context.Canceled) {
				return nil
			}
			return updateStream.Context().Err()
		}
	}
}