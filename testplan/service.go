package testplan

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/controller"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/telemetry"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/grpc"
)

var (
	testPlanExecName = "TPEX"
	rptLvlMap = map[fmtrpc.ReportLvl]string {
		fmtrpc.ReportLvl_INFO: "INFO",
		fmtrpc.ReportLvl_ERROR: "ERROR",
		fmtrpc.ReportLvl_DEBUG: "DEBUG",
		fmtrpc.ReportLvl_WARN: "WARN",
		fmtrpc.ReportLvl_FATAL: "FATAL",
	}
	minChamberPressure float64 = 0.00005
	defaultAuthor = "FMT"
	defaultGrpcAddr = "localhost"
	ErrNoTestPlanExecuting = fmt.Errorf("No test plan currently being executed")
	ErrTPEXNotStarted = fmt.Errorf("Test Plan Executor not started")
	ErrNoTestPlanLoaded = fmt.Errorf("No test plan has been loaded")
	ErrTestPlanAlreadyStarted = fmt.Errorf("Execution already started")
	ErrTestPlanAlreadyStopped = fmt.Errorf("Execution already stopped")
)

type goroutineSafeCSVWriter struct {
	*csv.Writer
	Mutex sync.RWMutex
}

type TestPlanService struct {
	fmtrpc.UnimplementedTestPlanExecutorServer
	Running				int32
	Executing			int32
	Logger				*zerolog.Logger
	name				string
	testPlan			*TestPlan
	CancelChan			chan struct{}
	testPlanCleanup		func()
	stateWriter 		*json.Encoder
	reportWriter 		goroutineSafeCSVWriter
	collectorClient		func() (fmtrpc.DataCollectorClient, func(), error)
	stateStore			*state.Store
	wg					sync.WaitGroup
}

// Compile time check to ensure TestPlanService implements api.RestProxyService
var _ api.RestProxyService = (*TestPlanService)(nil)

// writeHeaderToReport writes a header including the date, name of the test and version number for the FMTD
func writeHeaderToReport(report goroutineSafeCSVWriter, testName string) error {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	if err := report.Write([]string{"Date", "Test Name", "FMTD Version"}); err != nil {
		return err
	}
	ct := time.Now()
	timestamp := fmt.Sprintf("%d-%d-%d", ct.Day(), ct.Month(), ct.Year())
	return report.Write([]string{timestamp, testName, utils.AppVersion})
}

// writeMsgToReport writes a timestamped message to the report
func writeMsgToReport(report goroutineSafeCSVWriter, lvl fmtrpc.ReportLvl, msg, author string, ts time.Time) error {
	report.Mutex.Lock()
	defer report.Mutex.Unlock()
	timestamp := fmt.Sprintf("%d-%d-%d %2d:%2d:%d", ts.Day(), ts.Month(), ts.Year(), ts.Hour(), ts.Minute(), ts.Second())
	lvlStr := fmt.Sprintf("%s", lvl)
	return report.Write([]string{timestamp, lvlStr, msg})
}

// NewTestPlanService instantiates the TestPlanService struct
func NewTestPlanService(logger *zerolog.Logger, getConnection func() (fmtrpc.DataCollectorClient, func(), error), store *state.Store) *TestPlanService {
	return &TestPlanService{
		Logger: logger,
		name: testPlanExecName,
		CancelChan: make(chan struct{}),
		collectorClient: getConnection,
		stateStore: store,
	}
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (s *TestPlanService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterTestPlanExecutorServer(grpcServer, s)
	return nil
}

// RegisterWithRestProxy registers the TestPlanService with the REST proxy
func(s *TestPlanService) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := fmtrpc.RegisterTestPlanExecutorHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
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
	if atomic.LoadInt32(&s.Executing) == 1 {
		_ = s.stopTestPlan()
	}
	s.Logger.Info().Msg("Test Plan Executor stopped.")
	return nil
}

// Name satisfies the data.Service interface
func (s *TestPlanService) Name() string {
	return s.name
}

// LoadTestPlan is called by gRPC client and CLI to load a given testplan
func (s *TestPlanService) LoadTestPlan(ctx context.Context, req *fmtrpc.LoadTestPlanRequest) (*fmtrpc.LoadTestPlanResponse, error) {
	if atomic.LoadInt32(&s.Running) != 1 {
		return nil, ErrTPEXNotStarted
	}
	s.Logger.Info().Msg(fmt.Sprintf("Loading test plan file at: %s", req.PathToFile))
	tp, err := loadTestPlan(req.PathToFile)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Could not load test plan file: %v", err))
		return &fmtrpc.LoadTestPlanResponse{
			Msg: fmt.Sprintf("Could not load test plan: %v", err),
		}, err
	}
	s.Logger.Info().Msg("Test plan successfully loaded.")
	s.testPlan = tp
	return &fmtrpc.LoadTestPlanResponse{
		Msg: "Test plan successfully loaded.",
	}, nil
}

// startTestPlan will prepare the execution of the test plan and enter the executeTestPlan method in a goroutine
func (s *TestPlanService) startTestPlan() error {
	if s.testPlan == nil {
		return ErrNoTestPlanLoaded
	}
	if ok := atomic.CompareAndSwapInt32(&s.Executing, 0, 1); !ok {
		return ErrTestPlanAlreadyStarted
	}
	// reset waitgroup
	var wg sync.WaitGroup
	s.wg = wg
	// make a new cancel chan
	s.CancelChan = make(chan struct{})
	//Open state file and json encoder
	state_file, err := os.OpenFile(s.testPlan.TestStateFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // write state to state file periodically for fault tolerance
	if err != nil {
		return fmt.Errorf("Could not open state file: %v", err)
	}
	stateWriter := json.NewEncoder(state_file)
	// Open report file and csv encoder
	report_file, err := os.OpenFile(s.testPlan.TestReportFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // write key events to report file
	if err != nil {
		return fmt.Errorf("Could not open report file: %v", err)
	}
	reportWriter := goroutineSafeCSVWriter{
		Writer: csv.NewWriter(report_file),
	}
	err = writeHeaderToReport(reportWriter, s.testPlan.Name)
	if err != nil {
		return fmt.Errorf("Cannot write to report: %v", err)
	}
	s.stateWriter = stateWriter
	s.reportWriter = reportWriter
	// Setup cleanup function
	s.testPlanCleanup = func() {
		_ = state_file.Close()
		s.reportWriter.Flush()
		_ = report_file.Close()
		s.testPlan = nil
	}
	s.wg.Add(1)
	go s.executeTestPlan()
	return nil
}

// executeTestPlan enables the server->client stream of realtime data and passes it along to a handler
func (s *TestPlanService) executeTestPlan() {
	defer s.wg.Done()
	// we want to call this function anytime we exit this function that way it appears first before the "Test stopped" statement
	stoppingTest := func() {
		s.Logger.Info().Msg("Stopping test...")
		err := writeMsgToReport(
			s.reportWriter, 
			fmtrpc.ReportLvl_INFO, 
			"Stopping test...", 
			defaultAuthor,
			time.Now(),
		)
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
	}
	defer func() {
		s.Logger.Info().Msg("Test stopped")
		err := writeMsgToReport(
			s.reportWriter, 
			fmtrpc.ReportLvl_INFO, 
			"Test stopped", 
			defaultAuthor,
			time.Now(),
		)
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
	}()
	s.Logger.Info().Msg("Starting test...")
	err := writeMsgToReport(
		s.reportWriter, 
		fmtrpc.ReportLvl_INFO, 
		"Starting test...", 
		defaultAuthor,
		time.Now(),
	)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		stoppingTest()
		return
	}
	ctx := context.Background()
	// First we need to start recording with the telemetry
	client, clientCleanup, err := s.collectorClient()
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot connect to Data Collector service: %v", err))
		err := writeMsgToReport(
			s.reportWriter,
			fmtrpc.ReportLvl_ERROR,
			fmt.Sprintf("Cannot connect to Data Collector service: %v", err),
			defaultAuthor,
			time.Now(),
		)
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
		stoppingTest()
		return
	}
	telemetryRecordReq := &fmtrpc.RecordRequest{
		PollingInterval: int64(5),
		Type: fmtrpc.RecordService_TELEMETRY,
	}
	_, err = client.StartRecording(ctx, telemetryRecordReq)
	if err != nil && err != telemetry.ErrAlreadyRecording {
		s.Logger.Fatal().Msg(fmt.Sprintf("Cannot start telemetry data recording: %v", err))
		err := writeMsgToReport(
			s.reportWriter,
			fmtrpc.ReportLvl_FATAL,
			fmt.Sprintf("Cannot start telemetry data recording: %v", err),
			defaultAuthor,
			time.Now(),
		)
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
		stoppingTest()
		return
	}
	clientCleanup()
	err = writeMsgToReport(
		s.reportWriter,
		fmtrpc.ReportLvl_INFO,
		"Started telemetry Data Recording",
		defaultAuthor,
		time.Now(),
	)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	// subscribe to state change updates
	stateChangeChan, unsub := s.stateStore.Subscribe(s.name)
	defer func() {
		unsub(s.stateStore, s.name)
	}()
	ticker := time.NewTicker(time.Second) // going to check things every second
	defer ticker.Stop()
	type NewLine struct{
		RptLvl	fmtrpc.ReportLvl
		Text	string
		Author	string
		Time	time.Time
	}
	var (
		rtd 		 *fmtrpc.RealTimeData
		rgaRecording bool
		readyAlerts  []*Alert
		ticks 		 int
		wgTwo		 sync.WaitGroup
	)
	for {
		newLines := make([]NewLine, 0)
		select {
		case <-ticker.C:
			// First perform any ready Alerts 
			for _, alert := range readyAlerts {
				s.Logger.Info().Msg(fmt.Sprintf("Executing %s: %s(%v) alert...", alert.Name, alert.ActionName, alert.ActionArg))
				err = writeMsgToReport(
					s.reportWriter,
					fmtrpc.ReportLvl_INFO,
					fmt.Sprintf("Executing %s: %s(%v) alert...", alert.Name, alert.ActionName, alert.ActionArg),
					defaultAuthor,
					time.Now(),
				)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
				alert.ActionFunc(alert.ActionArg)
				s.Logger.Info().Msg(fmt.Sprintf("%s alert executed.", alert.Name))
				err = writeMsgToReport(
					s.reportWriter,
					fmtrpc.ReportLvl_INFO,
					fmt.Sprintf("%s alert executed", alert.Name),
					defaultAuthor,
					time.Now(),
				)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
				alert.ExecutionState = ALERTSTATE_COMPLETED
			}
			readyAlerts = nil
			// Second let's check if there are any alerts we need to process in the next tick
			for _, alert := range s.testPlan.Alerts {
				if ticks == alert.ActionStartTime-1 && alert.ExecutionState == ALERTSTATE_PENDING {
					s.Logger.Info().Msg(fmt.Sprintf("Preparing %s: %s(%v) alert...", alert.Name, alert.ActionName, alert.ActionArg))
					readyAlerts = append(readyAlerts, alert)
				} else if ticks > alert.ActionStartTime && alert.ExecutionState == ALERTSTATE_PENDING {
					s.Logger.Info().Msg(fmt.Sprintf("%s alert expired.", alert.Name))
					newLines = append(newLines, NewLine{
						RptLvl: fmtrpc.ReportLvl_ERROR,
						Text: fmt.Sprintf("%s alert expired", alert.Name),
						Author: defaultAuthor,
						Time: time.Now(),
					})
					alert.ExecutionState = ALERTSTATE_EXPIRED
				}
			}
			// Next if rga recording hasn't started and pressure is below 0.00005 Torr, then start recording
			// TODO:SSSOCPaulCote - make this generic. Iteratre through service dependencies and start based on provided conditions
			if !rgaRecording && rtd != nil {
				if rtd.Data[drivers.TelemetryPressureChannel].Value != 0 && rtd.Data[drivers.TelemetryPressureChannel].Value < minChamberPressure {
					client, clientCleanup, err = s.collectorClient()
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot connect to Data Collector service: %v", err))
						newLines = append(newLines, NewLine{
							RptLvl: fmtrpc.ReportLvl_ERROR,
							Text: fmt.Sprintf("Cannot connect to Data Collector service: %v", err),
							Author: defaultAuthor,
							Time: time.Now(),
						})
						// call write function
						continue
					}
					rgaRecordReq := &fmtrpc.RecordRequest{
						PollingInterval: int64(15),
						Type: fmtrpc.RecordService_RGA,
					}
					_, err = client.StartRecording(ctx, rgaRecordReq)
					if err != nil && err != telemetry.ErrAlreadyRecording {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot start RGA data recording: %v", err))
						newLines = append(newLines, NewLine{
							RptLvl: fmtrpc.ReportLvl_ERROR,
							Text: fmt.Sprintf("Cannot start RGA data recording: %v", err),
							Author: defaultAuthor,
							Time: time.Now(),
						})
						// call write function
						clientCleanup()
						continue
					}
					clientCleanup()
					newLines = append(newLines, NewLine{
						RptLvl: fmtrpc.ReportLvl_INFO,
						Text: "RGA data recording started",
						Author: defaultAuthor,
						Time: time.Now(),
					})
					rgaRecording = true
				}
			}
			ticks++
		case <-stateChangeChan:
			currentRtd := s.stateStore.GetState()
			cRtd, ok := currentRtd.(controller.ControllerInitialState)
			if !ok {
				s.Logger.Error().Msg(fmt.Sprintf("Invalid type, expected fmtrpc.RealtimeData, received %v", reflect.TypeOf(currentRtd)))
			}
			rtd = &cRtd.RealTimeData
		case <- s.CancelChan:
			stoppingTest()
			client, clientCleanup, err = s.collectorClient()
			if err != nil {
				s.Logger.Error().Msg(fmt.Sprintf("Cannot connect to Data Collector service: %v", err))
				err = writeMsgToReport(
					s.reportWriter, 
					fmtrpc.ReportLvl_ERROR, 
					fmt.Sprintf("Cannot connect to Data Collector service: %v", err),
					defaultAuthor,
					time.Now(),
				)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
			}
			// stop rga recording if recording
			if rgaRecording {
				stopRgaReq := &fmtrpc.StopRecRequest{
					Type: fmtrpc.RecordService_RGA,
				}
				_, err := client.StopRecording(ctx, stopRgaReq)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot stop RGA data recording: %v", err))
					err = writeMsgToReport(
						s.reportWriter, 
						fmtrpc.ReportLvl_ERROR, 
						fmt.Sprintf("Cannot stop RGA data recording: %v", err),
						defaultAuthor,
						time.Now(),
					)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
				} else {
					rgaRecording = false
				}
			}
			stoptelemetryReq := &fmtrpc.StopRecRequest{
				Type: fmtrpc.RecordService_TELEMETRY,
			}
			_, err := client.StopRecording(ctx, stoptelemetryReq)
			if err != nil {
				s.Logger.Error().Msg(fmt.Sprintf("Cannot stop telemetry data recording: %v", err))
				err = writeMsgToReport(
					s.reportWriter, 
					fmtrpc.ReportLvl_ERROR, 
					fmt.Sprintf("Cannot stop telemetry data recording: %v", err),
					defaultAuthor,
					time.Now(),
				)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
			}
			clientCleanup()
			wgTwo.Wait()
			return
		}
		// write to csv
		for _, line := range newLines {
			wgTwo.Add(1)
			go func() {
				defer wgTwo.Done()
				err := writeMsgToReport(
					s.reportWriter,
					line.RptLvl,
					line.Text,
					line.Author,
					line.Time,
				)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
			}()
		}
	}
}

// StartTestPlan is called by gRPC client and CLI to begin the execution of the loaded test plan
func (s *TestPlanService) StartTestPlan(ctx context.Context, req *fmtrpc.StartTestPlanRequest) (*fmtrpc.StartTestPlanResponse, error) {
	s.Logger.Info().Msg("Starting execution of a test plan")
	err := s.startTestPlan()
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Could not start execution of test plan: %v", err))
		return nil, err
	}
	s.Logger.Info().Msg("Test plan execution successfully started.")
	return &fmtrpc.StartTestPlanResponse{
		Msg: fmt.Sprintf("Execution of %s test plan successfully started.", s.testPlan.Name),
	}, nil
}

// stopTestPlan will end the execution of the test plan
func (s *TestPlanService) stopTestPlan() error {
	if ok := atomic.CompareAndSwapInt32(&s.Executing, 1, 0); !ok {
		s.Logger.Error().Msg("Could not stop test plan execution: execution already stopped.")
		err := writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_ERROR, "Could not stop test plan execution: execution already stopped.", defaultAuthor, time.Now())
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
		return ErrTestPlanAlreadyStopped
	}
	close(s.CancelChan)
	s.wg.Wait()
	s.testPlanCleanup()
	return nil
}

// StopTestPlan is called by gRPC client and CLI to end the execution of the loaded test plan
func (s *TestPlanService) StopTestPlan(ctx context.Context, req *fmtrpc.StopTestPlanRequest) (*fmtrpc.StopTestPlanResponse, error) {
	err := s.stopTestPlan()
	if err != nil {
		return nil, err
	}
	return &fmtrpc.StopTestPlanResponse{
		Msg: "Execution of test plan successfully stopped",
	}, nil
}

// InsertROIMarker creates a new entry in the rest report
func (s *TestPlanService) InsertROIMarker(ctx context.Context, req *fmtrpc.InsertROIRequest) (*fmtrpc.InsertROIResponse, error) {
	if atomic.LoadInt32(&s.Executing) != 1 {
		return nil, ErrNoTestPlanExecuting
	}
	err := writeMsgToReport(s.reportWriter, req.ReportLvl, req.Text, req.Author, time.Now())
	if err != nil {
		return nil, fmt.Errorf("Cannot write to report: %v", err)
	}
	return &fmtrpc.InsertROIResponse{
		Msg: "Region of Interest marker inserted successfully.",
	}, nil
}