package data

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"time"
	"sync/atomic"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
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
	defaultTCPBufferSize int64 = 1024
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
	Msg		string
}

type RTDService struct {
	fmtrpc.UnimplementedDataCollectorServer
	Running				int32 // used atomically
	TCPRunning			int32 // used atomically
	TCPPort				int64
	TCPAddr				string
	Logger				*zerolog.Logger
	StateChangeChans	map[string]chan *StateChangeMsg
	ServiceRecStates	map[string]bool
	ServiceFilePaths	map[fmtrpc.RecordService]string
	name				string
	tcpServer			net.Listener
	stateStore			*state.Store
}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewRTDService(log *zerolog.Logger, tcpAddr string, tcpPort int64, s *state.Store) *RTDService {
	return &RTDService{
		ServiceRecStates: 	make(map[string]bool),
		StateChangeChans: 	make(map[string]chan *StateChangeMsg),
		ServiceFilePaths:	make(map[fmtrpc.RecordService]string),
		Logger: log,
		name: RtdName,
		TCPAddr: tcpAddr,
		TCPPort: tcpPort,
		stateStore: s,
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
		s.StateChangeChans[rpcEnumMap[req.Type]] <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil, Msg: fmt.Sprintf("%v", req.PollingInterval)}
		resp := <-s.StateChangeChans[rpcEnumMap[req.Type]]
		if resp.ErrMsg != nil {
			return &fmtrpc.RecordResponse{
				Msg: fmt.Sprintf("Could not start %s data recording: %v", rpcEnumMap[req.Type], resp.ErrMsg),
			}, resp.ErrMsg
		}
		s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.State
		s.ServiceFilePaths[req.Type] = resp.Msg
	case fmtrpc.RecordService_RGA:
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
		s.ServiceRecStates[rpcEnumMap[req.Type]] = resp.State
		s.ServiceFilePaths[req.Type] = resp.Msg
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
	updateChan := make(chan struct{})
	lastSentRTDTimestamp := int64(0)
	idx, unsub := s.stateStore.Subscribe(func() {
		updateChan<-struct{}{}
	})
	cleanUp := func() {
		unsub(s.stateStore, idx)
		time.Sleep(10*time.Second)
		close(updateChan)
	}
	defer cleanUp()
	for {
		select {	
		case <-updateChan:
			currentState := s.stateStore.GetState()
			RTD, ok := currentState.(fmtrpc.RealTimeData)
			if !ok {
				return fmt.Errorf("Invalid type %v", reflect.TypeOf(currentState))
			}
			if RTD.Timestamp < lastSentRTDTimestamp {
				continue
			} 
			if err := updateStream.Send(&RTD); err != nil {
				return err
			}
			lastSentRTDTimestamp = RTD.Timestamp
		case <-updateStream.Context().Done():
			if errors.Is(updateStream.Context().Err(), context.Canceled) {
				return nil
			}
			return updateStream.Context().Err()
		}
	}
}

// startTCPServer starts a tcp server and starts a goroutine for listening
func (s *RTDService) startTCPServer() (error) {
	tcpS, err := net.Listen("tcp", s.TCPAddr+":"+strconv.FormatInt(s.TCPPort, 10))
    if err != nil {
        return fmt.Errorf("Unable to listen at %s:%v: %v", s.TCPAddr, s.TCPPort, err)
    }
	s.tcpServer = tcpS
	return nil
}

//tcpServerListen listens for incoming connections and handles them
func (s *RTDService) tcpServerListen(file_type fmtrpc.RecordService) {
	shutdown := func() {
		s.Logger.Info().Msg("Shutting down tcp server...")
		s.tcpServer.Close()
		if ok := atomic.CompareAndSwapInt32(&s.TCPRunning, 1, 0); !ok {
			s.Logger.Error().Msg("TCP server already stopped")
		}
		s.Logger.Info().Msg("TCP shutdown complete.")
	}
	defer shutdown()
	conn, err := s.tcpServer.Accept()
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Unable to accept connection: %v", err))
		return
	}
	s.tcpConnHandler(conn, file_type)
	return
}

// tcpConnHandler handles the incoming tcp connection and sends the requested file
func (s *RTDService) tcpConnHandler(conn net.Conn, file_type fmtrpc.RecordService) {
    defer conn.Close()
	filepath := s.ServiceFilePaths[file_type]
	if filepath == "" {
		s.Logger.Error().Msg(fmt.Sprintf("Unable to find file path for service type %v", file_type))
		return
	}
	fileBuf := make([]byte, defaultTCPBufferSize)
	file, err := os.Open(filepath)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Unable to open file at path %v", filepath))
		return
	}
	defer file.Close()
	s.Logger.Info().Msg("Sending file...")
	_, err = io.CopyBuffer(conn, file, fileBuf)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Error sending file: %v", err))
		return
	}
	s.Logger.Info().Msg("File sent.")
	return
}

// DownloadHistoricalData is a gRPC endpoint to establish a tcp connection and upload a csv file to the remote client app
func (s *RTDService) DownloadHistoricalData(ctx context.Context, req *fmtrpc.HistoricalDataRequest) (*fmtrpc.HistoricalDataResponse, error) {
	if ok := atomic.CompareAndSwapInt32(&s.TCPRunning, 0, 1); !ok {
		s.Logger.Info().Msg("TCP server already started")
	} else {
		s.Logger.Info().Msg("Starting TCP server...")
		err := s.startTCPServer()
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot start TCP server: %v", err))
			return nil, err
		}
		s.Logger.Info().Msg("TCP server started")
		go s.tcpServerListen(req.Source)
	}
	return &fmtrpc.HistoricalDataResponse{
		ServerPort: s.TCPPort,
		BufferSize: defaultTCPBufferSize,
	}, nil
}