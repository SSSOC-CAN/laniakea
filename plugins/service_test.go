/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/07/08
*/

package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
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
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
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

// initTestingSetup
func initTestingSetup(t *testing.T, ctx context.Context, cfgs []*fmtrpc.PluginConfig) (*PluginManager, fmtrpc.PluginAPIClient, func()) {
	lis = bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	pluginDir := getPluginDir(t)
	pluginManager := initPluginManager(t, pluginDir, cfgs)
	_ = pluginManager.RegisterWithGrpcServer(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	err := pluginManager.Start(ctx)
	if err != nil {
		t.Fatalf("Could not start plugin manager: %v", err)
	}
	// Now set up client connection
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	client := fmtrpc.NewPluginAPIClient(conn)
	cleanUp := func() {
		grpcServer.Stop()
		pluginManager.Stop()
		conn.Close()
	}
	return pluginManager, client, cleanUp
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
	rngPluginCfg      = &fmtrpc.PluginConfig{
		Name:        rngPluginName,
		Type:        DATASOURCE_STR,
		ExecName:    "test_rng_plugin",
		Timeout:     defaultPluginTimeout,
		MaxTimeouts: defaultPluginMaxTimeouts,
	}
	timeoutPluginCfg = &fmtrpc.PluginConfig{
		Name:        timeoutPluginName,
		Type:        DATASOURCE_STR,
		ExecName:    "test_timeout_plugin",
		Timeout:     15,
		MaxTimeouts: defaultPluginMaxTimeouts,
	}
	errPluginCfg = &fmtrpc.PluginConfig{
		Name:        errorPluginName,
		Type:        DATASOURCE_STR,
		ExecName:    "test_error_plugin",
		Timeout:     15,
		MaxTimeouts: defaultPluginMaxTimeouts,
	}
	datasourceCfgs = []*fmtrpc.PluginConfig{
		rngPluginCfg,
		timeoutPluginCfg,
		errPluginCfg,
	}
)

// bufDialer is a callback used for the gRPC client
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// TestDatasourcePlugin tests a dummy datasource plugin to ensure PluginAPI functionality
func TestDatasourcePlugin(t *testing.T) {
	ctx := context.Background()
	pluginManager, client, cleanUp := initTestingSetup(t, ctx, datasourceCfgs)
	defer cleanUp()
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
		if plug := pluginManager.pluginRegistry[errorPluginName]; plug.getState() != fmtrpc.PluginState_UNKNOWN {
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
		if plug.getState() != fmtrpc.PluginState_UNRESPONSIVE {
			t.Errorf("Plugin in unexpected state after timing out: %v", plug.getState())
		}
		// Wait 10 more seconds and plugin manager should restart the plugin
		time.Sleep(10 * time.Second)
		if plug.getState() != fmtrpc.PluginState_READY {
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
		if plug.getState() != fmtrpc.PluginState_KILLED {
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
				for j := 0; j < rand.Intn(10-5)+5; j++ {
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
	t.Run("start record-timeout while streaming", func(t *testing.T) {
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		stream, err := client.Subscribe(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling Susbcribe: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				resp, err := stream.Recv()
				if resp == nil || err == io.EOF {
					return
				}
				if err != nil {
					t.Errorf("Unexpected error occured receiving stream from Command: %v", err)
					return
				}
			}
		}()
		time.Sleep(5 * time.Second)
		plug := pluginManager.pluginRegistry[rngPluginName]
		plug.setUnresponsive()
		wg.Wait()
		time.Sleep(3 * time.Second)
		_, err = client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		stream, err = client.Subscribe(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling Susbcribe: %v", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				resp, err := stream.Recv()
				if resp == nil || err == io.EOF {
					return
				}
				if err != nil {
					t.Errorf("Unexpected error occured receiving stream from Command: %v", err)
					return
				}
			}
		}()
		time.Sleep(5 * time.Second)
		_, err = client.StopRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StopRecord: %v", err)
		}
		wg.Wait()
	})
}

var (
	addPluginCfgs = []*fmtrpc.PluginConfig{
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
	}
	invalidPlugNameCfg = &fmtrpc.PluginConfig{
		Name: "not.asd$vaslidplugin",
	}
	invalidPlugExecCfgs = []*fmtrpc.PluginConfig{
		{
			Name:     "test-plugin",
			ExecName: "nasdnf$.exe",
		},
		{
			Name:     "test-plugin",
			ExecName: "nasdnf_asQfd123.ex_e",
		},
	}
	invalidPlugTypeCfg = &fmtrpc.PluginConfig{
		Name:     "test-plugin",
		ExecName: "test_plugin",
		Type:     "not-a-valid-plugin-type",
	}
	invalidPlugNoFileCfg = &fmtrpc.PluginConfig{
		Name:     "test-plugin",
		ExecName: "test_plugin",
		Type:     DATASOURCE_STR,
	}
	invalidPlugAlreadyExistsCfg = &fmtrpc.PluginConfig{
		Name:     rngPluginName,
		ExecName: "test_rng_plugin",
		Type:     DATASOURCE_STR,
	}
	validAddPlugin = &fmtrpc.PluginConfig{
		Name:        errorPluginName,
		Type:        DATASOURCE_STR,
		ExecName:    "test_error_plugin",
		Timeout:     15,
		MaxTimeouts: defaultPluginMaxTimeouts,
	}
)

// TestPluginAPI tests the AddPlugin, GetPlugin and ListPlugins gRPC methods
func TestPluginAPI(t *testing.T) {
	ctx := context.Background()
	_, client, cleanUp := initTestingSetup(t, ctx, addPluginCfgs)
	defer cleanUp()
	t.Run("add plugin-invalid plugin name", func(t *testing.T) {
		_, err := client.AddPlugin(ctx, invalidPlugNameCfg)
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from AddPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrInvalidPluginName.Error() {
			t.Errorf("Unexpected error when calling AddPlugin: %v", err)
		}
	})
	t.Run("add plugin-invalid plugin executables", func(t *testing.T) {
		for _, cfg := range invalidPlugExecCfgs {
			_, err := client.AddPlugin(ctx, cfg)
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from AddPlugin gRPC method: %v", err)
			}
			if st.Message() != ErrInvalidPluginExec.Error() {
				t.Errorf("Unexpected error when calling AddPlugin: %v", err)
			}
		}
	})
	t.Run("add plugin-invalid plugin type", func(t *testing.T) {
		_, err := client.AddPlugin(ctx, invalidPlugTypeCfg)
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from AddPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrInvalidPluginType.Error() {
			t.Errorf("Unexpected error when calling AddPlugin: %v", err)
		}
	})
	t.Run("add plugin-exec not found", func(t *testing.T) {
		_, err := client.AddPlugin(ctx, invalidPlugNoFileCfg)
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from AddPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrPluginExecNotFound.Error() {
			t.Errorf("Unexpected error when calling AddPlugin: %v", err)
		}
	})
	t.Run("add plugin-plugin already registered", func(t *testing.T) {
		_, err := client.AddPlugin(ctx, invalidPlugAlreadyExistsCfg)
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from AddPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrDuplicatePluginName.Error() {
			t.Errorf("Unexpected error when calling AddPlugin: %v", err)
		}
	})
	t.Run("add plugin-valid", func(t *testing.T) {
		resp, err := client.AddPlugin(ctx, validAddPlugin)
		if err != nil {
			t.Errorf("Unexpected error when calling AddPlugin: %v", err)
		}
		plugType, err := getPluginCodeFromType(validAddPlugin.Type)
		if err != nil {
			t.Errorf("Unexpected error when calling getPluginCodeFromType: %v", err)
		}
		if resp.Name != validAddPlugin.Name {
			t.Errorf("Unexpected plugin name mismatch: expected %v got %v", validAddPlugin.Name, resp.Name)
		} else if resp.Type != plugType {
			t.Errorf("Unexpected plugin type mismatch: expected %v got %v", plugType, resp.Type)
		} else if resp.State != fmtrpc.PluginState_READY {
			t.Errorf("Unexpected plugin state mismatch: expected %v got %v", fmtrpc.PluginState_READY, resp.State)
		}
	})
}

var (
	getPluginCfgs         = datasourceCfgs
	getPluginUnregistered = &fmtrpc.PluginRequest{
		Name: "not-a-registered-plugin",
	}
	getPluginValid = &fmtrpc.PluginRequest{
		Name: errorPluginName,
	}
)

// TestGetPlugins will test the GetPlugin and ListPlugins gRPC methods
func TestGetPlugins(t *testing.T) {
	ctx := context.Background()
	_, client, cleanUp := initTestingSetup(t, ctx, getPluginCfgs)
	defer cleanUp()
	t.Run("get plugin-unregistered plugin", func(t *testing.T) {
		_, err := client.GetPlugin(ctx, getPluginUnregistered)
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error type coming from GetPlugin gRPC method: %v", err)
		}
		if st.Message() != ErrUnregsiteredPlugin.Error() {
			t.Errorf("Unexpected error when calling GetPlugin: %v", err)
		}
	})
	t.Run("get plugin-valid", func(t *testing.T) {
		resp, err := client.GetPlugin(ctx, getPluginValid)
		if err != nil {
			t.Errorf("Unexpected error when calling GetPlugin: %v", err)
		}
		plugType, err := getPluginCodeFromType(errPluginCfg.Type)
		if err != nil {
			t.Errorf("Unexpected error when calling getPluginCodeFromType: %v", err)
		}
		if resp.Name != errPluginCfg.Name {
			t.Errorf("Unexpected plugin name mismatch: expected %v got %v", errPluginCfg.Name, resp.Name)
		}
		if resp.Type != plugType {
			t.Errorf("Unexpected plugin type mismatch: expected %v got %v", plugType, resp.Type)
		}
	})
	t.Run("list plugins-valid", func(t *testing.T) {
		resp, err := client.ListPlugins(ctx, &proto.Empty{})
		if err != nil {
			t.Errorf("Unexpected error when calling ListPlugins: %v", err)
		}
		plugCfgMap := map[string]*fmtrpc.PluginConfig{
			rngPluginName:     rngPluginCfg,
			errorPluginName:   errPluginCfg,
			timeoutPluginName: timeoutPluginCfg,
		}
		for _, plug := range resp.Plugins {
			cfg, ok := plugCfgMap[plug.Name]
			if !ok {
				t.Errorf("Unexpected plugin list mismatch: %s plugin not given to plugin manager", plug.Name)
			}
			plugType, err := getPluginCodeFromType(cfg.Type)
			if err != nil {
				t.Errorf("Unexpected error when calling getPluginCodeFromType: %v", err)
			}
			if plug.Name != cfg.Name {
				t.Errorf("Unexpected plugin name mismatch: expected %v got %v", cfg.Name, plug.Name)
			}
			if plug.Type != plugType {
				t.Errorf("Unexpected plugin type mismatch: expected %v got %v", plugType, plug.Type)
			}
		}
	})
}

var (
	ctrlPluginName = "test-ctrl-plugin"
	ctrlPluginCfg  = &fmtrpc.PluginConfig{
		Name:        ctrlPluginName,
		Type:        CONTROLLER_STR,
		ExecName:    "test_ctrl_plugin",
		Timeout:     10,
		MaxTimeouts: 3,
	}
	controllerCfgs = []*fmtrpc.PluginConfig{
		ctrlPluginCfg,
	}
)

// TestControllerPlugin will use dummy controller plugins to test some controller plugin functionality
func TestControllerPlugin(t *testing.T) {
	ctx := context.Background()
	pluginManager, client, cleanUp := initTestingSetup(t, ctx, controllerCfgs)
	defer cleanUp()
	t.Run("command-unregistered plugin", func(t *testing.T) {
		stream, err := client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: "unregistered-plugin"})
		if err == nil {
			_, err = stream.Recv()
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
			}
			if st.Message() != ErrUnregsiteredPlugin.Error() {
				t.Errorf("Unexpected error when calling Command: %v", err)
			}
		} else {
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
			}
			if st.Message() != ErrUnregsiteredPlugin.Error() {
				t.Errorf("Unexpected error when calling Command: %v", err)
			}
		}
	})
	t.Run("command-invalid plugin state", func(t *testing.T) {
		plug := pluginManager.pluginRegistry[ctrlPluginName]
		plug.setBusy()
		defer plug.setReady()
		stream, err := client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: ctrlPluginName})
		if err == nil {
			_, err = stream.Recv()
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
			}
			if st.Message() != ErrPluginNotReady.Error() {
				t.Errorf("Unexpected error when calling Command: %v", err)
			}
		} else {
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
			}
			if st.Message() != ErrPluginNotReady.Error() {
				t.Errorf("Unexpected error when calling Command: %v", err)
			}
		}
	})
	t.Run("command-empty frame", func(t *testing.T) {
		stream, err := client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: ctrlPluginName, Frame: &proto.Frame{}})
		if err == nil {
			_, err = stream.Recv()
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
			}
			if st.Message() != "invalid frame type" {
				t.Errorf("Unexpected error when calling Command: %v", err)
			}
		} else {
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
			}
			if st.Message() != "invalid frame type" {
				t.Errorf("Unexpected error when calling Command: %v", err)
			}
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
		stream, err := client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: ctrlPluginName, Frame: &proto.Frame{
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
	t.Run("command-stream of frames", func(t *testing.T) {
		type examplePayload struct {
			Command string `json:"command"`
			Arg     string `json:"arg"`
		}
		pay := examplePayload{
			Command: "rng",
		}
		p, err := json.Marshal(pay)
		if err != nil {
			t.Errorf("Could not format request payload: %v", err)
		}
		stream, err := client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: ctrlPluginName, Frame: &proto.Frame{
			Source:    "client",
			Type:      "application/json",
			Timestamp: time.Now().UnixMilli(),
			Payload:   p,
		}})
		if err != nil {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				resp, err := stream.Recv()
				if resp == nil || err == io.EOF {
					return
				}
				if err != nil {
					t.Errorf("Unexpected error occured receiving stream from Command: %v", err)
					return
				}
			}
		}()
		time.Sleep(5 * time.Second)
		_, err = client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: ctrlPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
		wg.Wait()
	})
	t.Run("command-timeout while streaming", func(t *testing.T) {
		type examplePayload struct {
			Command string `json:"command"`
			Arg     string `json:"arg"`
		}
		pay := examplePayload{
			Command: "rng",
		}
		p, err := json.Marshal(pay)
		if err != nil {
			t.Errorf("Could not format request payload: %v", err)
		}
		stream, err := client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: ctrlPluginName, Frame: &proto.Frame{
			Source:    "client",
			Type:      "application/json",
			Timestamp: time.Now().UnixMilli(),
			Payload:   p,
		}})
		if err != nil {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				resp, err := stream.Recv()
				if resp == nil || err == io.EOF {
					return
				}
				if err != nil {
					t.Errorf("Unexpected error occured receiving stream from Command: %v", err)
					return
				}
			}
		}()
		time.Sleep(5 * time.Second)
		plug := pluginManager.pluginRegistry[ctrlPluginName]
		plug.setUnresponsive()
		wg.Wait()
		time.Sleep(10 * time.Second)
		stream, err = client.Command(ctx, &fmtrpc.ControllerPluginRequest{Name: ctrlPluginName, Frame: &proto.Frame{
			Source:    "client",
			Type:      "application/json",
			Timestamp: time.Now().UnixMilli(),
			Payload:   p,
		}})
		if err != nil {
			t.Errorf("Unexpected error type coming from Command gRPC method: %v", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				resp, err := stream.Recv()
				if resp == nil || err == io.EOF {
					return
				}
				if err != nil {
					t.Errorf("Unexpected error occured receiving stream from Command: %v", err)
					return
				}
			}
		}()
		time.Sleep(5 * time.Second)
		_, err = client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: ctrlPluginName})
		if err != nil {
			t.Errorf("Unexpected error occured when calling StopPlugin gRPC function: %v", err)
		}
		wg.Wait()
	})
}

var (
	subStateCfgs = datasourceCfgs
)

// TestSubscribePluginState tests the SubscribePluginState gRPC method
func TestSubscribePluginState(t *testing.T) {
	ctx := context.Background()
	_, client, cleanUp := initTestingSetup(t, ctx, subStateCfgs)
	defer cleanUp()
	t.Run("individual-unregistered plugin", func(t *testing.T) {
		stream, err := client.SubscribePluginState(ctx, &fmtrpc.PluginRequest{Name: "not-a-registered-plugin"})
		if err == nil {
			_, err := stream.Recv()
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error format when calling SusbcribePluginState: expected gRPC status format, got %v", err)
			}
			if st.Message() != ErrUnregsiteredPlugin.Error() {
				t.Errorf("Unexpected errore when calling SusbscribePluginState: %v", err)
			}
		} else {
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Unexpected error format when calling SusbcribePluginState: expected gRPC status format, got %v", err)
			}
			if st.Message() != ErrUnregsiteredPlugin.Error() {
				t.Errorf("Unexpected errore when calling SusbscribePluginState: %v", err)
			}
		}
	})
	t.Run("individual-busy", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			stream, err := client.SubscribePluginState(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
			if err != nil {
				t.Errorf("Unexpected error when calling SusbcribePluginState: %v", err)
			}
			stateUpdate, err := stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_READY {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
			stateUpdate, err = stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_BUSY {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
		}()
		time.Sleep(time.Second)
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: rngPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		wg.Wait()
	})
	t.Run("individual-timeout", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			stream, err := client.SubscribePluginState(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
			if err != nil {
				t.Errorf("Unexpected error when calling SusbcribePluginState: %v", err)
			}
			stateUpdate, err := stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_BUSY {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
			stateUpdate, err = stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_READY {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
			stateUpdate, err = stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_UNRESPONSIVE {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
		}()
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: timeoutPluginName})
		if err != nil {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		wg.Wait()
	})
	t.Run("individual-unknown and stopped", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			stream, err := client.SubscribePluginState(ctx, &fmtrpc.PluginRequest{Name: errorPluginName})
			if err != nil {
				t.Errorf("Unexpected error when calling SusbcribePluginState: %v", err)
			}
			stateUpdate, err := stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_UNKNOWN {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
			stateUpdate, err = stream.Recv()
			if err != nil {
				t.Errorf("Unexpected error when reading SusbcribePluginState stream: %v", err)
			}
			if stateUpdate.State != fmtrpc.PluginState_STOPPED {
				t.Errorf("Unexpected state received from SusbcribePluginState stream: %v", stateUpdate.State)
			}
		}()
		_, err := client.StartRecord(ctx, &fmtrpc.PluginRequest{Name: errorPluginName})
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error format when calling StartRecord: expected gRPC status format, got %v", err)
		}
		if st.Message() != "I don't wanna" {
			t.Errorf("Unexpected error when calling StartRecord: %v", err)
		}
		time.Sleep(time.Second)
		_, err = client.StopPlugin(ctx, &fmtrpc.PluginRequest{Name: errorPluginName})
		st, ok = status.FromError(err)
		if !ok {
			t.Errorf("Unexpected error format when calling StopPlugin: expected gRPC status format, got %v", err)
		}
		if st.Message() != "can't stop, won't stop" {
			t.Errorf("Unexpected error when calling StopPlugin: %v", err)
		}
		wg.Wait()
	})
}
