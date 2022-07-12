/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/07/08
*/

package plugins

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	rootDirRegexp   = `/(?:fmtd)$`
	pluginDirRegexp = `/(?:fmtd)/(?:plugins)$`
)

// initPluginManager will init the plugin manager
func initPluginManager(t *testing.T, pluginDir string, cfgs []*fmtrpc.PluginConfig) *PluginManager {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	return NewPluginManager(pluginDir, cfgs, logger)
}

var (
	restProxyCfgs = []*fmtrpc.PluginConfig{
		{
			Name:        "test-plugin",
			Type:        DATASOURCE_STR,
			ExecName:    "test-plugin-exec.exe",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
	}
)

//TestRegisterWithRestProxy tests the RegisterWithRestProxy function
func TestRegisterWithRestProxy(t *testing.T) {
	// make temporary directory
	tempDir, err := ioutil.TempDir("", "plugin-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
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
		filepath.Join(tempDir, "tls.cert"),
		filepath.Join(tempDir, "tls.key"),
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
	// initialize plugin manager
	pluginManager := initPluginManager(t, tempDir, restProxyCfgs)
	err = pluginManager.RegisterWithRestProxy(context.Background(), mux, restDialOpts, restProxyDest)
	if err != nil {
		t.Errorf("Could not register with Rest Proxy: %v", err)
	}
}

var (
	startPluginDir     = "plugins/testing"
	startDuplicateCfgs = []*fmtrpc.PluginConfig{
		{
			Name:        "test-rng-plugin",
			Type:        DATASOURCE_STR,
			ExecName:    "test_rng_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
		{
			Name:        "test-rng-plugin",
			Type:        CONTROLLER_STR,
			ExecName:    "test_rng_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
	}
	startInvalidTypeCfgs = []*fmtrpc.PluginConfig{
		{
			Name:        "test-rng-plugin",
			Type:        DATASOURCE_STR,
			ExecName:    "test_rng_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
		{
			Name:        "invalid-plugin",
			Type:        "not-a-valid-type",
			ExecName:    "invalid_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
	}
	startInvalidVersionCfgs = []*fmtrpc.PluginConfig{
		{
			Name:        "test-rng-plugin",
			Type:        DATASOURCE_STR,
			ExecName:    "test_rng_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
		{
			Name:        "test-version-plugin",
			Type:        DATASOURCE_STR,
			ExecName:    "test_version_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
	}
	startValidCfgs = []*fmtrpc.PluginConfig{
		{
			Name:        "test-rng-plugin",
			Type:        DATASOURCE_STR,
			ExecName:    "test_rng_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
	}
)

func getPluginDir(t *testing.T) string {
	// first get cwd, determine if root or plugins directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get current working directory: %v", err)
	}
	// root dir
	var pluginDir string
	match, err := regexp.MatchString(rootDirRegexp, dir)
	if err != nil {
		t.Fatalf("could not run regular expression check: %v", err)
	}
	if !match {
		// plugins dir
		match, err = regexp.MatchString(pluginDirRegexp, dir)
		if err != nil {
			t.Fatalf("could not run regular expression check: %v", err)
		}
		if !match {
			t.Fatalf("Running test from an unexpected location, please run from root directory or within plugins package directory")
		} else {
			pluginDir = filepath.Join(dir, startPluginDir)
		}
	} else {
		pluginDir = filepath.Join(dir, "plugins", startPluginDir)
	}
	return pluginDir
}

// TestStartPluginManager tests the Start, plugin manager function
func TestStartPluginManager(t *testing.T) {
	pluginDir := getPluginDir(t)
	t.Run("duplicate plugin names", func(t *testing.T) {
		pluginManager := initPluginManager(t, pluginDir, startDuplicateCfgs)
		err := pluginManager.Start(context.Background())
		if err != ErrDuplicatePluginName {
			t.Errorf("Unexpected error when calling Start: %v", err)
		}
		defer pluginManager.Stop()
	})
	t.Run("invalid plugin type", func(t *testing.T) {
		pluginManager := initPluginManager(t, pluginDir, startInvalidTypeCfgs)
		err := pluginManager.Start(context.Background())
		if err != ErrInvalidPluginType {
			t.Errorf("Unexpected error when calling start: %v", err)
		}
		defer pluginManager.Stop()
	})
	t.Run("incompatible versions", func(t *testing.T) {
		pluginManager := initPluginManager(t, pluginDir, startInvalidVersionCfgs)
		err := pluginManager.Start(context.Background())
		if err != sdk.ErrLaniakeaVersionMismatch {
			t.Errorf("Unexpected error when calling start: %v", err)
		}
		defer pluginManager.Stop()
	})
	t.Run("valid", func(t *testing.T) {
		pluginManager := initPluginManager(t, pluginDir, startValidCfgs)
		err := pluginManager.Start(context.Background())
		if err != nil {
			t.Errorf("Unexpected error when calling start: %v", err)
		}
		defer pluginManager.Stop()
	})
}

var (
	bufSize           = 1 * 1024 * 1024
	lis               *bufconn.Listener
	rngPluginName     = "test-rng-plugin"
	timeoutPluginName = "test-timeout-plugin"
	errorPluginName   = "test-error-plugin"
	datasourceCfgs    = []*fmtrpc.PluginConfig{
		{
			Name:        rngPluginName,
			Type:        DATASOURCE_STR,
			ExecName:    "test_rng_plugin",
			Timeout:     defaultPluginTimeout,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
		{
			Name:        timeoutPluginName,
			Type:        DATASOURCE_STR,
			ExecName:    "test_timeout_plugin",
			Timeout:     15,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
		{
			Name:        errorPluginName,
			Type:        DATASOURCE_STR,
			ExecName:    "test_error_plugin",
			Timeout:     15,
			MaxTimeouts: defaultPluginMaxTimeouts,
		},
	}
)

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// TestDatasourcePlugin tests a dummy datasource plugin to ensure PluginAPI functionality
func TestDatasourcePlugin(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	pluginDir := getPluginDir(t)
	pluginManager := initPluginManager(t, pluginDir, datasourceCfgs)
	_ = pluginManager.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()
	ctx := context.Background()
	err := pluginManager.Start(ctx)
	if err != nil {
		t.Fatalf("Could not start plugin manager: %v", err)
	}
	defer pluginManager.Stop()

	// Now set up client connection

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewPluginAPIClient(conn)
	t.Run("start record-plugin not registered", func(t *testing.T) {
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: "invalid-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartRecord gRPC method: %v", err)
		}
		if st.Message() != ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling StartRecord: %v", st.Message())
		}
	})
	t.Run("start record-plugin invalid state", func(t *testing.T) {
		rngPlugin, ok := pluginManager.pluginRegistry[rngPluginName]
		if !ok {
			t.Fatalf("Unexpected error: plugin %s should be in plugin registry but isn't", rngPluginName)
		}
		rngPlugin.setBusy()
		defer rngPlugin.setReady()
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartRecord gRPC method: %v", err)
		}
		if st.Message() != ErrPluginNotReady.Error() {
			t.Errorf("Unexpected error when calling StartRecord: %v", st.Message())
		}
	})
	t.Run("start record-already recording", func(t *testing.T) {
		rngPlugin, ok := pluginManager.pluginRegistry[rngPluginName]
		if !ok {
			t.Fatalf("Unexpected error: plugin %s should be in plugin registry but isn't", rngPluginName)
		}
		rngPlugin.setRecording()
		defer rngPlugin.setNotRecording()
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartRecord gRPC method: %v", err)
		}
		if st.Message() != errors.ErrAlreadyRecording.Error() {
			t.Errorf("Unexpected error when calling StartRecord: %v", st.Message())
		}
	})
	t.Run("start record-plugin error", func(t *testing.T) {
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: errorPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartRecord gRPC method: %v", err)
		}
		if st.Message() != "I don't wanna" {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		// check if the plugin state is Unknown
		if plug := pluginManager.pluginRegistry[errorPluginName]; plug.getState() != fmtrpc.Plugin_UNKNOWN {
			t.Errorf("Unexpected plugin state after error when StartRecord: %v", plug.getState())
		}
	})
	t.Run("start record-plugin timeout", func(t *testing.T) {
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		plug := pluginManager.pluginRegistry[timeoutPluginName]
		// Sleep until the plugin has timed out
		time.Sleep(time.Duration(plug.cfg.Timeout+6) * time.Second)
		if plug.getState() != fmtrpc.Plugin_UNRESPONSIVE {
			t.Errorf("Plugin in unexpected state after timing out: %v", plug.getState())
		}
		// Wait 10 more seconds and plugin manager should restart the plugin
		time.Sleep(10 * time.Second)
		if plug.getState() != fmtrpc.Plugin_READY {
			t.Errorf("Plugin in unexpected state after timeout restart: %v", plug.getState())
		} else if plug.getTimeoutCount() != 1 {
			t.Errorf("Plugin timeout counter unexpected value: %v", plug.getTimeoutCount())
		}
		// now we make the plugin timeout 2 more times to confirm it gets killed
		for i := 0; i < int(plug.cfg.MaxTimeouts); i++ {
			_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
			if err != nil {
				t.Errorf("Unexpected error when calling StartRecord: %v", err)
			}
			time.Sleep(time.Duration(plug.cfg.Timeout+16) * time.Second)
		}
		if plug.getState() != fmtrpc.Plugin_KILLED {
			t.Errorf("Plugin in unexpected state after reaching max timeouts: %v timeout counter: %v", plug.getState(), plug.getTimeoutCount())
		} else if plug.getTimeoutCount() != int(plug.cfg.MaxTimeouts) {
			t.Errorf("Plugin timeout counter unexpected value: %v", plug.getTimeoutCount())
		}
	})
	t.Run("start record-rng plugin", func(t *testing.T) {
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		defer func() {
			_, _ = client.StopRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		}()
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				stream, err := client.Subscribe(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
				if err != nil {
					t.Errorf("Unexpected error when calling Susbcribe: %v", err)
				}
				for j := 0; j < 10; j++ {
					_, err := stream.Recv()
					if err != nil {
						if err.Error() == PluginEOF {
							break
						} else {
							t.Errorf("Unexpected error when reading data stream: %v", err)
						}
					}
				}
				return
			}()
		}
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		t.Logf("Alloc=%vMiB TotalAlloc=%vMiB Sys=%vMiB NumGC=%v", m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
		wg.Wait()
	})
	t.Run("start plugin-unregistered plugin", func(t *testing.T) {
		_, err := client.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: "invalid-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("start plugin-invalid plugin state", func(t *testing.T) {
		_, err := client.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: errorPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartPlugin gRPC method: %v", err)
		}
		if st.Message() != errors.ErrServiceAlreadyStarted.Error() {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("start plugin-invalid plugin type", func(t *testing.T) {
		plug := pluginManager.pluginRegistry[errorPluginName]
		plug.cfg.Type = "invalid-plugin-type"
		plug.setKilled()
		defer func() {
			plug.cfg.Type = DATASOURCE_STR
			plug.setReady()
		}()
		_, err := client.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: errorPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrInvalidPluginType.Error() {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("start plugin-valid", func(t *testing.T) {
		_, err := client.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartPlugin: %v", err)
		}
	})
	t.Run("stop plugin-unregistered plugin", func(t *testing.T) {
		_, err := client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: "invalid-plugin"})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StartPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
	t.Run("stop plugin-valid", func(t *testing.T) {
		_, err := client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
	t.Run("stop plugin-plugin already stopped", func(t *testing.T) {
		_, err := client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StopPlugin gRPC method: %v", err)
		}
		if st.Message() != errors.ErrServiceAlreadyStopped.Error() {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
	t.Run("stop plugin-plugin not started", func(t *testing.T) {
		plug := pluginManager.pluginRegistry[timeoutPluginName]
		plug.setReady()
		defer plug.setKilled()
		_, err := client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from StopPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrPluginNotStarted.Error() {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
	})
}

var (
	invalidPluginCfg = &fmtrpc.PluginConfig{
		Name:        "invalid-plugin",
		Type:        "not-a-valid-type",
		ExecName:    "invalid_plugin",
		Timeout:     defaultPluginTimeout,
		MaxTimeouts: defaultPluginMaxTimeouts,
	}
)

// TestPluginAPI tests the AddPlugin, GetPlugin and ListPlugins gRPC methods
func TestPluginAPI(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	pluginDir := getPluginDir(t)
	pluginManager := initPluginManager(t, pluginDir, datasourceCfgs)
	_ = pluginManager.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()
	ctx := context.Background()
	err := pluginManager.Start(ctx)
	if err != nil {
		t.Fatalf("Could not start plugin manager: %v", err)
	}
	defer pluginManager.Stop()

	// Now set up client connection

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := fmtrpc.NewPluginAPIClient(conn)
	t.Run("")
}
