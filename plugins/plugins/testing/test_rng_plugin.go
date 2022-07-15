package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginName            = "test-rng-plugin"
	pluginVersion         = "0.1.0"
	laniVersionConstraint = ">= 0.2.0"
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
	frameChan := make(chan *proto.Frame)
	e.Add(1)
	go func() {
		defer e.Done()
		defer close(frameChan)
		time.Sleep(1 * time.Second) // sleep for a second while laniakea sets up the plugin
		for {
			select {
			case <-e.quitChan:
				return
			default:
				rnd := make([]byte, 8)
				_, err := rand.Read(rnd)
				if err != nil {
					log.Println(err)
				}
				frameChan <- &proto.Frame{
					Source:    pluginName,
					Type:      "application/json",
					Timestamp: time.Now().UnixMilli(),
					Payload:   rnd,
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return frameChan, nil
}

// Implements the Datasource interface funciton StopRecord
func (e *DatasourceExample) StopRecord() error {
	e.quitChan <- struct{}{}
	return nil
}

// Implements the Datasource interface funciton Stop
func (e *DatasourceExample) Stop() error {
	close(e.quitChan)
	e.Wait()
	return nil
}

func main() {
	impl := &DatasourceExample{quitChan: make(chan struct{})}
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
