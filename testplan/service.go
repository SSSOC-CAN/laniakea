package testplan

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
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
)

type TestPlanService struct {
	fmtrpc.UnimplementedTestPlanExecutorServer
	Running				int32
	Executing			int32
	Logger				*zerolog.Logger
	client				fmtrpc.DataCollectorClient
	clientConn			*grpc.ClientConn
	name				string
	testPlan			*TestPlan
	CancelChan			chan struct{}
	testPlanCleanup		func()
	stateWriter 		*json.Encoder
	reportWriter 		*csv.Writer
	tlsCertPath			string
	adminMacPath		string
	grpcPort			int64
	wg					sync.WaitGroup
}

//getDataCollectorClient returns the DataCollectorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getDataCollectorClient(macaroon_timeout, grpcPort int64, tlsCertPath, adminMacPath string) (fmtrpc.DataCollectorClient, *grpc.ClientConn, error) {
	conn, err := auth.GetClientConn(tlsCertPath, adminMacPath, false, macaroon_timeout, grpcPort)
	if err != nil {
		return nil, nil, err
	}
	return fmtrpc.NewDataCollectorClient(conn), conn, nil
}

// writeHeaderToReport writes a header including the date, name of the test and version number for the FMTD
func writeHeaderToReport(report *csv.Writer, testName string) error {
	if err := report.Write([]string{"Date", "Test Name", "FMTD Version"}); err != nil {
		return err
	}
	ct := time.Now()
	timestamp := fmt.Sprintf("%2d-%2d-%d", ct.Day(), ct.Month(), ct.Year())
	return report.Write([]string{timestamp, testName, utils.AppVersion})
}

// writeMsgToReport writes a timestamped message to the report
func writeMsgToReport(report *csv.Writer, lvl fmtrpc.ReportLvl, msg, author string) error {
	ct := time.Now()
	timestamp := fmt.Sprintf("%2d:%2d:%2d", ct.Hour(), ct.Minute(), ct.Second())
	lvlStr := fmt.Sprintf("%s", lvl)
	return report.Write([]string{timestamp, lvlStr, msg})
}

// NewTestPlanService instantiates the TestPlanService struct
func NewTestPlanService(logger *zerolog.Logger, tlsCertPath, adminMacPath string, grpcPort int64) *TestPlanService {
	var wg sync.WaitGroup
	return &TestPlanService{
		Logger: logger,
		name: testPlanExecName,
		CancelChan: make(chan struct{}),
		tlsCertPath: tlsCertPath,
		adminMacPath: adminMacPath,
		grpcPort: grpcPort,
		wg: wg,
	}
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (s *TestPlanService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterTestPlanExecutorServer(grpcServer, s)
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
		return nil, fmt.Errorf("Cannot load test plan: Test Plan Executor not started.")
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
		return fmt.Errorf("Could not start test plan execution: no test plan has been loaded.")
	}
	if ok := atomic.CompareAndSwapInt32(&s.Executing, 0, 1); !ok {
		return fmt.Errorf("Could not start test plan execution: execution already started.")
	}
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
	reportWriter := csv.NewWriter(report_file)
	err = writeHeaderToReport(reportWriter, s.testPlan.Name)
	if err != nil {
		return fmt.Errorf("Cannot write to report: %v", err)
	}
	s.stateWriter = stateWriter
	s.reportWriter = reportWriter
	reportWriter.Write([]string{"Test", "1", "2", "3"})
	// Get DataCollector client connection
	dataCollectorClient, clientConn, err := getDataCollectorClient(s.testPlan.TestDuration, s.grpcPort, s.tlsCertPath, s.adminMacPath)
	if err != nil {
		return fmt.Errorf("Could not get Data Collector Client connection: %v", err)
	}
	s.client, s.clientConn = dataCollectorClient, clientConn // now we have our client connection
	// Setup cleanup function
	s.testPlanCleanup = func() {
		_ = state_file.Close()
		s.reportWriter.Flush()
		_ = report_file.Close()
	}
	s.wg.Add(1)
	go s.executeTestPlan()
	return nil
}

// executeTestPlan enables the server->client stream of realtime data and passes it along to a handler
func (s *TestPlanService) executeTestPlan() {
	defer s.wg.Done()
	s.Logger.Info().Msg("Starting test...")
	err := writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, "Starting test...", defaultAuthor)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	ctx := context.Background()
	// First we need to start recording with the fluke
	flukeRecordReq := &fmtrpc.RecordRequest{
		PollingInterval: int64(10),
		Type: fmtrpc.RecordService_FLUKE,
	}
	_, err = s.client.StartRecording(ctx, flukeRecordReq)
	if err != nil {
		s.Logger.Fatal().Msg(fmt.Sprintf("Cannot start fluke data recording: %v", err))
	}
	err = writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, "Started Fluke Data Recording", defaultAuthor)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	dataChan := make(chan *fmtrpc.RealTimeData)
	defer close(dataChan)
	go func () {
		dataStreamReq := &fmtrpc.SubscribeDataRequest{}
		stream, err := s.client.SubscribeDataStream(ctx, dataStreamReq)
		if err != nil {
			s.Logger.Fatal().Msg(fmt.Sprintf("Cannot subscribe to data stream: %v", err))
		}
		err = writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, "Started listening for real time data", defaultAuthor)
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
		s.Logger.Info().Msg("Receiving realtime data.")
		for {
			resp, err := stream.Recv()
			if err != nil {
				s.Logger.Error().Msg(fmt.Sprintf("Data stream interrupted: %v", err))
				return
			}
			dataChan <- resp
		}
	}()
	ticker := time.NewTicker(time.Second) // going to check things every second
	var rtd *fmtrpc.RealTimeData
	var rgaRecording bool
	var readyAlerts []*Alert
	ticks := 0
	for {
		select {
		case <-ticker.C:
			// First perform any ready Alerts 
			for _, alert := range readyAlerts {
				s.Logger.Info().Msg(fmt.Sprintf("Executing %s: %s(%v) alert...", alert.Name, alert.ActionName, alert.ActionArg))
				err := writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, fmt.Sprintf("Executing %s: %s(%v) alert...", alert.Name, alert.ActionName, alert.ActionArg), defaultAuthor)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
				}
				alert.ActionFunc(alert.ActionArg)
				s.Logger.Info().Msg(fmt.Sprintf("%s alert executed.", alert.Name))
				err = writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, fmt.Sprintf("%s alert executed", alert.Name), defaultAuthor)
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
					err := writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_ERROR, fmt.Sprintf("%s alert expired", alert.Name), defaultAuthor)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
					alert.ExecutionState = ALERTSTATE_EXPIRED
				}
			}
			// Next if rga recording hasn't started and pressure is below 0.00005 Torr, then start recording
			// TODO:SSSOCPaulCote - make this generic. Iteratre through service dependencies and start based on provided conditions
			if !rgaRecording && rtd != nil {
				if rtd.Data[122].Value != 0 && rtd.Data[122].Value < minChamberPressure {
					rgaRecordReq := &fmtrpc.RecordRequest{
						PollingInterval: int64(3),
						Type: fmtrpc.RecordService_RGA,
					}
					_, err = s.client.StartRecording(ctx, rgaRecordReq)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot start RGA data recording: %v", err))
					}
					err = writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, "RGA data recording started", defaultAuthor)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
					}
					rgaRecording = true
				}
			}
			ticks++
		case rtd = <-dataChan:
		case <- s.CancelChan:
			s.Logger.Debug().Msg("Shut 'er down")
			// stop rga recording if recording
			if rgaRecording {
				stopRgaReq := &fmtrpc.StopRecRequest{
					Type: fmtrpc.RecordService_RGA,
				}
				_, err := s.client.StopRecording(ctx, stopRgaReq)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Cannot stop RGA data recording: %v", err))
				} else {
					rgaRecording = false
				}
			}
			stopFlukeReq := &fmtrpc.StopRecRequest{
				Type: fmtrpc.RecordService_FLUKE,
			}
			_, err := s.client.StopRecording(ctx, stopFlukeReq)
			if err != nil {
				s.Logger.Error().Msg(fmt.Sprintf("Cannot stop Fluke data recording: %v", err))
			}
			s.clientConn.Close()
			ticker.Stop()
			return
		}
	}
}

// StartTestPlan is called by gRPC client and CLI to begin the execution of the loaded test plan
func (s *TestPlanService) StartTestPlan(ctx context.Context, req *fmtrpc.StartTestPlanRequest) (*fmtrpc.StartTestPlanResponse, error) {
	s.Logger.Info().Msg("Starting execution of a test plan")
	err := s.startTestPlan()
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Could not start execution of test plan: %v", err))
		return nil, fmt.Errorf("Could not start executing test plan: %v", err)
	}
	s.Logger.Info().Msg("Test plan execution successfully started.")
	return &fmtrpc.StartTestPlanResponse{
		Msg: fmt.Sprintf("Execution of %s test plan successfully started.", s.testPlan.Name),
	}, nil
}

// stopTestPlan will end the execution of the test plan
func (s *TestPlanService) stopTestPlan() error {
	s.Logger.Info().Msg("Stopping test...")
	err := writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, "Stopping test...", defaultAuthor)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	if ok := atomic.CompareAndSwapInt32(&s.Executing, 1, 0); !ok {
		s.Logger.Error().Msg("Could not stop test plan execution: execution already stopped.")
		err = writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_ERROR, "Could not stop test plan execution: execution already stopped.", defaultAuthor)
		if err != nil {
			s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
		}
		return fmt.Errorf("Could not stop test plan execution: execution already stopped.")
	}
	close(s.CancelChan)
	s.wg.Wait()
	err = writeMsgToReport(s.reportWriter, fmtrpc.ReportLvl_INFO, "Test stopped.", defaultAuthor)
	if err != nil {
		s.Logger.Error().Msg(fmt.Sprintf("Cannot write to report: %v", err))
	}
	s.testPlanCleanup()
	s.Logger.Info().Msg("Test stopped.")
	return nil
}

// StopTestPlan is called by gRPC client and CLI to end the execution of the loaded test plan
func (s *TestPlanService) StopTestPlan(ctx context.Context, req *fmtrpc.StopTestPlanRequest) (*fmtrpc.StopTestPlanResponse, error) {
	err := s.stopTestPlan()
	if err != nil {
		return nil, fmt.Errorf("Could not stop executing test plan: %v", err)
	}
	return &fmtrpc.StopTestPlanResponse{
		Msg: fmt.Sprintf("Execution of %s test plan successfully stopped.", s.testPlan.Name),
	}, nil
}

// InsertROIMarker creates a new entry in the rest report
func (s *TestPlanService) InsertROIMarker(ctx context.Context, req *fmtrpc.InsertROIRequest) (*fmtrpc.InsertROIResponse, error) {
	if atomic.LoadInt32(&s.Executing) != 1 {
		return nil, fmt.Errorf("Could not insert Region of Interest marker: no test plan currently being executed.")
	}
	err := writeMsgToReport(s.reportWriter, req.ReportLvl, req.Text, req.Author)
	if err != nil {
		return nil, fmt.Errorf("Cannot write to report: %v", err)
	}
	return &fmtrpc.InsertROIResponse{
		Msg: "Region of Interest marker inserted successfully.",
	}, nil
}