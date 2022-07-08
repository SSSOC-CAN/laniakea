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
	"strconv"
	"testing"

	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
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

// TestStartPluginManager tests the Start, plugin manager function
func TestStartPluginManager(t *testing.T) {
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
