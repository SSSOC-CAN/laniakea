package main

import (
	"errors"
	"sync"

	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginName            = "test-ctrl-error-plugin"
	pluginVersion         = "0.2.0"
	laniVersionConstraint = ">= 0.2.0"
	ErrCommand            = errors.New("nope not gonna do it")
	ErrStop               = errors.New("can't stop, won't stop")
)

type ControllerExample struct {
	sdk.ControllerBase
	quitChan chan struct{}
	sync.WaitGroup
}

// Compile time check to ensure ControllerExample satisfies the Datasource interface
var _ sdk.Controller = (*ControllerExample)(nil)

// Implements the Datasource interface funciton Command
func (e *ControllerExample) Command(req *proto.Frame) (chan *proto.Frame, error) {
	return nil, ErrCommand
}

// Implements the Datasource interface funciton Stop
func (e *ControllerExample) Stop() error {
	return ErrStop
}

func main() {
	impl := &ControllerExample{}
	impl.SetPluginVersion(pluginVersion)              // set the plugin version before serving
	impl.SetVersionConstraints(laniVersionConstraint) // set required laniakea version before serving
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginName: &sdk.ControllerPlugin{Impl: impl},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
