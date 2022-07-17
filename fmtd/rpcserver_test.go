/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package fmtd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/utils"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	bufSize            = 1 * 1024 * 1024
	lis                *bufconn.Listener
	testMacPermissions = []*fmtrpc.MacaroonPermission{
		&fmtrpc.MacaroonPermission{
			Entity: "uri",
			Action: "/fmtrpc.Fmt/AdminTest",
		},
	}
	defaultTestPassword = []byte("test")
)

// initRpcServer is a helper function to initialize the RPC server struct
func initRpcServer(t *testing.T) (*RpcServer, func()) {
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		t.Fatalf("Could not initialize interceptor: %v", err)
	}
	defer shutdownInterceptor.Close()
	cfg, err := InitConfig(true)
	if err != nil {
		t.Fatalf("Could not initialize config: %v", err)
	}
	cfg.GrpcPort = 3567 // override config to not interfere with any currently running nodes
	log, err := InitLogger(&cfg)
	if err != nil {
		t.Fatalf("Could not initialize logger: %v", err)
	}
	//shutdownInterceptor.Logger = &log
	rpcServer, err := NewRpcServer(shutdownInterceptor, &cfg, &log)
	if err != nil {
		t.Fatalf("Could not initialize RPC server: %v", err)
	}
	cleanUp := func() {
		_ = rpcServer.Listener.Close()
		shutdownInterceptor.RequestShutdown()
	}
	return rpcServer, cleanUp
}

// TestStartStopRpcServer tests if we can initialize, start and stop a new RPC server
func TestStartStopRpcServer(t *testing.T) {
	rpcServer, cleanUp := initRpcServer(t)
	defer cleanUp()
	t.Run("Start RPC server", func(t *testing.T) {
		err := rpcServer.Start()
		if err != nil {
			t.Fatalf("Could not start RPC server: %v", err)
		}
	})
	t.Run("Start RPC server invalid", func(t *testing.T) {
		err := rpcServer.Start()
		if err != errors.ErrServiceAlreadyStarted {
			t.Errorf("Unexpected error when starting RPC server: %v", err)
		}
	})
	t.Run("Stop RPC server", func(t *testing.T) {
		err := rpcServer.Stop()
		if err != nil {
			t.Fatalf("Could not stop RPC server: %v", err)
		}
	})
	t.Run("Stop RPC server invalid", func(t *testing.T) {
		err := rpcServer.Stop()
		if err != errors.ErrServiceAlreadyStopped {
			t.Errorf("Unexpected error when stopping RPC server: %v", err)
		}
	})
}

// TestRegisterWithRestProxy tests if we can successfully register with the REST proxy
func TestRegisterWithRestProxy(t *testing.T) {
	time.Sleep(1 * time.Second) // gives enough time to shutdownInterceptor to shutdown
	rpcServer, cleanUp := initRpcServer(t)
	defer cleanUp()
	// context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Proxy Serve Mux
	customMarshalerOption := proxy.WithMarshalerOption(
		proxy.MIMEWildcard, &proxy.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
		},
	)
	mux := proxy.NewServeMux(customMarshalerOption)
	// TLS REST config
	_, restDialOpts, _, tlsCleanUp, err := cert.GetTLSConfig(
		rpcServer.cfg.TLSCertPath,
		rpcServer.cfg.TLSKeyPath,
		rpcServer.cfg.ExtraIPAddr,
	)
	if err != nil {
		t.Fatalf("Could not get TLS Config: %v", err)
	}
	defer tlsCleanUp()
	restProxyDestNet, err := utils.NormalizeAddresses([]string{fmt.Sprintf("localhost:%d", rpcServer.cfg.GrpcPort)}, strconv.FormatInt(rpcServer.cfg.GrpcPort, 10), net.ResolveTCPAddr)
	if err != nil {
		t.Fatalf("Could not normalize address: %v", err)
	}
	restProxyDest := restProxyDestNet[0].String()
	// RegisterWithRestProxy
	err = rpcServer.RegisterWithRestProxy(ctx, mux, restDialOpts, restProxyDest)
	if err != nil {
		t.Fatalf("Could not register with REST proxy: %v", err)
	}
}

// initGrpcServer initializes the gRPC server and registers it with the RPC server
func initGrpcServer(t *testing.T) func() {
	rpcServer, cleanUp := initRpcServer(t)
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	_ = rpcServer.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	bigCleanUp := func() {
		grpcServer.Stop()
		cleanUp()
	}
	return bigCleanUp
}

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// TestCommands tests the all RPC endpoint for the Fmt Service
func TestCommands(t *testing.T) {
	time.Sleep(1 * time.Second)
	cleanUp := initGrpcServer(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewFmtClient(conn)
	// Test Command
	t.Run("fmtcli test", func(t *testing.T) {
		resp, err := client.TestCommand(ctx, &fmtrpc.TestRequest{})
		if err != nil {
			t.Fatalf("Unexpected error when calling TestCommand RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Admin Test
	t.Run("fmtcli admin-test", func(t *testing.T) {
		resp, err := client.AdminTest(ctx, &fmtrpc.AdminTestRequest{})
		if err != nil {
			t.Fatalf("Unexpected error when calling AdminTest RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Bake Macaroon
	t.Run("fmtcli bake-macaroon", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(0),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: testMacPermissions,
		})
		if err == nil {
			t.Fatalf("Expected error when calling Bake Macaroon command")
		}
		t.Log(resp)
	})
	// Stop Daemon
	t.Run("fmtcli stop", func(t *testing.T) {
		resp, err := client.StopDaemon(ctx, &fmtrpc.StopRequest{})
		if err != nil {
			t.Fatalf("Unexpected error when calling StopDaemon Command RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
}

// initGrpcServer initializes the gRPC server and registers it with the RPC server
func initGrpcServerMac(t *testing.T) func() {
	rpcServer, cleanUp := initRpcServer(t)
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	_ = rpcServer.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	// temporary directory for macaroon db
	tempDir, err := ioutil.TempDir("", "macaroon-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	db, err := kvdb.NewDB(path.Join(tempDir, "macaroon.db"))
	if err != nil {
		t.Fatalf("Could not initialize kvdb: %v", err)
	}
	// client for init func
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewFmtClient(conn)
	t.Run("Macaroon Service is Nil", func(t *testing.T) {
		_, err := client.BakeMacaroon(context.Background(), &fmtrpc.BakeMacaroonRequest{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != errors.ErrMacSvcNil.Error() {
			t.Errorf("Unexpected error when calling Bake macaroon command: %v", st.Message())
		}
	})
	// macaroon service
	macaroonService, err := macaroons.InitService(*db, "fmtd", zerolog.New(os.Stderr).With().Timestamp().Logger(), []string{}, []string{})
	if err != nil {
		t.Errorf("Could not initialize macaroon service: %v", err)
	}
	err = macaroonService.CreateUnlock(&defaultTestPassword)
	if err != nil {
		t.Errorf("Could not unlock macaroon store: %v", err)
	}
	rpcServer.AddMacaroonService(macaroonService)
	t.Run("gRPC Middleware is Nil", func(t *testing.T) {
		_, err := client.BakeMacaroon(context.Background(), &fmtrpc.BakeMacaroonRequest{})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrGRPCMiddlewareNil.Error() {
			t.Errorf("Unexpected error when calling Bake macaroon command: %v", st.Message())
		}
	})
	// gRPC Middleware
	grpcInterceptor := intercept.NewGrpcInterceptor(rpcServer.SubLogger, true)
	err = grpcInterceptor.AddPermissions(MainGrpcServerPermissions())
	if err != nil {
		t.Errorf("Could not add permissions to gRPC middleware: %v", err)
	}
	rpcServer.AddGrpcInterceptor(grpcInterceptor)
	grpcInterceptor.SetRPCActive()
	bigCleanUp := func() {
		grpcServer.Stop()
		cleanUp()
	}
	return bigCleanUp
}

// TestBakeMacaroon tests the bake macaroon command
func TestBakeMacaroon(t *testing.T) {
	time.Sleep(1 * time.Second)
	cleanUp := initGrpcServerMac(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewFmtClient(conn)
	// With empty permissions array
	t.Run("no permissions", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(0),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: make([]*fmtrpc.MacaroonPermission, 0),
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrEmptyPermissionsList.Error() {
			t.Fatalf("Unexpected error when calling Bake Macaroon command: %v", st.Message())
		}
		t.Log(resp)
	})
	// With invalid permission entity
	t.Run("invalid entity", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(0),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "invalid",
					Action: "action",
				},
			},
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrInvalidMacEntity.Error() {
			t.Errorf("Unexpected error when calling Bake Macaroon command: %v", st.Message())
		}
		t.Log(resp)
	})
	// With invalid permission action
	t.Run("invalid action", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(0),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "fmtd",
					Action: "invalid",
				},
			},
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrInvalidMacAction.Error() {
			t.Errorf("Unexpected error when calling Bake Macaroon command: %v", st.Message())
		}
		t.Log(resp)
	})
	// With invalid permission action, uri entity
	t.Run("invalid action custom uri", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(0),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "uri",
					Action: "invalid",
				},
			},
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrInvalidMacAction.Error() {
			t.Errorf("Unexpected error when calling Bake Macaroon command: %v", st.Message())
		}
		t.Log(resp)
	})
	// Valid call custom uri
	t.Run("valid call custom uri", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(60),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "uri",
					Action: "/fmtrpc.Fmt/AdminTest",
				},
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error when calling Bake Macaroon command: %v", err)
		}
		t.Log(resp)
	})
	// valid call fmtd
	t.Run("valid call uri", func(t *testing.T) {
		resp, err := client.BakeMacaroon(ctx, &fmtrpc.BakeMacaroonRequest{
			Timeout:     int64(60),
			TimeoutType: fmtrpc.TimeoutType_SECOND,
			Permissions: []*fmtrpc.MacaroonPermission{
				&fmtrpc.MacaroonPermission{
					Entity: "fmtd",
					Action: "read",
				},
			},
		})
		if err != nil {
			t.Fatalf("Unexpected error when calling Bake Macaroon command: %v", err)
		}
		t.Log(resp)
	})
}
