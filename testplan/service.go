package testplan

import (
	"context"
	"os"
	"sync/atomic"
	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"google.golang.org/grpc"
)

var (
	testPlanExecName = "TPEX"
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
}

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

// NewTestPlanService instantiates the TestPlanService struct
func NewTestPlanService(logger *zerolog.Logger) *TestPlanService {
	return &TestPlanService{
		Logger: logger,
		name: testPlanExecName,
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

func (s *TestPlanService) StartTestPlan(ctx context.Context, req *fmtrpc.StartTestPlanRequest) (*fmtrpc.StartTestPlanResponse, error) {}

func (s *TestPlanService) StopTestPlan(ctx context.Context, req *fmtrpc.StopTestPlanRequest) (*fmtrpc.StopTestPlanResponse, error) {}

func (s *TestPlanService) InsertROIMarker(ctx context.Context, req *fmtrpc.InsertROIRequest) (*fmtrpc.InsertROIResponse, error) {}