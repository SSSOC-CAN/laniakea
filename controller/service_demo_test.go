// +build demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/fmtrpc/demorpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type testingSetTemp struct {
	caseName		string
	setPoint		float64
	rate			float64
	expectedTempMin	float64
	expectedTempMax float64
}

var (
	makeInitialDataMap = func() map[int64]*fmtrpc.DataField {
		m := make(map[int64]*fmtrpc.DataField)
		for i := 0; i < 96; i++ {
			m[int64(i)] = &fmtrpc.DataField{
				Name: fmt.Sprintf("Value #%v", i+1),
				Value: (rand.Float64()*0.1)+25,
			}
		}
		return m
	}
	rtdInitialState = data.InitialRtdState{
		AverageTemperature: float64(25.0),
		TelPollingInterval: int64(5),
		RealTimeData: fmtrpc.RealTimeData{
			Source: "TEL",
			IsScanning: false,
			Timestamp: time.Now().UnixMilli(),
			Data: makeInitialDataMap(),
		},
	}
	rtdReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialRtdState)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(data.InitialRtdState)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState.RealTimeData
			oldState.AverageTemperature = newState.AverageTemperature
			return oldState, nil
		case "rga/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState
			return oldState, nil
		case "telemetry/polling_interval/update":
			// assert type of payload
			newPol, ok := a.Payload.(int64)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.TelPollingInterval = newPol
			return oldState, nil
		default:
			return nil, state.ErrInvalidAction
		} 
	}
	bufSize = 1 * 1024 * 1024
	lis *bufconn.Listener
	testSetTempCases = []testingSetTemp{
		{"temp_increase_no_rate", 30.0, float64(0), 30.0, 30.6},
		{"temp_decrease_no_rate", 20.0, float64(0), 20.0, 20.6},
		{"temp_increase_rate", 25.0, 5.0, 24.0, 26.0},
		{"temp_decrease_rate", 20.0, 5.0, 19.0, 21.0},
	}
)

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// initController initializes a controller service
func initController(t *testing.T) (*ControllerService, func(string) error, string) {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tmp_dir, err := ioutil.TempDir("", "controller_test-")
	if err != nil {
		t.Fatalf("Could not create a temporary directory: %v", err)
	}
	stateStore := state.CreateStore(rtdInitialState, rtdReducer)
	ctrlStore := state.CreateStore(InitialState, ControllerReducer)
	ctrlConn, _ := drivers.ConnectToController()
	return NewControllerService(
		&log,
		stateStore,
		ctrlStore,
		ctrlConn,
	), os.RemoveAll, tmp_dir
}

// TestStartStopControllerService tests the start/stop functions of the ControllerService
func TestStartStopControllerService(t *testing.T) {
	controllerService, cleanUp , tmpDir := initController(t)
	defer cleanUp(tmpDir)
	t.Run("Start Controller Service", func(t *testing.T) {
		err := controllerService.Start()
		if err != nil {
			t.Errorf("Could not start controller service: %v", err)
		}
	})
	t.Run("Start Controller Service Invalid", func(t *testing.T) {
		err := controllerService.Start()
		if err != ErrServiceAlreadyStarted {
			t.Errorf("Unexpected error when starting controller service: %v", err)
		}
	})
	t.Run("Stop Controller Service", func(t *testing.T) {
		err := controllerService.Stop()
		if err != nil {
			t.Errorf("Could not stop controller service: %v", err)
		}
	})
	t.Run("Stop Controller Service Invalid", func(t *testing.T) {
		err := controllerService.Stop()
		if err != ErrServiceAlreadyStopped {
			t.Errorf("Unexpected error when starting controller service: %v", err)
		}
	})
}

// TestRegisterWithRestProxy tests if the Controller Service can successfully be registered with the REST proxy
func TestRegisterWithRestProxy(t *testing.T) {
	controllerService, cleanUp, tempDir := initController(t)
	defer cleanUp(tempDir)
	// prereqs for RegisterWithRestProxy
	// context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Proxy Serve Mux
	customMarshalerOption := proxy.WithMarshalerOption(
		proxy.MIMEWildcard, &proxy.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
				EmitUnpopulated: true,
			},
		},
	)
	mux := proxy.NewServeMux(customMarshalerOption)
	// TLS REST config
	_, restDialOpts, _, tlsCleanUp, err := cert.GetTLSConfig(
		path.Join(tempDir, "tls.cert"),
		path.Join(tempDir, "tls.key"),
		make([]string, 0),
	)
	if err != nil {
		t.Errorf("Could not get TLS Config: %v", err)
	}
	defer tlsCleanUp()
	restProxyDestNet, err := utils.NormalizeAddresses([]string{fmt.Sprintf("localhost:%d", 3567)}, strconv.FormatInt(3567, 10), net.ResolveTCPAddr)
	if err != nil {
		t.Errorf("Could not normalize address: %v", err)
	}
	restProxyDest := restProxyDestNet[0].String()
	err = controllerService.RegisterWithRestProxy(ctx, mux, restDialOpts, restProxyDest)
	if err != nil {
		t.Errorf("Could not register with Rest Proxy: %v", err)
	}
}

// TestControllerAPI tests the various controller API endpoints with multiple cases
func TestControllerAPI(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	controllerService, cleanUp, tmpDir := initController(t)
	err := controllerService.Start()
	if err != nil {
		t.Errorf("Could not start controller service: %v", err)
	}
	_ = controllerService.RegisterWithGrpcServer(grpcServer)
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	defer func() {
		_ = controllerService.Stop()
		grpcServer.Stop()
		cleanUp(tmpDir)
	}()
    client := demorpc.NewControllerClient(conn)
	// goroutine mimicking telemetry service
	quitChan := make(chan struct{})
	ticker := time.NewTicker(time.Duration(int64(5)) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				var (
					cumSum float64
					cnt	   int64
				)
				dataField := make(map[int64]*fmtrpc.DataField)
				// get current pressure set point
				currentState := controllerService.ctrlStateStore.GetState()
				cState, ok := currentState.(data.InitialCtrlState)
				if !ok {
					return
				}
				for i := 0; i < 96; i++ {
					v := (rand.Float64()*0.1)+cState.TemperatureSetPoint
					cumSum += v
					cnt += 1
					dataField[int64(i)]= &fmtrpc.DataField{
						Name: fmt.Sprintf("Value #%v", i+1),
						Value: v,
					}
				}
				err := controllerService.rtdStateStore.Dispatch(
					state.Action{
						Type: 	 "telemetry/update",
						Payload: data.InitialRtdState{
							RealTimeData: fmtrpc.RealTimeData{
								Source: "TEL",
								IsScanning: false,
								Timestamp: time.Now().UnixMilli(),
								Data: dataField,
							},
							AverageTemperature: cumSum/float64(cnt),
						},
					},
				)
				if err != nil {
					return
				}
			case <-quitChan:
				return 
			}
		}
	}()
	defer func() {
		ticker.Stop()
		close(quitChan)
	}()
	// Iterate through test cases
	for _, c := range testSetTempCases {
		t.Run(c.caseName, func(t *testing.T) {
			stream, err := client.SetTemperature(ctx, &demorpc.SetTempRequest{
				TempSetPoint: c.setPoint,
				TempChangeRate: c.rate,
			})
			if err != nil {
				t.Errorf("Could not send SetTemperature command: %v", err)
			}
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Errorf("Could not receive SetTemperature response: %v", err)
				}
				t.Log(resp)
			}
			currentState := controllerService.rtdStateStore.GetState()
			cState, ok := currentState.(data.InitialRtdState)
			if !ok {
				t.Errorf("%v", state.ErrInvalidStateType)
			}
			if cState.AverageTemperature > c.expectedTempMax || cState.AverageTemperature < c.expectedTempMin {
				t.Errorf("Unexpected average temperature value after changing setpoint: %v", cState.AverageTemperature)
			}
		})
	}
}