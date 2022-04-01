// +build demo

package fmtd

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/SSSOC-CAN/fmtd/controller"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/rga"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/telemetry"
	"github.com/SSSOC-CAN/fmtd/testplan"
	"github.com/SSSOC-CAN/fmtd/unlocker"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	testingConsoleOutput bool = false
	defaultTestingPwd = []byte("abcdefgh")
)

func initFmtd(t *testing.T, shutdownInterceptor *intercept.Interceptor, tempDir string, readySigChan chan struct{}) {
	config, err := InitConfig(true)
	if err != nil {
		t.Fatalf("Could not initialize config: %v", err)
	}
	// Manually change some config params
	config.ConsoleOutput = testingConsoleOutput
	config.DefaultLogDir = false
	config.LogFileDir = tempDir
	config.DataOutputDir = tempDir
	config.MacaroonDBPath = path.Join(tempDir, "macaroon.db")
	config.TLSCertPath = path.Join(tempDir, "tls.cert")
	config.TLSKeyPath = path.Join(tempDir, "tls.key")
	config.AdminMacPath = path.Join(tempDir, "admin.macaroon")
	config.TestMacPath = path.Join(tempDir, "test.macaroon")
	// logger
	log, err := InitLogger(&config)
	if err != nil {
		t.Fatalf("Could not initialize logger: %v", err)
	}
	shutdownInterceptor.Logger = &log
	server, err := InitServer(&config, &log)
	if err != nil {
		t.Fatalf("Could not initialize server: %v", err)
	}
	// Now we replicate Main
	var services []data.Service
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create State stores
	rtdStateStore := state.CreateStore(RtdInitialState, RtdReducer)
	ctrlStateStore := state.CreateStore(controller.InitialState, controller.ControllerReducer)
	
	// Starting main server
	err = server.Start()
	if err != nil {
		t.Error("Could not start server")
	}
	defer server.Stop()

	// Instantiating RPC server
	rpcServer, err := NewRpcServer(shutdownInterceptor, server.cfg, server.logger)
	if err != nil {
		t.Errorf("Could not initialize RPC server: %v", err)
	}
	t.Log("RPC Server Initialized.")

	// Creating gRPC server and Server options
	grpc_interceptor := intercept.NewGrpcInterceptor(rpcServer.SubLogger, false)
	err = grpc_interceptor.AddPermissions(MainGrpcServerPermissions())
	if err != nil {
		t.Errorf("Could not add permissions to gRPC middleware: %v", err)
	}
	rpcServerOpts := grpc_interceptor.CreateGrpcOptions()
	grpc_server := grpc.NewServer(rpcServerOpts...)
	defer grpc_server.Stop()
	rpcServer.RegisterWithGrpcServer(grpc_server)
	rpcServer.AddGrpcInterceptor(grpc_interceptor)

	// Instantiate RTD Service
	t.Log("Instantiating RTD subservice...")
	rtdService := data.NewRTDService(&NewSubLogger(server.logger, "RTD").SubLogger, server.cfg.TCPAddr, server.cfg.TCPPort, rtdStateStore)
	err = rtdService.RegisterWithGrpcServer(grpc_server)
	if err != nil {
		t.Errorf("Unable to register RTD Service with gRPC server: %v", err)
	}
	services = append(services, rtdService)
	t.Log("RTD service instantiated.")

	// Instantiate Telemetry service and register with gRPC server but NOT start
	t.Log("Instantiating RPC subservices and registering with gRPC server...")
	daqConn, err := drivers.ConnectToDAQ()
	if err != nil {
		t.Errorf("Unable to connect to DAQ: %v", err)
	}
	daqConnAssert, ok := daqConn.(*drivers.DAQConnection)
	if !ok {
		t.Errorf("Unable to connect to DAQ: %v", errors.ErrInvalidType)
	}
	telemetryService := telemetry.NewTelemetryService(
		&NewSubLogger(server.logger, "TEL").SubLogger,
		server.cfg.DataOutputDir,
		rtdStateStore,
		ctrlStateStore,
		daqConnAssert,
	)
	telemetryService.RegisterWithRTDService(rtdService)
	services = append(services, telemetryService)

	// Instantiate RGA service and register with gRPC server but NOT start
	rgaConn, err := drivers.ConnectToRGA()
	if err != nil {
		t.Errorf("Unable to connect to RGA: %v", err)
	}
	rgaConnAssert, ok := rgaConn.(*drivers.RGAConnection)
	if !ok {
		t.Errorf("Unable to connect to RGA: %v", errors.ErrInvalidType)
	}
	rgaService := rga.NewRGAService(
		&NewSubLogger(server.logger, "RGA").SubLogger,
		server.cfg.DataOutputDir,
		rtdStateStore,
		ctrlStateStore,
		rgaConnAssert,
	)
	rgaService.RegisterWithRTDService(rtdService)
	services = append(services, rgaService)

	// Instantiate Controller Service and register with gRPC server but not start
	ctrlConn, err := drivers.ConnectToController()
	if err != nil {
		t.Errorf("Unable to connect to controller: %v", err)
	}
	ctrlConnAssert, ok := ctrlConn.(*drivers.ControllerConnection)
	if !ok {
		t.Errorf("Unable to connect to controller: %v", errors.ErrInvalidType)
	}
	controllerService := controller.NewControllerService(
		&NewSubLogger(server.logger, "CTRL").SubLogger,
		rtdStateStore,
		ctrlStateStore,
		ctrlConnAssert,
	)
	err = controllerService.RegisterWithGrpcServer(grpc_server)
	if err != nil {
		t.Errorf("Unable to register controller Service with gRPC server: %v", err)
	}
	services = append(services, controllerService)
	t.Log("Controller service instantiated")

	// Instantiate Test Plan Executor and register with gRPC server but NOT start
	getConnectionFunc := func() (fmtrpc.DataCollectorClient, func(), error) {
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		cleanUp := func() {
			conn.Close()
		}
		return fmtrpc.NewDataCollectorClient(conn), cleanUp, nil
	}
	testPlanExecutor := testplan.NewTestPlanService(
		&NewSubLogger(server.logger, "TPEX").SubLogger,
		getConnectionFunc,
		rtdStateStore,
	)
	testPlanExecutor.RegisterWithGrpcServer(grpc_server)
	services = append(services, testPlanExecutor)
	
	t.Log("RPC subservices instantiated and registered successfully.")

	// Starting kvdb
	t.Log("Opening database...")
	db, err := kvdb.NewDB(server.cfg.MacaroonDBPath)
	if err != nil {
		t.Errorf("Could not initialize Macaroon DB: %v", err)
	}
	t.Log("Database successfully opened.")
	defer db.Close()

	// Instantiate Unlocker Service and register with gRPC server
	t.Log("Initializing unlocker service...")
	unlockerService, err := unlocker.InitUnlockerService(db, []string{server.cfg.AdminMacPath, server.cfg.TestMacPath})
	if err != nil {
		t.Errorf("Could not initialize unlocker service: %v", err)
	}
	unlockerService.RegisterWithGrpcServer(grpc_server)
	t.Log("Unlocker service initialized.")

	// Starting RPC server
	err = rpcServer.Start()
	if err != nil {
		t.Errorf("Could not start RPC server: %v", err)
	}
	defer rpcServer.Stop()

	// Start gRPC listening
	lis = bufconn.Listen(bufSize)
	err = startGrpcListen(grpc_server, lis)
	if err != nil {
		t.Fatalf("Could not start gRPC listen on %v:%v", lis.Addr(), err)
	}
	t.Logf("gRPC listening on %v", lis.Addr())
	
	// TODO:SSSOCPaulCote - BufConn RESTProxy
	readySigChan<-struct{}{}
	// Wait for password
	grpc_interceptor.SetDaemonLocked()
	t.Log("Waiting for password. Use `fmtcli setpassword` to set a password for the first time, " +
	"`fmtcli login` to unlock the daemon with an existing password, or `fmtcli changepassword` to change the " +
	"existing password and unlock the daemon.")
	pwd, err := waitForPassword(unlockerService, shutdownInterceptor.ShutdownChannel())
	if err != nil {
		t.Fatalf("Error while awaiting password: %v", err)
	}
	grpc_interceptor.SetDaemonUnlocked()
	t.Log("Login successful")

	// Instantiating Macaroon Service
	t.Log("Initiating macaroon service...")
	macaroonService, err := macaroons.InitService(*db, "fmtd")
	if err != nil {
		t.Errorf("Unable to instantiate Macaroon service: %v", err)
	}
	t.Log("Macaroon service initialized.")
	defer macaroonService.Close()

	// Unlock Macaroon Store
	t.Log("Unlocking macaroon store...")
	err = macaroonService.CreateUnlock(&pwd)
	if err != nil {
		t.Errorf("Unable to unlock macaroon store: %v", err)
	}
	t.Log("Macaroon store unlocked.")

	// Baking Macaroons
	t.Log("Baking macaroons...")
	if !utils.FileExists(server.cfg.AdminMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.AdminMacPath, adminPermissions(), false, 0,
		)
		if err != nil {
			t.Errorf("Unable to create admin macaroon: %v", err)
		}
	}
	if !utils.FileExists(server.cfg.TestMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.TestMacPath, readPermissions, true, 120,
		)
		if err != nil {
			t.Errorf("Unable to create test macaroon: %v", err)
		}
	}
	grpc_interceptor.AddMacaroonService(macaroonService)
	rpcServer.AddMacaroonService(macaroonService)
	t.Log("Macaroons baked successfully.")

	// Start all Services
	for _, s := range services {
		err = s.Start()
		if err != nil {
			t.Errorf("Unable to start %s service: %v", s.Name(), err)
		}
		defer s.Stop()
	}

	// Change RPC state to active
	grpc_interceptor.SetRPCActive()
	<-shutdownInterceptor.ShutdownChannel()
}

// unlockFMTD is a helper function to unlock the FMT Daemon
func unlockFMTD(t *testing.T, ctx context.Context, unlockerClient fmtrpc.UnlockerClient) {
	_, err := unlockerClient.Login(ctx, &fmtrpc.LoginRequest{
		Password: defaultTestingPwd,
	})
	if err != nil {
		t.Errorf("Error unlocking FMTD: %v", err)
	}
}

type unlockerCases struct {
	caseName	string
	command		string
	oldPwd		[]byte
	newPwd		[]byte
	newMacKey	bool
	expectErr   bool
}

var (
	chngePwdCmd string = "changepassword"
	loginCmd    string = "login"
	setPwdCmd   string = "setpassword"
	unlockerTestCases = []unlockerCases{
		{"change-password-before-setting-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), false, true},
		{"change-password-before-setting-2", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), true, true},
		{"login-before-setting", loginCmd, defaultTestingPwd, nil, false, true},
		{"set-password-valid", setPwdCmd, defaultTestingPwd, nil, false, false},
		{"set-password-after-setting", setPwdCmd, []byte("bbbbbbbb"), nil, false, true},
		{"login-after-setting", loginCmd, defaultTestingPwd, nil, false, true},
		{"change-password-after-setting-invalid-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), false, true},
		{"change-password-after-setting-invalid-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), true, true},
	}
)

// TestUnlockerGrpcApi tests the various API endpoints using the gRPC protocol
func TestUnlockerGrpcApi(t *testing.T) {
	// temp dir
	tempDir, err := ioutil.TempDir("", "integration-testing-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	// SIGINT interceptor
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		t.Fatalf("Could not initialize the shutdown interceptor: %v", err)
	}
	defer shutdownInterceptor.RequestShutdown()
	readySignal := make(chan struct{})
	defer close(readySignal)
	go initFmtd(t, shutdownInterceptor, tempDir, readySignal)
	_ = <-readySignal
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewUnlockerClient(conn)
	// Test ChangePassword before setting pwd
	for _, c := range unlockerTestCases {
		t.Run(c.caseName, func(t *testing.T) {
			switch c.command {
			case chngePwdCmd:
				resp, err := client.ChangePassword(ctx, &fmtrpc.ChangePwdRequest{
					CurrentPassword: c.oldPwd,
					NewPassword: c.newPwd,
					NewMacaroonRootKey: c.newMacKey,
				})
				t.Log(resp)
				if c.expectErr {
					if err == nil {
						t.Error("Expected an error and got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected errror occured: %v", err)
					}
				}
			case loginCmd:
				resp, err := client.Login(ctx, &fmtrpc.LoginRequest{
					Password: c.oldPwd,
				})
				t.Log(resp)
				if c.expectErr {
					if err == nil {
						t.Error("Expected an error and got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected errror occured: %v", err)
					}
				}
			case setPwdCmd:
				resp, err := client.SetPassword(ctx, &fmtrpc.SetPwdRequest{
					Password: c.oldPwd,
				})
				t.Log(resp)
				if c.expectErr {
					if err == nil {
						t.Error("Expected an error and got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected errror occured: %v", err)
					}
				}
			}
		})
	}
}