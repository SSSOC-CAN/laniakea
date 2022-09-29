package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginName            = "test-ctrl-plugin"
	pluginVersion         = "0.1.0"
	laniVersionConstraint = ">= 1.0.0"
	ErrInvalidFrameType   = errors.New("invalid frame type")
	ErrInvalidCommand     = errors.New("invalid command")
	ErrEmptyRequest       = errors.New("request arg is nil")
)

type ControllerExample struct {
	sdk.ControllerBase
	quitChan chan struct{}
	sync.WaitGroup
}

type CtrlExampleCommand struct {
	Command string `json:"command"`
	Arg     string `json:"arg"`
}

// Compile time check to ensure ControllerExample satisfies the Datasource interface
var _ sdk.Controller = (*ControllerExample)(nil)

// Implements the Controller interface funciton Command
func (e *ControllerExample) Command(req *proto.Frame) (chan *proto.Frame, error) {
	if req == nil {
		return nil, ErrEmptyRequest
	}
	frameChan := make(chan *proto.Frame)
	switch req.Type {
	case "application/json":
		var cmd CtrlExampleCommand
		err := json.Unmarshal(req.Payload, &cmd)
		if err != nil {
			return nil, err
		}
		switch cmd.Command {
		case "echo":
			e.Add(1)
			go func() {
				defer e.Done()
				defer close(frameChan)
				time.Sleep(1 * time.Second)
				frameChan <- &proto.Frame{
					Source:    pluginName,
					Type:      "application/string",
					Timestamp: time.Now().UnixMilli(),
					Payload:   []byte("test"),
				}
				frameChan <- &proto.Frame{
					Source:    pluginName,
					Type:      "application/string",
					Timestamp: time.Now().UnixMilli(),
					Payload:   []byte(string(cmd.Arg)),
				}
			}()
			return frameChan, nil
		case "rng":
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
		default:
			return nil, ErrInvalidCommand
		}
	default:
		return nil, ErrInvalidFrameType
	}
}

// Implements the Datasource interface funciton Stop
func (e *ControllerExample) Stop() error {
	close(e.quitChan)
	e.Wait()
	return nil
}

func main() {
	impl := &ControllerExample{quitChan: make(chan struct{})}
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
