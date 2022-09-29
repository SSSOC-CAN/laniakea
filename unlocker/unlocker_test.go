/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20
*/

package unlocker

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/SSSOC-CAN/laniakea/cert"
	"github.com/SSSOC-CAN/laniakea/kvdb"
	"github.com/SSSOC-CAN/laniakea/lanirpc"
	"github.com/SSSOC-CAN/laniakea/macaroons"
	"github.com/SSSOC-CAN/laniakea/utils"
	"github.com/btcsuite/btcwallet/snacl"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	bufSize             = 1 * 1024 * 1024
	lis                 *bufconn.Listener
	defaultTestPassword = []byte("test")
)

func initService(t *testing.T) (*UnlockerService, func(), string) {
	// make temporary directory
	tempDir, err := ioutil.TempDir("", "unlocker-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	// make blank macaroon file
	adminMacPath := path.Join(tempDir, "admin.macaroon")
	_, err = os.OpenFile(adminMacPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		t.Fatalf("Could not open/create file: %v", err)
	}
	// initialize DB
	db, err := kvdb.NewDB(path.Join(tempDir, "macaroon.db"))
	if err != nil {
		t.Fatalf("Could not create macaroon.db in temporary directory: %v", err)
	}
	cleanUp := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}
	// initialize Unlocker Service
	unlockerService, err := InitUnlockerService(db, []string{adminMacPath})
	if err != nil {
		t.Errorf("Could not initialize unlocker service: %v", err)
	}
	return unlockerService, cleanUp, tempDir
}

// TestInitUnlockerService tests if initializing the Unlocker Service is possible
func TestInitUnlockerService(t *testing.T) {
	_, cleanUp, _ := initService(t)
	defer cleanUp()
}

// TestRegisterWithRestProxy tests if we can register the unlocker service with the REST proxy
func TestRegisterWithRestProxy(t *testing.T) {
	unlockerService, cleanUp, tempDir := initService(t)
	defer cleanUp()
	// prereqs for RegisterWithRestProxy
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
	err = unlockerService.RegisterWithRestProxy(ctx, mux, restDialOpts, restProxyDest)
	if err != nil {
		t.Errorf("Could not register with Rest Proxy: %v", err)
	}
}

// initGrpcServer initializes the gRPC server and registers it with the RPC server
func initGrpcServer(t *testing.T) (func(), string) {
	unlockerService, cleanUp, tempDir := initService(t)
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	_ = unlockerService.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	shutdownChan := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-unlockerService.PasswordMsgs:
				t.Log(msg.Err)
			case <-shutdownChan:
				return
			}
		}
	}()
	bigCleanUp := func() {
		grpcServer.Stop()
		close(shutdownChan)
		cleanUp()
	}
	return bigCleanUp, tempDir
}

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// TestCommands tests all the RPC commands for the Unlocker Service
func TestCommands(t *testing.T) {
	cleanUp, tempDir := initGrpcServer(t)
	defer cleanUp()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Errorf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := lanirpc.NewUnlockerClient(conn)
	// Login before setting password
	t.Run("lanicli login pre-pwd-set", func(t *testing.T) {
		resp, err := client.Login(ctx, &lanirpc.LoginRequest{
			Password: defaultTestPassword,
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrPasswordNotSet.Error() {
			t.Errorf("Unexpected error when calling Login command: %v", st.Message())
		}
		t.Log(resp)
	})
	// ChangePassword before setting password
	t.Run("lanicli changepassword pre-pwd-set", func(t *testing.T) {
		resp, err := client.ChangePassword(ctx, &lanirpc.ChangePwdRequest{
			CurrentPassword:    defaultTestPassword,
			NewPassword:        []byte("password123"),
			NewMacaroonRootKey: false,
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrPasswordNotSet.Error() {
			t.Errorf("Unexpected error when calling ChangePassword command: %v", st.Message())
		}
		t.Log(resp)
	})
	// SetPassword
	t.Run("lanicli setpassword", func(t *testing.T) {
		resp, err := client.SetPassword(ctx, &lanirpc.SetPwdRequest{
			Password: defaultTestPassword,
		})
		if err != nil {
			t.Errorf("Unexpected error when calling SetPassword RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// SetPassword invalid
	t.Run("lanicli setpassword invalid", func(t *testing.T) {
		resp, err := client.SetPassword(ctx, &lanirpc.SetPwdRequest{
			Password: defaultTestPassword,
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrPasswordAlreadySet.Error() {
			t.Errorf("Unexpected error when calling SetPassword command: %v", st.Message())
		}
		t.Log(resp)
	})
	// Login wrong password
	t.Run("lanicli login wrong-pwd", func(t *testing.T) {
		resp, err := client.Login(ctx, &lanirpc.LoginRequest{
			Password: []byte("wrongpassword"),
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrWrongPassword.Error() {
			t.Errorf("Unexpected error when calling Login command: %v", st.Message())
		}
		t.Log(resp)
	})
	// Login
	t.Run("lanicli login", func(t *testing.T) {
		resp, err := client.Login(ctx, &lanirpc.LoginRequest{
			Password: defaultTestPassword,
		})
		if err != nil {
			t.Errorf("Unexpected error when calling Login RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// ChangePassword don't remove macaroons
	t.Run("lanicli changepassword keep-macs", func(t *testing.T) {
		resp, err := client.ChangePassword(ctx, &lanirpc.ChangePwdRequest{
			CurrentPassword:    defaultTestPassword,
			NewPassword:        []byte("password123"),
			NewMacaroonRootKey: false,
		})
		// err = macaroon encryption key not found
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != macaroons.ErrEncKeyNotFound.Error() {
			t.Errorf("Unexpected error when calling ChangePassword command: %v", st.Message())
		}
		t.Log(resp)
		if !utils.FileExists(path.Join(tempDir, "admin.macaroon")) {
			t.Error("File should exist but does not")
		}
	})
	// Login again
	t.Run("lanicli login wrong password", func(t *testing.T) {
		resp, err := client.Login(ctx, &lanirpc.LoginRequest{
			Password: defaultTestPassword,
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrWrongPassword.Error() {
			t.Errorf("Unexpected error when calling Login command: %v", st.Message())
		}
		t.Log(resp)
	})
	// Login again
	t.Run("lanicli login again", func(t *testing.T) {
		resp, err := client.Login(ctx, &lanirpc.LoginRequest{
			Password: []byte("password123"),
		})
		if err != nil {
			t.Errorf("Unexpected error when calling Login RPC endpoint: %v", err)
		}
		t.Log(resp)
	})
	// Changepassword wrong password
	t.Run("lanicli changepassword wrong password", func(t *testing.T) {
		resp, err := client.ChangePassword(ctx, &lanirpc.ChangePwdRequest{
			CurrentPassword:    defaultTestPassword,
			NewPassword:        []byte("password123"),
			NewMacaroonRootKey: false,
		})
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != ErrWrongPassword.Error() {
			t.Errorf("Unexpected error when calling ChangePassword command: %v", st.Message())
		}
		t.Log(resp)
	})
	// Changepassword erase macs
	t.Run("lanicli changepassword yeet-macs", func(t *testing.T) {
		resp, err := client.ChangePassword(ctx, &lanirpc.ChangePwdRequest{
			CurrentPassword:    []byte("password123"),
			NewPassword:        defaultTestPassword,
			NewMacaroonRootKey: true,
		})
		// err = invalid password
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Error is not a gRPC status error")
		}
		if st.Message() != snacl.ErrInvalidPassword.Error() {
			t.Errorf("Unexpected error when calling ChangePassword command: %v", st.Message())
		}
		t.Log(resp)
		if utils.FileExists(path.Join(tempDir, "admin.macaroon")) {
			t.Error("File should not exist but does")
		}
	})
}
