//go:build demo
// +build demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package fmtd

import (
	"context"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/controller"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/fmtrpc/demorpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/rga"
	"github.com/SSSOC-CAN/fmtd/telemetry"
	"github.com/SSSOC-CAN/fmtd/testplan"
	"github.com/SSSOC-CAN/fmtd/unlocker"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/SSSOCPaulCote/gux"
	e "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	macaroon "gopkg.in/macaroon.v2"
)

const (
	ErrMacNotExpired = bg.Error("macaroon didn't expire")
)

var (
	testingConsoleOutput bool = false
	defaultTestingPwd         = []byte("abcdefgh")
	defaultBucketName         = "test"
	defaultOrgName            = "sssoc"
)

func initFmtd(t *testing.T, shutdownInterceptor *intercept.Interceptor, readySigChan chan struct{}, wg *sync.WaitGroup, tempDir string) {
	defer wg.Done()
	// create config and log
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
	rtdStateStore := gux.CreateStore(RtdInitialState, RtdReducer)
	ctrlStateStore := gux.CreateStore(controller.InitialState, controller.ControllerReducer)

	// Starting main server
	err = server.Start()
	if err != nil {
		t.Error("Could not start server")
	}
	defer func() {
		t.Log("Stopping main server...")
		err := server.Stop()
		if err != nil {
			t.Errorf("Could not stop main server: %v", err)
		}
		t.Log("Main server stopped.")
	}()

	// Get TLS config
	t.Log("Loading TLS configuration...")
	serverOpts, _, _, cleanUp, err := cert.GetTLSConfig(server.cfg.TLSCertPath, server.cfg.TLSKeyPath, server.cfg.ExtraIPAddr)
	if err != nil {
		t.Errorf("Could not load TLS configuration: %v", err)
	}
	t.Log("TLS configuration successfully loaded.")
	defer cleanUp()

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
	serverOpts = append(serverOpts, rpcServerOpts...)
	grpc_server := grpc.NewServer(serverOpts...)
	defer grpc_server.Stop()
	rpcServer.RegisterWithGrpcServer(grpc_server)
	rpcServer.AddGrpcInterceptor(grpc_interceptor)

	// Instantiate RTD Service
	t.Log("Instantiating RTD subservice...")
	rtdService := data.NewRTDService(&NewSubLogger(server.logger, "RTD").SubLogger, rtdStateStore)
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
		rtdStateStore,
		ctrlStateStore,
		daqConnAssert,
		config.InfluxURL,
		config.InfluxAPIToken,
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
		rtdStateStore,
		ctrlStateStore,
		rgaConnAssert,
		config.InfluxURL,
		config.InfluxAPIToken,
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
	defer func() {
		t.Log("Closing database...")
		err := db.Close()
		if err != nil {
			t.Errorf("Could not close database: %v", err)
		}
		t.Log("Database closed")
	}()

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
	defer func() {
		t.Log("Shutting down RPC server...")
		err := rpcServer.Stop()
		if err != nil {
			t.Errorf("Could not shutdown RPC server: %v", err)
		}
		t.Log("RPC server shutdown")
	}()

	// Start gRPC listening
	lis = bufconn.Listen(bufSize)
	err = startGrpcListen(grpc_server, lis)
	if err != nil {
		t.Fatalf("Could not start gRPC listen on %v:%v", lis.Addr(), err)
	}
	t.Logf("gRPC listening on %v", lis.Addr())

	// TODO:SSSOCPaulCote - BufConn RESTProxy
	readySigChan <- struct{}{}
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
	macaroonService, err := macaroons.InitService(*db, "fmtd", zerolog.New(os.Stderr).With().Timestamp().Logger(), []string{}, []string{})
	if err != nil {
		t.Errorf("Unable to instantiate Macaroon service: %v", err)
	}
	t.Log("Macaroon service initialized.")
	defer func() {
		t.Log("Shutting down macaroon service...")
		err := macaroonService.Close()
		if err != nil {
			t.Errorf("Could not shutdown macaroon service: %v", err)
		}
		t.Log("Macaroon service shutdown")
	}()

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
		t.Logf("Starting %s service...", s.Name())
		err = s.Start()
		if err != nil {
			t.Errorf("Unable to start %s service: %v", s.Name(), err)
		}
		t.Logf("%s service started", s.Name())
	}
	// Stop Services CleanUp
	cleanUpServices := func() {
		for i := len(services) - 1; i > -1; i-- {
			t.Logf("Shutting down %s service...", services[i].Name())
			err := services[i].Stop()
			if err != nil {
				t.Errorf("Could not shutdown %s service: %v", services[i].Name(), err)
			}
			t.Logf("%s service shutdown", services[i].Name())
		}
	}
	defer cleanUpServices()

	// Change RPC state to active
	grpc_interceptor.SetRPCActive()
	<-shutdownInterceptor.ShutdownChannel()
	return
}

// unlockFMTD is a helper function to unlock the FMT Daemon
func unlockFMTD(ctx context.Context, unlockerClient fmtrpc.UnlockerClient) error {
	_, err := unlockerClient.SetPassword(ctx, &fmtrpc.SetPwdRequest{
		Password: defaultTestingPwd,
	})
	if err != nil {
		return err
	}
	return nil
}

type unlockerCases struct {
	caseName  string
	command   string
	oldPwd    []byte
	newPwd    []byte
	newMacKey bool
	expectErr bool
}

var (
	chngePwdCmd          string = "changepassword"
	loginCmd             string = "login"
	setPwdCmd            string = "setpassword"
	unlockerTestCasesOne        = []unlockerCases{
		{"change-password-before-setting-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), false, true},
		{"change-password-before-setting-2", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), true, true},
		{"login-before-setting", loginCmd, defaultTestingPwd, nil, false, true},
		{"set-password-valid", setPwdCmd, defaultTestingPwd, nil, false, false},
		{"set-password-after-setting", setPwdCmd, []byte("bbbbbbbb"), nil, false, true},
		{"login-after-setting", loginCmd, defaultTestingPwd, nil, false, true},
		{"change-password-after-setting-invalid-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), false, true},
		{"change-password-after-setting-invalid-2", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), true, true},
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
	var wg sync.WaitGroup
	// SIGINT interceptor
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		t.Fatalf("Could not initialize the shutdown interceptor: %v", err)
	}
	defer shutdownInterceptor.Close()
	readySignal := make(chan struct{})
	defer close(readySignal)
	wg.Add(1)
	go initFmtd(t, shutdownInterceptor, readySignal, &wg, tempDir)
	<-readySignal
	ctx := context.Background()
	creds, err := credentials.NewClientTLSFromFile(path.Join(tempDir, "tls.cert"), "")
	if err != nil {
		t.Fatal("Could not get TLS credentials from file")
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(bufDialer),
	}
	conn, err := grpc.DialContext(ctx, "bufnet", opts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewUnlockerClient(conn)
	// Test ChangePassword before setting pwd
	for _, c := range unlockerTestCasesOne {
		t.Run(c.caseName, func(t *testing.T) {
			switch c.command {
			case chngePwdCmd:
				resp, err := client.ChangePassword(ctx, &fmtrpc.ChangePwdRequest{
					CurrentPassword:    c.oldPwd,
					NewPassword:        c.newPwd,
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
	time.Sleep(1 * time.Second)
	shutdownInterceptor.RequestShutdown()
	wg.Wait()
	if !utils.FileExists(path.Join(tempDir, "admin.macaroon")) || !utils.FileExists(path.Join(tempDir, "test.macaroon")) {
		t.Error("Macaroon files don't exist!")
	}
}

// getMacaroonGrpcCreds gets the appropriate grpc DialOption for macaroon authentication
func getMacaroonGrpcCreds(tlsCertPath, adminMacPath string, macTimeout int64) ([]grpc.DialOption, error) {
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		return nil, e.Wrap(err, "could not get TLS credentials from file")
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	adminMac, err := os.ReadFile(adminMacPath)
	if err != nil {
		return nil, e.Wrapf(err, "Could not read macaroon at %v", adminMacPath)
	}
	macHex := hex.EncodeToString(adminMac)
	mac, err := auth.LoadMacaroon(auth.ReadPassword, macHex)
	if err != nil {
		return nil, err
	}
	// Add constraints to our macaroon
	macConstraints := []macaroons.Constraint{
		macaroons.TimeoutConstraint(macTimeout), // prevent a replay attack
	}
	constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
	if err != nil {
		return nil, err
	}
	cred, err := macaroons.NewMacaroonCredential(constrainedMac)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithPerRPCCredentials(cred))
	return opts, nil
}

type bakeMacCases struct {
	caseName   string
	bakeMacReq *fmtrpc.BakeMacaroonRequest
	expectErr  bool
	callback   func(string, grpc.DialOption) (*grpc.ClientConn, error)
}

var (
	bakeMacTestCases = []bakeMacCases{
		{"bake-mac-empty-permissions", &fmtrpc.BakeMacaroonRequest{}, true, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
			return nil, nil
		}},
		{"bake-mac-invalid-permission-entity", &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(-10),
			TimeoutType: fmtrpc.TimeoutType_DAY,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "not a real entity",
					Action: "invalid action",
				},
			},
		}, true, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
			return nil, nil
		}},
		{"bake-mac-invalid-permission-action", &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(-10),
			TimeoutType: fmtrpc.TimeoutType_DAY,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "tpex",
					Action: "invalid action",
				},
			},
		}, true, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
			return nil, nil
		}},
		{"bake-mac-10-sec-timeout", &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(10),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "uri",
					Action: "/fmtrpc.Fmt/AdminTest",
				},
			},
		}, false, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
			macBytes, err := hex.DecodeString(macHex)
			if err != nil {
				return nil, e.Wrap(err, "unable to hex decode macaroon")
			}
			mac := &macaroon.Macaroon{}
			if err = mac.UnmarshalBinary(macBytes); err != nil {
				return nil, e.Wrap(err, "unable to decode macaroon")
			}
			cred, err := macaroons.NewMacaroonCredential(mac)
			if err != nil {
				return nil, e.Wrap(err, "unable to get gRPC macaroon credential")
			}
			ctx := context.Background()
			opts := []grpc.DialOption{
				tlsDialOpt,
				grpc.WithPerRPCCredentials(cred),
				grpc.WithContextDialer(bufDialer),
			}
			conn, err := grpc.DialContext(ctx, "bufnet", opts...)
			if err != nil {
				return nil, e.Wrap(err, "cannot dial bufnet")
			}
			client := fmtrpc.NewFmtClient(conn)
			_, err = client.AdminTest(ctx, &fmtrpc.AdminTestRequest{})
			if err != nil {
				return nil, e.Wrap(err, "unable to invoke admin test command")
			}
			_, err = client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{})
			if err == nil {
				return nil, errors.ErrNoError
			}
			time.Sleep(10 * time.Second)
			_, err = client.AdminTest(ctx, &fmtrpc.AdminTestRequest{})
			if err == nil {
				return nil, ErrMacNotExpired
			}
			return conn, nil
		}},
	}
)

// TestFmtGrpcApi tests the Fmt API service
func TestFmtGrpcApi(t *testing.T) {
	// temp dir
	tempDir, err := ioutil.TempDir("", "integration-testing-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	var wg sync.WaitGroup
	// SIGINT interceptor
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		t.Fatalf("Could not initialize the shutdown interceptor: %v", err)
	}
	defer shutdownInterceptor.Close()
	readySignal := make(chan struct{})
	defer close(readySignal)
	wg.Add(1)
	go initFmtd(t, shutdownInterceptor, readySignal, &wg, tempDir)
	<-readySignal
	ctx := context.Background()
	creds, err := credentials.NewClientTLSFromFile(path.Join(tempDir, "tls.cert"), "")
	if err != nil {
		t.Fatal("Could not get TLS credentials from file")
	}
	unlockerOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(bufDialer),
	}
	conn, err := grpc.DialContext(ctx, "bufnet", unlockerOpts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	unlockerClient := fmtrpc.NewUnlockerClient(conn)
	err = unlockFMTD(ctx, unlockerClient)
	if err != nil {
		t.Fatalf("Could not set FMTD password: %v", err)
	}
	time.Sleep(1 * time.Second)
	// now we get FmtClient
	opts := []grpc.DialOption{grpc.WithContextDialer(bufDialer)}
	macDialOpts, err := getMacaroonGrpcCreds(path.Join(tempDir, "tls.cert"), path.Join(tempDir, "admin.macaroon"), int64(60))
	opts = append(opts, macDialOpts...)
	authConn, err := grpc.DialContext(ctx, "bufnet", opts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer authConn.Close()
	fmtClient := fmtrpc.NewFmtClient(authConn)
	// test command
	t.Run("test-command", func(t *testing.T) {
		resp, err := fmtClient.TestCommand(ctx, &fmtrpc.TestRequest{})
		if err != nil {
			t.Errorf("Unable to invoke test command: %v", err)
		}
		if resp.Msg != "This is a regular test" {
			t.Errorf("Unexpected test command response: %v", resp.Msg)
		}
	})
	// admin-test
	t.Run("admin-test", func(t *testing.T) {
		resp, err := fmtClient.AdminTest(ctx, &fmtrpc.AdminTestRequest{})
		if err != nil {
			t.Errorf("Unable to invoke admin test command: %v", err)
		}
		if resp.Msg != "This is an admin test" {
			t.Errorf("Unexpected admin test response: %v", resp.Msg)
		}
	})
	// bake-macaroon
	for _, c := range bakeMacTestCases {
		t.Run(c.caseName, func(t *testing.T) {
			resp, err := fmtClient.BakeMacaroon(ctx, c.bakeMacReq)
			if c.expectErr {
				if err == nil {
					t.Error("Expected an error and got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unable to bake macaroon: %v", err)
				}
				testConn, err := c.callback(resp.Macaroon, grpc.WithTransportCredentials(creds))
				if err != nil {
					t.Errorf("Error when testing freshly baked macaroon: %v", err)
				}
				if testConn != nil {
					defer testConn.Close()
				}
			}
		})
	}
	// Stop Daemon
	t.Run("stop-daemon", func(t *testing.T) {
		_, err := fmtClient.StopDaemon(ctx, &fmtrpc.StopRequest{})
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Error was not a gRPC status error")
			}
			if st.Message() != "error reading from server: EOF" {
				t.Errorf("Unable to stop daemon: %v", st.Message())
			}
		}
	})
	wg.Wait()
}

// TestDataCollectorGrpcApi tests the data collector API service
func TestDataCollectorGrpcApi(t *testing.T) {
	t.Skip("Cannot be currently tested in this release")
	// temp dir
	tempDir, err := ioutil.TempDir("", "integration-testing-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	var wg sync.WaitGroup
	// SIGINT interceptor
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		t.Fatalf("Could not initialize the shutdown interceptor: %v", err)
	}
	defer shutdownInterceptor.Close()
	readySignal := make(chan struct{})
	defer close(readySignal)
	wg.Add(1)
	go initFmtd(t, shutdownInterceptor, readySignal, &wg, tempDir)
	<-readySignal
	ctx := context.Background()
	creds, err := credentials.NewClientTLSFromFile(path.Join(tempDir, "tls.cert"), "")
	if err != nil {
		t.Fatal("Could not get TLS credentials from file")
	}
	unlockerOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(bufDialer),
	}
	conn, err := grpc.DialContext(ctx, "bufnet", unlockerOpts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	unlockerClient := fmtrpc.NewUnlockerClient(conn)
	err = unlockFMTD(ctx, unlockerClient)
	if err != nil {
		t.Fatalf("Could not set FMTD password: %v", err)
	}
	time.Sleep(1 * time.Second)
	// now we get DataCollectorClient
	opts := []grpc.DialOption{grpc.WithContextDialer(bufDialer)}
	macDialOpts, err := getMacaroonGrpcCreds(path.Join(tempDir, "tls.cert"), path.Join(tempDir, "admin.macaroon"), int64(600))
	opts = append(opts, macDialOpts...)
	authConn, err := grpc.DialContext(ctx, "bufnet", opts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer authConn.Close()
	client := fmtrpc.NewDataCollectorClient(authConn)
	t.Run("start-record-invalid-polling-interval", func(t *testing.T) {
		resp, err := client.StartRecording(ctx, &fmtrpc.RecordRequest{
			PollingInterval: int64(1),
			Type:            fmtrpc.RecordService_TELEMETRY,
			OrgName:         defaultOrgName,
			BucketName:      defaultBucketName,
		})
		if err == nil {
			t.Error("Expected an error but got none")
		}
		t.Log(resp)
	})
	t.Run("start-record-rga-service-before-telemetry", func(t *testing.T) {
		resp, err := client.StartRecording(ctx, &fmtrpc.RecordRequest{
			Type:    fmtrpc.RecordService_RGA,
			OrgName: defaultOrgName,
		})
		if err == nil {
			t.Error("Expected an error but got none")
		}
		t.Log(resp)
	})
	t.Run("start-record-telemetry", func(t *testing.T) {
		resp, err := client.StartRecording(ctx, &fmtrpc.RecordRequest{
			Type:       fmtrpc.RecordService_TELEMETRY,
			OrgName:    defaultOrgName,
			BucketName: defaultBucketName,
		})
		if err != nil {
			t.Errorf("Could not start telemetry: %v", err)
		}
		t.Log(resp)
	})
	t.Run("start-record-telemetry-after-starting", func(t *testing.T) {
		resp, err := client.StartRecording(ctx, &fmtrpc.RecordRequest{
			Type:       fmtrpc.RecordService_TELEMETRY,
			OrgName:    defaultOrgName,
			BucketName: defaultBucketName,
		})
		if err == nil {
			t.Error("Expected an error but got none")
		}
		t.Log(resp)
	})
	t.Run("stop-record-rga", func(t *testing.T) {
		resp, err := client.StopRecording(ctx, &fmtrpc.StopRecRequest{
			Type: fmtrpc.RecordService_RGA,
		})
		if err == nil {
			t.Error("Expected an error but got none")
		}
		t.Log(resp)
	})
	t.Run("stop-record-telemetry", func(t *testing.T) {
		resp, err := client.StopRecording(ctx, &fmtrpc.StopRecRequest{
			Type: fmtrpc.RecordService_TELEMETRY,
		})
		if err != nil {
			t.Errorf("Could not stop telemetry recording: %v", err)
		}
		t.Log(resp)
	})
	t.Run("stop-record-telemetry-after-stopping", func(t *testing.T) {
		resp, err := client.StopRecording(ctx, &fmtrpc.StopRecRequest{
			Type: fmtrpc.RecordService_TELEMETRY,
		})
		if err == nil {
			t.Error("Expected an error but got none")
		}
		t.Log(resp)
	})
	t.Run("restart-record-telemetry", func(t *testing.T) {
		resp, err := client.StartRecording(ctx, &fmtrpc.RecordRequest{
			Type:       fmtrpc.RecordService_TELEMETRY,
			OrgName:    defaultOrgName,
			BucketName: defaultBucketName,
		})
		if err != nil {
			t.Errorf("Could not start telemetry: %v", err)
		}
		t.Log(resp)
	})
	t.Run("subscribe-datastream", func(t *testing.T) {
		stream, err := client.SubscribeDataStream(ctx, &fmtrpc.SubscribeDataRequest{})
		if err != nil {
			t.Errorf("Could not start telemetry: %v", err)
		}
		ticker := time.NewTicker(time.Duration(int64(10)) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				return
			default:
				_, err = stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					t.Error(err)
					return
				}
			}
		}
	})
	t.Run("subscribe-datastream-100-subscribers", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ticker := time.NewTicker(time.Duration(int64(15)) * time.Second)
				defer ticker.Stop()
				stream, err := client.SubscribeDataStream(ctx, &fmtrpc.SubscribeDataRequest{})
				if err != nil {
					t.Errorf("Could not start telemetry: %v", err)
				}
				for {
					select {
					case <-ticker.C:
						return
					default:
						_, err = stream.Recv()
						if err == io.EOF {
							return
						}
						if err != nil {
							t.Error(err)
							return
						}
					}
				}
			}()
		}
		wg.Wait()
	})
	t.Run("subscribe-datastream-stress-test", func(t *testing.T) {
		var wg sync.WaitGroup
		ticker := time.NewTicker(time.Duration(int64(10)) * time.Second)
		var (
			shutdownChans []chan struct{}
			cumGoRoutines int = 0
		)
		for i := 0; i < 10; i++ {
			select {
			case <-ticker.C:
				// if it's not the first round, then close 10 listener go routines at random times
				if i != 0 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for k := 0; k < 10; k++ {
							time.Sleep(time.Duration(int((rand.Float64()*1)*1000)) * time.Millisecond)
							close(shutdownChans[cumGoRoutines])
							cumGoRoutines += 1
						}
					}()
				}
				// spawn 10 subscribers
				for j := 0; j < 10; j++ {
					shutdownChan := make(chan struct{})
					shutdownChans = append(shutdownChans, shutdownChan)
					wg.Add(1)
					go func(quitChan chan struct{}) {
						defer wg.Done()
						stream, err := client.SubscribeDataStream(ctx, &fmtrpc.SubscribeDataRequest{})
						if err != nil {
							t.Errorf("Could not start telemetry: %v", err)
						}
						for {
							select {
							case <-quitChan:
								return
							default:
								_, err = stream.Recv()
								if err == io.EOF {
									return
								}
								if err != nil {
									t.Error(err)
									return
								}
							}
						}
					}(shutdownChan)
				}
			}
		}
		for k := 0; k < 10; k++ {
			time.Sleep(time.Duration(int((rand.Float64()*1)*1000)) * time.Millisecond)
			close(shutdownChans[cumGoRoutines])
			cumGoRoutines += 1
		}
		wg.Wait()
		ticker.Stop()
	})
	shutdownInterceptor.RequestShutdown()
	wg.Wait()
}

type Cases struct {
	caseName  string
	cmd       string
	expectErr bool
	setPoint  float64
	rampRate  float64
}

var (
	setTempCmd string = "set-temp"
	setPresCmd string = "set-pres"
	testCases         = []Cases{
		{"set-temp-negative-change-rate", setTempCmd, true, 30.0, -4.0},
		{"set-temp-hot-0-rate", setTempCmd, false, 30.0, 0.0},
		{"set-temp-cold-0-rate", setTempCmd, false, -20.0, 0.0},
		{"set-temp-hot-5-rate", setTempCmd, false, -15.0, 5.0},
		{"set-temp-cold-5-rate", setTempCmd, false, -20.0, 5.0},
		{"set-pres-negative-change-rate", setPresCmd, true, 30.0, -4.0},
		{"set-pres-negative-pressure", setPresCmd, true, -10.0, 0.0},
		{"set-pres-high-0-rate", setPresCmd, false, 800.0, 0.0},
		{"set-pres-low-0-rate", setPresCmd, false, 20.0, 0.0},
		{"set-pres-high-5-rate", setPresCmd, false, 25.0, 5.0},
		{"set-pres-low-5-rate", setPresCmd, false, 20.0, 5.0},
	}
)

// TestControllerGrpcApi tests the Controller API service
func TestControllerGrpcApi(t *testing.T) {
	t.Skip("Already thoroughly tested in controller package")
	// temp dir
	tempDir, err := ioutil.TempDir("", "integration-testing-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	var wg sync.WaitGroup
	// SIGINT interceptor
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		t.Fatalf("Could not initialize the shutdown interceptor: %v", err)
	}
	defer shutdownInterceptor.Close()
	readySignal := make(chan struct{})
	defer close(readySignal)
	wg.Add(1)
	go initFmtd(t, shutdownInterceptor, readySignal, &wg, tempDir)
	<-readySignal
	ctx := context.Background()
	creds, err := credentials.NewClientTLSFromFile(path.Join(tempDir, "tls.cert"), "")
	if err != nil {
		t.Fatal("Could not get TLS credentials from file")
	}
	unlockerOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(bufDialer),
	}
	conn, err := grpc.DialContext(ctx, "bufnet", unlockerOpts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	unlockerClient := fmtrpc.NewUnlockerClient(conn)
	err = unlockFMTD(ctx, unlockerClient)
	if err != nil {
		t.Fatalf("Could not set FMTD password: %v", err)
	}
	time.Sleep(1 * time.Second)
	// now we get ControllerClient
	opts := []grpc.DialOption{grpc.WithContextDialer(bufDialer)}
	macDialOpts, err := getMacaroonGrpcCreds(path.Join(tempDir, "tls.cert"), path.Join(tempDir, "admin.macaroon"), int64(1200))
	opts = append(opts, macDialOpts...)
	authConn, err := grpc.DialContext(ctx, "bufnet", opts...)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer authConn.Close()
	dataClient := fmtrpc.NewDataCollectorClient(authConn)
	_, err = dataClient.StartRecording(ctx, &fmtrpc.RecordRequest{
		Type:       fmtrpc.RecordService_TELEMETRY,
		OrgName:    defaultOrgName,
		BucketName: defaultBucketName,
	})
	if err != nil {
		t.Fatalf("Could not start telemetry data recording: %v", err)
	}
	client := demorpc.NewControllerClient(authConn)
	// Concurrent Set Temp
	t.Run("concurrent-set-temp", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			stream, err := client.SetTemperature(ctx, &demorpc.SetTempRequest{
				TempSetPoint:   30.0,
				TempChangeRate: 5.0,
			})
			if err != nil {
				t.Errorf("Could not set temperature: %v", err)
				return
			}
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					t.Error(err)
					return
				}
				t.Log(resp)
			}
		}()
		time.Sleep(5 * time.Second)
		tmpClient := demorpc.NewControllerClient(authConn)
		stream, _ := tmpClient.SetTemperature(ctx, &demorpc.SetTempRequest{
			TempSetPoint:   35.0,
			TempChangeRate: 0.0,
		})
		_, err := stream.Recv()
		if err == nil {
			t.Error("Expected an error and got none when doing concurrent temperature sets")
		}
		wg.Wait()
	})
	// Concurrent Set Pressure
	t.Run("concurrent-set-pres", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			stream, err := client.SetPressure(ctx, &demorpc.SetPresRequest{
				PressureSetPoint:   755.0,
				PressureChangeRate: 5.0,
			})
			if err != nil {
				t.Errorf("Could not set pressure: %v", err)
				return
			}
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					t.Error(err)
					return
				}
				t.Log(resp)
			}
		}()
		time.Sleep(5 * time.Second)
		tmpClient := demorpc.NewControllerClient(authConn)
		stream, _ := tmpClient.SetPressure(ctx, &demorpc.SetPresRequest{
			PressureSetPoint:   35.0,
			PressureChangeRate: 0.0,
		})
		_, err := stream.Recv()
		if err == nil {
			t.Error("Expected an error and got none when doing concurrent pressure sets")
		}
		wg.Wait()
	})
	for _, c := range testCases {
		t.Run(c.caseName, func(t *testing.T) {
			switch c.cmd {
			case "set-temp":
				stream, err := client.SetTemperature(ctx, &demorpc.SetTempRequest{
					TempSetPoint:   c.setPoint,
					TempChangeRate: c.rampRate,
				})
				if err != nil {
					t.Errorf("Could not invoke set temperature command: %v", err)
				}
				for {
					resp, err := stream.Recv()
					if err == io.EOF {
						break
					}
					if c.expectErr {
						if err == nil {
							t.Error("Expected an error and got none")
						}
						break
					} else {
						if err != nil {
							t.Errorf("Could not receive from stream: %v", err)
							break
						}
					}
					t.Log(resp)
				}
			case "set-pres":
				stream, err := client.SetPressure(ctx, &demorpc.SetPresRequest{
					PressureSetPoint:   c.setPoint,
					PressureChangeRate: c.rampRate,
				})
				if err != nil {
					t.Errorf("Could not invoke set pressure command: %v", err)
				}
				for {
					resp, err := stream.Recv()
					if err == io.EOF {
						break
					}
					if c.expectErr {
						if err == nil {
							t.Error("Expected an error and got none")
						}
						break
					} else {
						if err != nil {
							t.Errorf("Could not receive from stream: %v", err)
							break
						}
					}
					t.Log(resp)
				}
			}
		})
	}
}
