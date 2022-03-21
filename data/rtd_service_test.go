package data

import (
	"context"
	"fmt"
	"os"
	"net"
	"testing"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	defaultTestingTCPAddr = "localhost"
	defaultTestingTCPPort int64 = 10024
	testReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		_, ok := s.(fmtrpc.RealTimeData)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			return newState, nil
		default:
			return nil, state.ErrInvalidAction
		} 
	}
	bufSize = 1 * 1024 * 1024
	lis *bufconn.Listener
	polling_interval int64
)

// TestRTDServiceStartStop tests if we can initialize, start and stop the RTD service
func TestRTDServiceStartStop(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	stateStore := state.CreateStore(fmtrpc.RealTimeData{}, testReducer)
	rtdService := NewRTDService(&logger,
		defaultTestingTCPAddr,
		defaultTestingTCPPort,
		stateStore,
	)
	err := rtdService.Start()
	if err != nil {
		t.Fatalf("Could not start rtdService: %v", err)
	}
	err = rtdService.Stop()
	if err != nil {
		t.Fatalf("Could not stop rtdService: %v", err)
	}
}

// init initializes the gRPC server
func Init(t *testing.T) func() {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	stateStore := state.CreateStore(fmtrpc.RealTimeData{}, testReducer)
	rtdService := NewRTDService(&logger, defaultTestingTCPAddr, defaultTestingTCPPort, stateStore)
	err := rtdService.Start()
	if err != nil {
		t.Fatalf("Could not start RTD Service: %v", err)
	}
	_ = rtdService.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	cleanUp := func() {
		_ = rtdService.Stop()
		grpcServer.Stop()
	}
	return cleanUp
}

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// TestStartRecordingFluke tests the gRPC StartRecording API endpoint. Expects an error
func TestStartRecordingFluke(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	_, err = client.StartRecording(ctx, &fmtrpc.RecordRequest{
		PollingInterval: polling_interval,
		Type: fmtrpc.RecordService_TELEMETRY,
	})
	if err == nil {
		t.Fatalf("Expected error and none were raised")
	}
}

// TestStopRecordingFluke tests the gRPC StopRecording API endpoint. Expects an error
func TestStopRecordingFluke(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	_, err = client.StopRecording(ctx, &fmtrpc.StopRecRequest{Type: fmtrpc.RecordService_TELEMETRY})
	if err == nil {
		t.Fatalf("Expected error and none were raised")
	}
}

// TestStartRecordingRga tests the gRPC StartRecording API endpoint. Expects an error
func TestStartRecordingRga(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	_, err = client.StartRecording(ctx, &fmtrpc.RecordRequest{
		PollingInterval: polling_interval,
		Type: fmtrpc.RecordService_RGA,
	})
	if err == nil {
		t.Fatalf("Expected error and none were raised")
	}
}

// TestStopRecordingRga tests the gRPC StopRecording API endpoint. Expects an error
func TestStopRecordingRga(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	_, err = client.StopRecording(ctx, &fmtrpc.StopRecRequest{Type: fmtrpc.RecordService_RGA})
	if err == nil {
		t.Fatalf("Expected error and none were raised")
	}
}

// TestSubscribeDataStream tests if we can susccessfully subscribe to the data stream
func TestSubscribeDataStream(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	stream, err := client.SubscribeDataStream(ctx, &fmtrpc.SubscribeDataRequest{})
	if err != nil {
		t.Fatalf("Could not subscribe to data stream: %v", err)
	}
	err = stream.CloseSend()
	if err != nil {
		t.Fatalf("Unexpected error when closing stream: %v", err)
	}
}

// TestDownloadHistoricalDataFluke will successfully request TCP connection info from the server but fail when receiving from TCP server
func TestDownloadHistoricalDataFluke(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	resp, err := client.DownloadHistoricalData(ctx, &fmtrpc.HistoricalDataRequest{Source: fmtrpc.RecordService_TELEMETRY})
	if err != nil {
		t.Fatalf("Unexpected error when requesting historical data: %v", err)
	}
	if resp.ServerPort != defaultTestingTCPPort {
		t.Fatalf("Unexpected TCP Port: expected %v actual %v", defaultTestingTCPPort, resp.ServerPort)
	}
	serverTCPAddr := fmt.Sprintf("localhost:%d", resp.ServerPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", serverTCPAddr)
	if err != nil {
		t.Fatalf("ResolveTCPAddr failed: %v", err)
	}
	tcpconn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.Fatalf("Unexpected error when conntecting to TCP server: %v", err)
	}
	defer tcpconn.Close()
	_, err = tcpconn.Write([]byte("Hello"))
    if err != nil {
        t.Fatalf("Unexpected error when writing to TCP server: %v", err)
    }
	_, err = tcpconn.Read(make([]byte, resp.BufferSize))
    if err == nil {
		t.Fatalf("Expected error when reading from TCP server")
	}
}

// TestDownloadHistoricalDataRga will successfully request TCP connection info from the server but fail when receiving from TCP server
func TestDownloadHistoricalDataRga(t *testing.T) {
	cleanUp := Init(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewDataCollectorClient(conn)
	resp, err := client.DownloadHistoricalData(ctx, &fmtrpc.HistoricalDataRequest{Source: fmtrpc.RecordService_RGA})
	if err != nil {
		t.Fatalf("Unexpected error when requesting historical data: %v", err)
	}
	if resp.ServerPort != defaultTestingTCPPort {
		t.Fatalf("Unexpected TCP Port: expected %v actual %v", defaultTestingTCPPort, resp.ServerPort)
	}
	serverTCPAddr := fmt.Sprintf("localhost:%d", resp.ServerPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", serverTCPAddr)
	if err != nil {
		t.Fatalf("ResolveTCPAddr failed: %v", err)
	}
	tcpconn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.Fatalf("Unexpected error when conntecting to TCP server: %v", err)
	}
	defer tcpconn.Close()
	_, err = tcpconn.Write([]byte("Hello"))
    if err != nil {
        t.Fatalf("Unexpected error when writing to TCP server: %v", err)
    }
	_, err = tcpconn.Read(make([]byte, resp.BufferSize))
    if err == nil {
		t.Fatalf("Expected error when reading from TCP server")
	}
}