/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package data

import (
	"context"
	"fmt"
	"os"
	"net"
	"testing"

	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOCPaulCote/gux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	testReducer gux.Reducer = func(s interface{}, a gux.Action) (interface{}, error) {
		// assert type of s
		_, ok := s.(fmtrpc.RealTimeData)
		if !ok {
			return nil, errors.ErrInvalidType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, errors.ErrInvalidType
			}
			return newState, nil
		default:
			return nil, errors.ErrInvalidAction
		} 
	}
	bufSize = 1 * 1024 * 1024
	lis *bufconn.Listener
	polling_interval int64
)

// TestRTDServiceStartStop tests if we can initialize, start and stop the RTD service
func TestRTDServiceStartStop(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	stateStore := gux.CreateStore(fmtrpc.RealTimeData{}, testReducer)
	rtdService := NewRTDService(
		&logger,
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
	stateStore := gux.CreateStore(fmtrpc.RealTimeData{}, testReducer)
	rtdService := NewRTDService(&logger, stateStore)
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