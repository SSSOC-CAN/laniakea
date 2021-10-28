package testplan

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"
	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/grpc"
)

var (
	testPlanExecName = "TPEX"
	rpt_INFO = "INFO"
	rpt_ERR = "ERROR"
	rpt_DEBUG = "DEBUG"
	rpt_WARN = "WARN"
	rpt_FATAL = "FATAL"
	minChamberPressure float64 = 0.00005
)

type TestPlanService struct {
	fmtrpc.UnimplementedTestPlanExecutorServer
	Running				int32
	Executing			int32
	Logger				*zerolog.Logger
	clientConn			fmtrpc.TestPlanExecutorClient
	clientConnCleanup	func()
	name				string
	testPlan			*TestPlan
	CancelChan			chan struct{}
	stateFile 			*json.Encoder
	reportFile 			*csv.Writer
}

type ReportLVL string

//getDataCollectorClient returns the DataCollectorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getDataCollectorClient(macaroon_timeout int64) (fmtrpc.DataCollectorClient, func(), error) {
	conn, err := auth.GetClientConn(false, macaroon_timeout)
	if err != nil {
		return nil, nil, err
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewDataCollectorClient(conn), cleanUp, nil
}

// writeHeaderToReport writes a header including the date, name of the test and version number for the FMTD
func writeHeaderToReport(report *csv.Writer,testName string) error {
	if err := report.Write([]string{"Date", "Test Name", "FMTD Version"}); err != nil {
		return err
	}
	ct := time.Now()
	timestamp := fmt.Sprintf("%2d-%2d-%d", ct.Day(), ct.Month(), ct.Year())
	return report.Write([]string{timestamp, testName, utils.AppVersion})
}

// writeMsgToReport writes a timestamped message to the report
func writeMsgToReport(report *csv.Writer, lvl ReportLVL, msg string) error {
	ct := time.Now()
	timestamp := fmt.Sprintf("%2d:%2d:%2d", ct.Hour(), ct.Minute(), ct.Second())
	lvlStr := fmt.Sprintf("%s", lvl)
	return report.Write([]string{timestamp, lvlStr, msg})
}

// NewTestPlanService instantiates the TestPlanService struct
func NewTestPlanService(logger *zerolog.Logger) *TestPlanService {
	return &TestPlanService{
		Logger: logger,
		name: testPlanExecName,
		CancelChan: make(chan struct{})
	}
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (s *TestPlanService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterDataCollectorServer(grpcServer, s)
	return nil
}

// Start starts the Test Plan service
func (s *TestPlanService) Start() error {
	s.Logger.Info().Msg("Starting Test Plan Executor...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start Test Plan Executor. Service already started.")
	}
	s.Logger.Info().Msg("Test Plan Executor successfully started.")
	return nil
}

// Stop stops the RTD service. It closes all its channels, lifting the burdens from Data providers
func (s *TestPlanService) Stop() error {
	s.Logger.Info().Msg("Stopping Test Plan Executor ...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop Test Plan Executor. Service already stopped.")
	}
	s.clientConnCleanup()
	s.Logger.Info().Msg("Test Plan Executor stopped.")
	return nil
}

// LoadTestPlan is called by gRPC client and CLI to load a given testplan
func (s *TestPlanService) LoadTestPlan(ctx context.Context, req *fmtrpc.LoadTestPlanRequest) (*fmtrpc.LoadTestPlanResponse, error) {
	if atomic.LoadInt32(&s.Running) != 1 {
		return nil, fmt.Errorf("Cannot load test plan: Test Plan Executor not started.")
	}
	tp, err := loadTestPlan(req.PathToFile)
	if err != nil {
		return &fmtrpc.LoadTestPlanResponse{
			Msg: fmt.Sprintf("Could not load test plan: %v", err)
		}, err
	}
	s.testPlan = tp
	return &fmtrpc.LoadTestPlanResponse{
		Msg: "Test plan successfully loaded."
	}, nil
}

// startTestPlan will prepare the execution of the test plan and enter the executeTestPlan method in a goroutine
func (s *TestPlanService) startTestPlan() error {
	if s.testPlan == nil {
		return fmt.Errorf("Could not start test plan execution: no test plan has been loaded.")
	}
	if ok := atomic.CompareAndSwapInt32(&s.Executing, 0, 1); !ok {
		return fmt.Errorf("Could not start test plan execution: execution already started.")
	}
	state_file, err := os.OpenFile(t.testPlan.TestStateFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775) // write state to state file periodically for fault tolerance
	if err != nil {
		return fmt.Errorf("Could not open state file: %v", err)
	}
	stateWriter := json.NewEncoder(state_file)
	report_file, err := os.OpenFile(t.testPlan.TestReportFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775) // write key events to report file
	if err != nil {
		return fmt.Errorf("Could not open report file: %v", err)
	}
	reportWriter := csv.NewWriter(report_file)
	err = writeHeaderToReport(reportWriter, s.testPlan.Name)
	if err != nil {
		return fmt.Errorf("Cannot write to report: %v", err)
	}
	s.stateWriter = stateWriter
	s.reportWriter = reportWriter
	dataCollectorClient, cleanUp, err := getDataCollectorClient(s.testPlan.TestDuration)
	if err != nil {
		return fmt.Errorf("Could not get Data Collector Client connection: %v", err)
	}
	s.clientConn, s.clientConnCleanup = dataCollectorClient, cleanUp // now we have our client connection
	s.Logger.Info().Msg(fmt.Sprintf("%s test is beginning...", s.testPlan.Name))
	go s.executeTestPlan()
	return nil
}

// executeTestPlan enables the server->client stream of realtime data and passes it along to a handler
func (s *TestPlanService) executeTestPlan() {
	startTestTime := time.Now()
	err := writeMsgToReport(s.reportFile, rpt_INFO, "Starting test...")
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	ctx := context.Background()
	// First we need to start recording with the fluke
	flukeRecordReq := &fmtrpc.RecordRequest{
		PollingInterval: int64(10),
		Type: fmtrpc.RecordService_FLUKE,
	}
	recordResponse, err := client.StartRecording(ctx, recordRequest)
	if err != nil {
		s.Logger.Fatal().Msg(fmt.Sprintf("Cannot start fluke data recording: %v", err))
	}
	err := writeMsgToReport(s.reportFile, rpt_INFO, "Started Fluke Data Recording")
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	dataChan := make(chan *fmtrpc.RealTimeData)
	go func () {
		dataStreamReq := &fmtrpc.SubscribeDataRequest{}
		stream, err := s.clientConn.SubscribeDataStream(ctx, dataStreamReq)
		if err != nil {
			s.Logger.Fatal().Msg(fmt.Sprintf("Cannot subscribe to data stream: %v", err))
		}
		err := writeMsgToReport(s.reportFile, rpt_INFO, "Started listening for real time data")
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
		for {
			resp, err := stream.Recv()
			if err != nil {
				s.Logger.Error().Msg(fmt.Sprintf("Data stream interrupted: %v", err))
				err := writeMsgToReport(s.reportFile, rpt_INFO, fmt.Sprintf("Real time data stream interrupted: %v", err))
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
				return
			}
			dataChan <- resp
		}
	}()
	ticker := time.NewTicker(time.Second) // going to check things every second
	var rtd *fmtrpc.RealTimeData
	var rgaRecording bool
	for {
		select {
		case <-ticker.C:
			// First let's check that we aren't at the end of our test
			if !time.Now().Before(startTestTime.Add(time.Duration(s.testPlan.TestDuration)*time.Hour)) {
				err := writeMsgToReport(s.reportFile, rpt_INFO, "Stopping test...")
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
				_ = s.stopTestPlan()
				err := writeMsgToReport(s.reportFile, rpt_INFO, "Test stopped")
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
			}
			// Second let's check if there are any alerts we need to process
			for _, alert := s.testPlan.Alerts {
				if time.Now().Equal(startTestTime.Add(time.Duration(alert.ActionStartTime)*time.Second)) {
					err := writeMsgToReport(s.reportFile, rpt_INFO, fmt.Sprintf("Executing %s: %s(%v) alert...", alert.Name, alert.ActionName, alert.ActionArg))
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
					alert.ActionFunc(alert.ActionArg)
					err := writeMsgToReport(s.reportFile, rpt_INFO, fmt.Sprintf("%s alert executed", alert.ActionName))
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
					alert.ExecutionState = ALERTSTATE_COMPLETED
				} else if time.Now().After(startTestTime.Add(time.Duration(alert.ActionStartTime)*time.Second)) && alert.ExecutionState == ALERTSTATE_PENDING {
					err := writeMsgToReport(s.reportFile, rpt_ERR, fmt.Sprintf("%s alert expired", alert.ActionName))
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
					alert.ExecutionState = ALERTSTATE_EXPIRED
				}
			}
			// Next if rga recording hasn't started and pressure is below 0.00005 Torr, then start recording
			// TODO:SSSOCPaulCote - make this generic. Iteratre through service dependencies and start based on provided conditions
			if !rgaRecording {
				if rtd.Data[122].Value != 0 && rtd.Data[122].Value < minChamberPressure {
					rgaRecordReq := &fmtrpc.RecordRequest{
						PollingInterval: int64(3),
						Type: fmtrpc.RecordService_RGA,
					}
					recordResponse, err := client.StartRecording(ctx, recordRequest)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot start fluke data recording: %v", err))
					}
					err := writeMsgToReport(s.reportFile, rpt_INFO, "RGA data recording started")
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
					rgaRecording = true
				}
			}
		case rtd = <-dataChan:
		case <- s.QuitChan:
			s.clientConnCleanup()
			break
		}
	}
}

// StartTestPlan is called by gRPC client and CLI to begin the execution of the loaded test plan
func (s *TestPlanService) StartTestPlan(ctx context.Context, req *fmtrpc.StartTestPlanRequest) (*fmtrpc.StartTestPlanResponse, error) {
	err := s.startTestPlan()
	if err != nil {
		return nil, fmt.Errorf("Could not start executing test plan: %v", err)
	}
	return &fmtrpc.StartTestPlanResponse{
		Msg: fmt.Sprintf("Execution of %s test plan successfully started.", s.testPlan.Name)
	}, nil
}

// stopTestPlan will end the execution of the test plan
func (s *TestPlanService) stopTestPlan() error {
	if ok := atomic.CompareAndSwapInt32(&s.Executing, 1, 0); !ok {
		return fmt.Errorf("Could not stop test plan execution: execution already stopped.")
	}
	close(s.QuitChan)
	return nil
}

// StopTestPlan is called by gRPC client and CLI to end the execution of the loaded test plan
func (s *TestPlanService) StopTestPlan(ctx context.Context, req *fmtrpc.StopTestPlanRequest) (*fmtrpc.StopTestPlanResponse, error) {
	err := s.stopTestPlan()
	if err != nil {
		return nil, fmt.Errorf("Could not stop executing test plan: %v", err)
	}
	return &fmtrpc.StopTestPlanResponse{
		Msg: fmt.Sprintf("Execution of %s test plan successfully stopped.", s.testPlan.Name)
	}, nil
}

// InsertROIMarker creates a new entry in the rest report
func (s *TestPlanService) InsertROIMarker(ctx context.Context, req *fmtrpc.InsertROIRequest) (*fmtrpc.InsertROIResponse, error) {
	err := writeMsgToReport(s.reportFile, rpt_INFO, req.Text)
	if err != nil {
		return nil, fmt.Errorf("Cannot write to report: %v", err)
	}
	return &fmtrpc.InsertROIResponse{
		Msg: "Region of Interest marker inserted successfully.",
	}, nil
}