package plugins

import (
	"path/filepath"

	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
)

const (
	ErrInvalidPluginExec  = bg.Error("invalid plugin executable")
	ErrPluginExecNotFound = bg.Error("plugin executable not found")
	ErrInvalidPluginType  = bg.Error("invalid plugin type")
)

var (
	defaultPluginTimeout     int64 = 30
	defaultPluginMaxTimeouts int64 = 3
)

// ValidatePluginConfig takes a PluginConfig and validates the parameters
func ValidatePluginConfig(cfg *fmtrpc.PluginConfig, pluginDir string) error {
	// checks that the plugin has a valid name and executable
	if !utils.ValidatePluginName(cfg.Name) {
		return errors.ErrInvalidPluginName
	} else if !utils.ValidatePluginExec(cfg.ExecName) {
		return ErrInvalidPluginExec
	}
	// check that the plugin type is either a datasource or controller
	if cfg.Type != DATASOURCE_STR && cfg.Type != CONTROLLER_STR {
		return ErrInvalidPluginType
	}
	if !utils.FileExists(filepath.Join(pluginDir, cfg.ExecName)) {
		return ErrPluginExecNotFound
	}
	// if timeout or maxtimeout or version are 0 set to default values
	if cfg.Timeout == int64(0) {
		cfg.Timeout = defaultPluginTimeout
	}
	if cfg.MaxTimeouts == int64(0) {
		cfg.MaxTimeouts = defaultPluginMaxTimeouts
	}
	return nil
}
