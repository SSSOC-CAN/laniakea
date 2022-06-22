/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

// TODO:SSSOCPaulCote - Get gRPC to return actual errors

package testplan

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/controller"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/telemetry"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/SSSOCPaulCote/gux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	bufSize = 1 * 1024 * 1024
	lis *bufconn.Listener
	defaultTestGrpcPort int64 = 5678 
	defaultTestingTCPAddr = "localhost"
	defaultTestingTCPPort int64 = 10024
	testDataCollectorClient = func() (fmtrpc.DataCollectorClient, func(), error) {
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to dial bufnet: %v", err)
		}
		cleanUp := func() {
			conn.Close()
		}
		return fmtrpc.NewDataCollectorClient(conn), cleanUp, nil
	}
	testPlanFileLines = []string{
		"plan_name: \"Test\"\n",
		"test_duration: 120\n", // 2 minute test
		"data_providers:\n",
		"  - provider_name: \"Fluke\"\n",
		"    driver: \"Fluke DAQ\"\n",
		"    num_data_points: 136\n",
		"  - provider_name: \"RGA\"\n",
		"    driver: \"RGA\"\n",
		"    num_data_points: 200\n",
		"alerts:\n",
		"  - alert_name: \"Wait 60 seconds\"\n",
		"    action: \"WaitForTime\"\n",
		"    action_arg: 60\n",
		"    action_start_time: 30\n",
		//"report_file_path: \"C:\\\\Users\\\\Michael Graham\\\\Downloads\\\\testplan_test.csv\"\n",
	}
	rtdInitialState = data.InitialRtdState{}
	rtdReducer gux.Reducer = func(s interface{}, a gux.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialRtdState)
		if !ok {
			return nil, gux.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(data.InitialRtdState)
			if !ok {
				return nil, gux.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState.RealTimeData
			oldState.AverageTemperature = newState.AverageTemperature
			return oldState, nil
		case "rga/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, gux.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState
			return oldState, nil
		case "telemetry/polling_interval/update":
			// assert type of payload
			newPol, ok := a.Payload.(int64)
			if !ok {
				return nil, gux.ErrInvalidPayloadType
			}
			oldState.TelPollingInterval = newPol
			return oldState, nil
		default:
			return nil, gux.ErrInvalidAction
		} 
	}
)

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestTestplan(t *testing.T) {
	t.Skip("Skipping test as changes to Telemetry service are needed for local writing")
	// make temporary directory
	tempDir, err := ioutil.TempDir("", "testplan-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	var testPlanPath string
	if runtime.GOOS == "windows" {
		tp_path := fmt.Sprintf("report_file_path: \"%v\\%v\"\n", utils.AppDataDir("fmtd", false), "test_testplan.csv")
		testPlanFileLines = append(
			testPlanFileLines,
			strings.Replace(tp_path, `\`, `\\`, -1),
		)
		testPlanPath = fmt.Sprintf("%v\\%v", tempDir, "testplan.yaml")
	} else {
		testPlanFileLines = append(
			testPlanFileLines,
			fmt.Sprintf("report_file_path: \"%v\"\n", path.Join(utils.AppDataDir("fmtd", false), "test_testplan.csv")),
		)
		testPlanPath = path.Join(tempDir, "testplan.yaml")
	}
	// testplan file
	tp_file, err := os.OpenFile(testPlanPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		t.Fatalf("Could not open/create file: %v", err)
	}
	for _, line := range testPlanFileLines {
		_, err = tp_file.WriteString(line)
		if err != nil {
			t.Fatalf("Could not write to test plan file: %v", err)
		}
	}
	tp_file.Sync()
	tp_file.Close()
	// gRPC server
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	// init logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	// state stores
	store := gux.CreateStore(rtdInitialState, rtdReducer)
	ctrlStore := gux.CreateStore(controller.InitialState, controller.ControllerReducer)
	// Telemetry Service
	telemetryLogger := logger.With().Str("subsystem", "TEL").Logger()
	// Connect to DAQ
	c, err := drivers.ConnectToDAQ()
	if err != nil {
		t.Fatalf("Could not connect to telemetry DAQ: %v", err)
	}
	daqConn, ok := c.(*drivers.DAQConnection)
	if !ok {
		t.Fatalf("Could not connect to telemetry DAQ OPC: %v", errors.ErrInvalidType)
	}
	defer daqConn.Close()
	telemetryService := telemetry.NewTelemetryService(
		&telemetryLogger,
		store,
		ctrlStore,
		daqConn,
		"",
		"",
	)
	// RTD Service
	rtdLogger := logger.With().Str("subsystem", "RTD").Logger()
	rtdService := data.NewRTDService(
		&rtdLogger,
		store,
	)
	err = rtdService.RegisterWithGrpcServer(grpcServer)
	if err != nil {
		t.Fatalf("Could not register RTD service with gRPC server: %v", err)
	}
	// Register telemetry with RTD
	telemetryService.RegisterWithRTDService(rtdService)
	// Testplan Executor
	tpexLogger := logger.With().Str("subsystem", "TPEX").Logger()
	tpexService := NewTestPlanService(
		&tpexLogger,
		testDataCollectorClient,
		store,
	)
	err = tpexService.RegisterWithGrpcServer(grpcServer)
	if err != nil {
		t.Fatalf("Could not register TPEX service with gRPC server: %v", err)
	}
	// Start gRPC
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	// Executor client
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewTestPlanExecutorClient(conn)
	// Load Test Plan before start
	t.Run("fmtcli load-testplan prestart", func(t *testing.T) {
		resp, err := client.LoadTestPlan(ctx, &fmtrpc.LoadTestPlanRequest{
			PathToFile: testPlanPath,
		})
		if err == nil {
			t.Fatalf("Expected an error when calling LoadTestPlan RPC endpoint")
		}
		t.Log(resp)
	})
	// Start telemetry, Start TPEX, Start gRPC
	err = telemetryService.Start()
	if err != nil {
		t.Errorf("Could not start telemetry service: %v", err)
	}
	err = tpexService.Start()
	if err != nil {
		t.Errorf("Could not start TPEX service: %v", err)
	}
	defer func() {
		tpexService.Stop()
		rtdService.Stop()
		telemetryService.Stop()
		grpcServer.Stop()
	}()
	// Load Test Plan
	t.Run("fmtcli load-testplan", func(t *testing.T) {
		resp, err := client.LoadTestPlan(ctx, &fmtrpc.LoadTestPlanRequest{
			PathToFile: testPlanPath,
		})
		if err != nil {
			t.Fatalf("Unexpected error when calling LoadTestPlan RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Start Test Plan
	t.Run("fmtcli start-testplan", func(t *testing.T) {
		resp, err := client.StartTestPlan(ctx, &fmtrpc.StartTestPlanRequest{})
		if err != nil {
			t.Fatalf("Unexpected error when calling StartTestPlan RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// wait 1 minute and a half
	time.Sleep(92*time.Second)
	// Insert ROI
	t.Run("fmtcli insert-roi", func(t *testing.T) {
		resp, err := client.InsertROIMarker(ctx, &fmtrpc.InsertROIRequest{
			Text: "YYYEEEEEETTTT",
			ReportLvl: fmtrpc.ReportLvl_DEBUG,
			Author: "TESTER MCGEE",
		})
		if err != nil {
			t.Fatalf("Unexpected error when calling InsertROIMarker RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Stop Test Plan
	t.Run("fmtcli stop-testplan", func(t *testing.T) {
		resp, err := client.StopTestPlan(ctx, &fmtrpc.StopTestPlanRequest{})
		if err != nil {
			t.Fatalf("Unexpected error when calling StopTestPlan RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Stop Test Plan invalid
	t.Run("fmtcli stop-testplan invalid", func(t *testing.T) {
		resp, err := client.StopTestPlan(ctx, &fmtrpc.StopTestPlanRequest{})
		if err == nil {
			t.Fatalf("Expected an error when calling StopTestPlan RPC endpoint")
		}
		t.Log(resp)
	})
	// Insert ROI invalid 
	t.Run("fmtcli insert-roi invalid", func(t *testing.T) {
		resp, err := client.InsertROIMarker(ctx, &fmtrpc.InsertROIRequest{
			Text: "YYYEEEEEETTTT",
			ReportLvl: fmtrpc.ReportLvl_DEBUG,
			Author: "TESTER MCGEE",
		})
		if err == nil {
			t.Fatalf("Expected an error when calling InsertROIMarker RPC endpoint")
		}
		t.Log(resp)
	})
	// Start Test Plan and wait for it to complete
	t.Run("fmtcli restart-testplan invalid", func(t *testing.T) {
		resp, err := client.StartTestPlan(ctx, &fmtrpc.StartTestPlanRequest{})
		if err == nil {
			t.Fatalf("Expected an error when calling DStartTestPlan RPC endpoint")
		}
		t.Log(resp)
	})
	// Reload Test Plan
	t.Run("fmtcli reload-testplan", func(t *testing.T) {
		resp, err := client.LoadTestPlan(ctx, &fmtrpc.LoadTestPlanRequest{
			PathToFile: testPlanPath,
		})
		if err != nil {
			t.Fatalf("Unexpected error when calling LoadTestPlan RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Restart Test Plan
	t.Run("fmtcli restart-testplan", func(t *testing.T) {
		resp, err := client.StartTestPlan(ctx, &fmtrpc.StartTestPlanRequest{})
		if err != nil {
			t.Fatalf("Unexpected error when calling StartTestPlan RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Start Test Plan Invalid
	t.Run("fmtcli start-testplan invalid", func(t *testing.T) {
		resp, err := client.StartTestPlan(ctx, &fmtrpc.StartTestPlanRequest{})
		if err == nil {
			t.Fatalf("Expected an error when calling StartTestPlan RPC endpoint")
		}
		t.Log(resp)
	})
	time.Sleep(122*time.Second)
}