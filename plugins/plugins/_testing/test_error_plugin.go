package main

import (
	"errors"
	"sync"

	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginName            = "test-error-plugin"
	pluginVersion         = "0.4.0"
	laniVersionConstraint = ">= 1.0.0"
	ErrStartRecord        = errors.New("I don't wanna")
	ErrStopRecord         = errors.New("no thank you")
	ErrStop               = errors.New("can't stop, won't stop")
)

type DatasourceExample struct {
	sdk.DatasourceBase
	quitChan chan struct{}
	sync.WaitGroup
}

// Compile time check to ensure DatasourceExample satisfies the Datasource interface
var _ sdk.Datasource = (*DatasourceExample)(nil)

// Implements the Datasource interface funciton StartRecord
func (e *DatasourceExample) StartRecord() (chan *proto.Frame, error) {
	return nil, ErrStartRecord
}

// Implements the Datasource interface funciton StopRecord
func (e *DatasourceExample) StopRecord() error {
	return ErrStopRecord
}

// Implements the Datasource interface funciton Stop
func (e *DatasourceExample) Stop() error {
	return ErrStop
}

func main() {
	impl := &DatasourceExample{}
	impl.SetPluginVersion(pluginVersion)              // set the plugin version before serving
	impl.SetVersionConstraints(laniVersionConstraint) // set required laniakea version before serving
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: sdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginName: &sdk.DatasourcePlugin{Impl: impl},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
