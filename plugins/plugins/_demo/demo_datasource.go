package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginName            = "demo-datasource-plugin"
	pluginVersion         = "1.0.0"
	laniVersionConstraint = ">= 0.2.0"
)

type DemoDatasource struct {
	sdk.DatasourceBase
	quitChan chan struct{}
	sync.WaitGroup
}

type DemoPayload struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type DemoFrame struct {
	Data []DemoPayload `json:"data"`
}

// Compile time check to ensure DemoDatasource satisfies the Datasource interface
var _ sdk.Datasource = (*DemoDatasource)(nil)

// Implements the Datasource interface funciton StartRecord
func (e *DemoDatasource) StartRecord() (chan *proto.Frame, error) {
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
				data := []DemoPayload{}
				df := DemoFrame{}
				for i := 0; i < 96; i++ {
					v := (rand.Float64()*1 + 25)
					n := fmt.Sprintf("Temperature: %v", i+1)
					data = append(data, DemoPayload{Name: n, Value: v})
				}
				df.Data = data[:]
				// transform to json string
				b, err := json.Marshal(&df)
				if err != nil {
					log.Println(err)
					return
				}
				frameChan <- &proto.Frame{
					Source:    pluginName,
					Type:      "application/json",
					Timestamp: time.Now().UnixMilli(),
					Payload:   b,
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return frameChan, nil
}

// Implements the Datasource interface funciton StopRecord
func (e *DemoDatasource) StopRecord() error {
	e.quitChan <- struct{}{}
	return nil
}

// Implements the Datasource interface funciton Stop
func (e *DemoDatasource) Stop() error {
	close(e.quitChan)
	e.Wait()
	return nil
}

func main() {
	impl := &DemoDatasource{quitChan: make(chan struct{})}
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
