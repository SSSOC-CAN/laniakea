package plugins

import (
	"testing"

	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
)

var (
	invalidPluginNameCfgs = []*fmtrpc.PluginConfig{
		{
			Name: "bing bong",
		},
		{
			Name: "BiGy0sHi.exe",
		},
		{
			Name: "%$%sad",
		},
	}
	invalidPluginExecCfgs = []*fmtrpc.PluginConfig{
		{
			Name:     "valid-plugin-name",
			ExecName: "love you.exe",
		},
		{
			Name:     "valid-plugin-name",
			ExecName: "love-you.ex_e",
		},
		{
			Name:     "valid-plugin-name",
			ExecName: "love-you.ex-e",
		},
		{
			Name:     "valid-plugin-name",
			ExecName: "love-you.ex#$",
		},
	}
	invalidPluginTypeCfg = &fmtrpc.PluginConfig{
		Name:     "valid-plugin-name",
		ExecName: "Valid-executable.exe",
		Type:     "not-a-valid-plugin-type",
	}
	invalidPluginExecFile = &fmtrpc.PluginConfig{
		Name:     "valid-plugin-name",
		ExecName: "Valid-executable.exe",
		Type:     DATASOURCE_STR,
	}
)

// TestValidatePluginConfig tests the ValidatePluginConfig function
func TestValidatePluginConfig(t *testing.T) {
	pluginDir := getPluginDir(t)
	t.Run("validate cfg-invalid plugin name", func(t *testing.T) {
		for _, cfg := range invalidPluginNameCfgs {
			err := ValidatePluginConfig(cfg, pluginDir)
			if err != errors.ErrInvalidPluginName {
				t.Errorf("Unexpected error when calling ValidatePluginConfig: %v", err)
			}
		}
	})
	t.Run("validate cfg-invalid plugin exec", func(t *testing.T) {
		for _, cfg := range invalidPluginExecCfgs {
			err := ValidatePluginConfig(cfg, pluginDir)
			if err != ErrInvalidPluginExec {
				t.Errorf("Unexpected error when calling ValidatePluginConfig: %v", err)
			}
		}
	})
	t.Run("validate cfg-invalid plugin type", func(t *testing.T) {
		err := ValidatePluginConfig(invalidPluginTypeCfg, pluginDir)
		if err != ErrInvalidPluginType {
			t.Errorf("Unexpected error when calling ValidatePluginConfig: %v", err)
		}
	})
	t.Run("validate cfg-invalid plugin exec file", func(t *testing.T) {
		err := ValidatePluginConfig(invalidPluginExecFile, pluginDir)
		if err != ErrPluginExecNotFound {
			t.Errorf("Unexpected error when calling ValidatePluginConfig: %v", err)
		}
	})
}
