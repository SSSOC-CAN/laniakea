/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package fmtd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/plugins"
	"github.com/SSSOC-CAN/fmtd/unlocker"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	bg "github.com/SSSOCPaulCote/blunderguard"
	e "github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	fmtdDirRegexp             = `/(?:fmtd)/(?:fmtd)$`
	rootDirRegexp             = `/(?:fmtd)$`
	startPluginDir            = "plugins/plugins/testing"
	pluginCfgs                = []*fmtrpc.PluginConfig{
		&fmtrpc.PluginConfig{
			Name:        "test-rng-plugin",
			Type:        "datasource",
			ExecName:    "test_rng_plugin",
			Timeout:     30,
			MaxTimeouts: 3,
		},
		&fmtrpc.PluginConfig{
			Name:        "test-timeout-plugin",
			Type:        "datasource",
			ExecName:    "test_timeout_plugin",
			Timeout:     15,
			MaxTimeouts: 3,
		},
		&fmtrpc.PluginConfig{
			Name:        "test-ctrl-plugin",
			Type:        "controller",
			ExecName:    "test_ctrl_plugin",
			Timeout:     30,
			MaxTimeouts: 3,
		},
		&fmtrpc.PluginConfig{
			Name:        "test-error-plugin",
			Type:        "datasource",
			ExecName:    "test_error_plugin",
			Timeout:     15,
			MaxTimeouts: 3,
		},
	}
)

// getPluginDir checks where we're running the test from and determines the path to the plugin directory
func getPluginDir(t *testing.T) string {
	// first get cwd, determine if root or plugins directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get current working directory: %v", err)
	}
	// root dir
	var pluginDir string
	match, err := regexp.MatchString(fmtdDirRegexp, dir)
	if err != nil {
		t.Fatalf("could not run regular expression check: %v", err)
	}
	if !match {
		// plugins dir
		match, err = regexp.MatchString(rootDirRegexp, dir)
		if err != nil {
			t.Fatalf("could not run regular expression check: %v", err)
		}
		if !match {
			t.Fatalf("Running test from an unexpected location, please run from root directory or within plugins package directory")
		} else {
			pluginDir = filepath.Join(dir, startPluginDir)
		}
	} else {
		pluginDir = filepath.Join(filepath.Dir(dir), startPluginDir)
	}
	return pluginDir
}

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
	config.PluginDir = getPluginDir(t)
	config.Plugins = pluginCfgs[:]
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
	var services []api.Service
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	// Initialize Plugin Manager
	t.Log("Initializing plugins...")
	pluginManager := plugins.NewPluginManager(config.PluginDir, config.Plugins, *server.logger, false)
	pluginManager.RegisterWithGrpcServer(grpc_server)
	err = pluginManager.AddPermissions(StreamingPluginAPIPermission())
	if err != nil {
		t.Errorf("Could not add permissions to plugin manager: %v", err)
	}
	err = pluginManager.Start(ctx)
	if err != nil {
		t.Errorf("Unable to start plugin manager: %v", err)
	}
	defer pluginManager.Stop()
	t.Log("plugins initialized")
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
	macaroonService, err := macaroons.InitService(*db, "fmtd", zerolog.New(os.Stderr).With().Timestamp().Logger(), []string{})
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
			ctx, macaroonService, server.cfg.AdminMacPath, adminPermissions(), 0, []string{},
		)
		if err != nil {
			t.Errorf("Unable to create admin macaroon: %v", err)
		}
	}
	if !utils.FileExists(server.cfg.TestMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.TestMacPath, readPermissions, 120, []string{},
		)
		if err != nil {
			t.Errorf("Unable to create test macaroon: %v", err)
		}
	}
	grpc_interceptor.AddMacaroonService(macaroonService)
	rpcServer.AddMacaroonService(macaroonService)
	pluginManager.AddMacaroonService(macaroonService)
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
	caseName    string
	command     string
	oldPwd      []byte
	newPwd      []byte
	newMacKey   bool
	expectedErr error
}

var (
	chngePwdCmd          string = "changepassword"
	loginCmd             string = "login"
	setPwdCmd            string = "setpassword"
	unlockerTestCasesOne        = []unlockerCases{
		{"change-password-before-setting-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), false, unlocker.ErrPasswordNotSet},
		{"change-password-before-setting-2", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), true, unlocker.ErrPasswordNotSet},
		{"login-before-setting", loginCmd, defaultTestingPwd, nil, false, unlocker.ErrPasswordNotSet},
		{"set-password-valid", setPwdCmd, defaultTestingPwd, nil, false, nil},
		{"set-password-after-setting", setPwdCmd, []byte("bbbbbbbb"), nil, false, intercept.ErrDaemonUnlocked},
		{"login-after-setting", loginCmd, defaultTestingPwd, nil, false, intercept.ErrDaemonUnlocked},
		{"change-password-after-setting-invalid-1", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), false, intercept.ErrDaemonUnlocked},
		{"change-password-after-setting-invalid-2", chngePwdCmd, defaultTestingPwd, []byte("aaaaaaaa"), true, intercept.ErrDaemonUnlocked},
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
				_, err := client.ChangePassword(ctx, &fmtrpc.ChangePwdRequest{
					CurrentPassword:    c.oldPwd,
					NewPassword:        c.newPwd,
					NewMacaroonRootKey: c.newMacKey,
				})
				if c.expectedErr == nil {
					if err != c.expectedErr {
						t.Fatalf("Unexpected error when calling ChangePassword: %v", err)
					}
				} else {
					st, ok := status.FromError(err)
					if !ok {
						t.Errorf("Unexpected error format when calling ChangePassword")
					}
					if st.Message() != c.expectedErr.Error() {
						t.Errorf("Unexpected error when calling ChangePassword: %v", err)
					}
				}
			case loginCmd:
				_, err := client.Login(ctx, &fmtrpc.LoginRequest{
					Password: c.oldPwd,
				})
				if c.expectedErr == nil {
					if err != c.expectedErr {
						t.Fatalf("Unexpected error when calling Login: %v", err)
					}
				} else {
					st, ok := status.FromError(err)
					if !ok {
						t.Errorf("Unexpected error format when calling Login")
					}
					if st.Message() != c.expectedErr.Error() {
						t.Errorf("Unexpected error when calling Login: %v", err)
					}
				}
			case setPwdCmd:
				_, err := client.SetPassword(ctx, &fmtrpc.SetPwdRequest{
					Password: c.oldPwd,
				})
				if c.expectedErr == nil {
					if err != c.expectedErr {
						t.Fatalf("Unexpected error when calling SetPassword: %v", err)
					}
				} else {
					st, ok := status.FromError(err)
					if !ok {
						t.Errorf("Unexpected error format when calling SetPassword")
					}
					if st.Message() != c.expectedErr.Error() {
						t.Errorf("Unexpected error when calling SetPassword: %v", err)
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
	caseName    string
	bakeMacReq  *fmtrpc.BakeMacaroonRequest
	expectedErr error
	callback    func(string, grpc.DialOption) (*grpc.ClientConn, error)
}

var (
	bakeMacTestCases = []bakeMacCases{
		{"bake-mac-empty-permissions", &fmtrpc.BakeMacaroonRequest{}, ErrEmptyPermissionsList, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
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
		}, ErrInvalidMacEntity, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
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
		}, ErrInvalidMacEntity, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
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
		}, nil, func(macHex string, tlsDialOpt grpc.DialOption) (*grpc.ClientConn, error) {
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
	unAuthedFmtClient := fmtrpc.NewFmtClient(conn)
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
			_, err := fmtClient.BakeMacaroon(ctx, c.bakeMacReq)
			if c.expectedErr == nil {
				if err != c.expectedErr {
					t.Fatalf("Unexpected error when calling BakeMacaroon: %v", err)
				}
			} else {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Unexpected error format when calling BakeMacaroon")
				}
				if st.Message() != c.expectedErr.Error() {
					t.Errorf("Unexpected error when calling BakeMacaroon: %v", err)
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
	t.Run("set-temperature", func(t *testing.T) {
		_, err := unAuthedFmtClient.SetTemperature(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling SetTemperature")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling SetTemperature: %v", err)
		}
	})
	t.Run("set-pressure", func(t *testing.T) {
		_, err := unAuthedFmtClient.SetPressure(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling SetPressure")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling SetPressure: %v", err)
		}
	})
	t.Run("start-recording", func(t *testing.T) {
		_, err := unAuthedFmtClient.StartRecording(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling StartRecording")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling StartRecording: %v", err)
		}
	})
	t.Run("stop-recording", func(t *testing.T) {
		_, err := unAuthedFmtClient.StopRecording(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling StopRecording")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling StopRecording: %v", err)
		}
	})
	t.Run("subscribe-datastream", func(t *testing.T) {
		_, err := unAuthedFmtClient.SubscribeDataStream(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling SubscribeDataStream")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling SubscribeDataStream: %v", err)
		}
	})
	t.Run("load-testplan", func(t *testing.T) {
		_, err := unAuthedFmtClient.LoadTestPlan(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling LoadTestPlan")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling LoadTestPlan: %v", err)
		}
	})
	t.Run("start-testplan", func(t *testing.T) {
		_, err := unAuthedFmtClient.StartTestPlan(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling StartTestPlan")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling StartTestPlan: %v", err)
		}
	})
	t.Run("stop-testplan", func(t *testing.T) {
		_, err := unAuthedFmtClient.StopTestPlan(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling StopTestPlan")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling StopTestPlan: %v", err)
		}
	})
	t.Run("insert-roi", func(t *testing.T) {
		_, err := unAuthedFmtClient.InsertROIMarker(ctx, &proto.Empty{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Unexpected error format when calling InsertROIMarker")
		}
		if st.Message() != ErrDeprecatedAction.Error() {
			t.Errorf("Unexpected error when calling InsertROIMarker: %v", err)
		}
	})
	shutdownInterceptor.RequestShutdown()
	wg.Wait()
}

// TestPluginAPI tests the PluginAPI service over bufnet gRPC with macaroons
func TestPluginAPI(t *testing.T) {
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
	plugClient := fmtrpc.NewPluginAPIClient(authConn)
	t.Run("start record-plugin not registered", func(t *testing.T) {
		_, err := plugClient.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "invalid-plugin-name"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error format when calling StartRecord: %v", err)
		}
		if st.Message() != plugins.ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
	})
	t.Run("start record-plugin invalid state", func(t *testing.T) {
		_, err := plugClient.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "test-timeout-plugin"})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		time.Sleep(20 * time.Second)
		_, err = plugClient.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "test-timeout-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error format when calling StartRecord: %v", err)
		}
		if st.Message() != plugins.ErrPluginNotReady.Error() {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
	})
	t.Run("start record-already recording", func(t *testing.T) {
		_, err := plugClient.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "test-rng-plugin"})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		_, err = plugClient.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "test-rng-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error format when calling StartRecord: %v", err)
		}
		if st.Message() != errors.ErrAlreadyRecording.Error() {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
	})
	t.Run("start record-plugin error", func(t *testing.T) {
		_, err := plugClient.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "test-error-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartRecord gRPC method: %v", err)
		}
		if st.Message() != "I don't wanna" {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
	})
	t.Run("start record-plugin timeout", func(t *testing.T) {
		pluginState, err := plugClient.GetPlugin(ctx, &fmtrpc.PluginRequest{Name: "test-timeout-plugin"})
		if err != nil {
			t.Errorf("Unexpected error when calling GetPlugin: %v", err)
		}
		if pluginState.State != fmtrpc.PluginState_UNRESPONSIVE {
			t.Errorf("Plugin in unexpected state: %v", pluginState.State)
		}
	})
	t.Run("start plugin-unregistered plugin", func(t *testing.T) {
		_, err := plugClient.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: "invalid-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartPlugin gRPC method: %v", err)
		}
		if st.Message() != plugins.ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("start plugin-already started", func(t *testing.T) {
		_, err := plugClient.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: "test-error-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartPlugin gRPC method: %v", err)
		}
		if st.Message() != errors.ErrServiceAlreadyStarted.Error() {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("start plugin-valid", func(t *testing.T) {
		_, err := plugClient.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: "test-rng-plugin"})
		if err != nil {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
		_, err = plugClient.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: "test-rng-plugin"})
		if err != nil {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("stop plugin-unregistered plugin", func(t *testing.T) {
		_, err := plugClient.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: "invalid-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StopPlugin gRPC method: %v", err)
		}
		if st.Message() != plugins.ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
	t.Run("stop plugin-valid", func(t *testing.T) {
		_, err := plugClient.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: "test-rng-plugin"})
		if err != nil {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
	t.Run("stop plugin-already stopped", func(t *testing.T) {
		_, err := plugClient.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: "test-rng-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StopPlugin gRPC method: %v", err)
		}
		if st.Message() != errors.ErrServiceAlreadyStopped.Error() {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
	t.Run("command-unregistered plugin", func(t *testing.T) {
		stream, _ := plugClient.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: "unregistered-plugin"})
		_, err := stream.Recv()
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		if st.Message() != plugins.ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling Command: %v", err)
		}
	})
	t.Run("Command-empty frame", func(t *testing.T) {
		stream, _ := plugClient.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: "test-ctrl-plugin", Frame: &proto.Frame{}})
		_, err := stream.Recv()
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		if st.Message() != "invalid frame type" {
			t.Errorf("Unexpected error when calling Command: %v", err)
		}
	})
	t.Run("command-single return frame", func(t *testing.T) {
		type examplePayload struct {
			Command string `json:"command"`
			Arg     string `json:"arg"`
		}
		pay := examplePayload{
			Command: "echo",
			Arg:     "foo bar",
		}
		p, err := json.Marshal(pay)
		if err != nil {
			t.Errorf("Could not format request payload: %v", err)
		}
		stream, err := plugClient.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: "test-ctrl-plugin", Frame: &proto.Frame{
			Source:    "client",
			Type:      "application/json",
			Timestamp: time.Now().UnixMilli(),
			Payload:   p,
		}})
		if err != nil {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		resp, err := stream.Recv()
		if err != nil {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		if resp == nil {
			t.Errorf("Unexpected response from Command gRPC method: response is nil")
		} else {
			if string(resp.Payload) != "foo bar" {
				t.Errorf("Unexpected response from Command gRPC method: %v", string(resp.Payload))
			}
			t.Logf("Echo response: %v", string(resp.Payload))
		}
	})
	shutdownInterceptor.RequestShutdown()
	wg.Wait()
}
