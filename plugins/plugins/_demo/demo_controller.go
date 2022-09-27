package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
)

var (
	pluginName            = "demo-controller-plugin"
	pluginVersion         = "1.0.0"
	laniVersionConstraint = ">= 1.0.0"
)

type DemoController struct {
	sdk.ControllerBase
	quitChan     chan struct{}
	tempSetPoint float64
	avgTemp      float64
	recording    int32 // used atomically
	sync.WaitGroup
	sync.RWMutex
}

type DemoPayload struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type DemoFrame struct {
	Data []DemoPayload `json:"data"`
}

type CtrlCommand struct {
	Command string            `json:"command"`
	Args    map[string]string `json:"args"`
}

// Compile time check to ensure DemoController satisfies the Controller interface
var _ sdk.Controller = (*DemoController)(nil)

// Implements the Controller interface funciton Command
func (e *DemoController) Command(req *proto.Frame) (chan *proto.Frame, error) {
	if req == nil {
		return nil, errors.New("request arg is nil")
	}
	frameChan := make(chan *proto.Frame)
	switch req.Type {
	case "application/json":
		var cmd CtrlCommand
		err := json.Unmarshal(req.Payload, &cmd)
		if err != nil {
			return nil, err
		}
		switch cmd.Command {
		case "set-temperature":
			if atomic.LoadInt32(&e.recording) == 1 {
				if setPtStr, ok := cmd.Args["temp_set_point"]; ok {
					if v, err := strconv.ParseFloat(setPtStr, 64); err == nil {
						e.Add(1)
						go func() {
							defer e.Done()
							defer close(frameChan)
							time.Sleep(1 * time.Second)
							e.RLock()
							frameChan <- &proto.Frame{
								Source:    pluginName,
								Type:      "application/json",
								Timestamp: time.Now().UnixMilli(),
								Payload:   []byte(fmt.Sprintf(`{ "average_temperature": %f, "current_set_point": %f}`, e.avgTemp, e.tempSetPoint)),
							}
							e.RUnlock()
							e.Lock()
							e.tempSetPoint = v
							e.Unlock()
							e.RLock()
							frameChan <- &proto.Frame{
								Source:    pluginName,
								Type:      "application/json",
								Timestamp: time.Now().UnixMilli(),
								Payload:   []byte(fmt.Sprintf(`{ "average_temperature": %f, "current_set_point": %f}`, e.avgTemp, e.tempSetPoint)),
							}
							e.RUnlock()
						}()
						return frameChan, nil
					}
					return nil, errors.New("not recording")
				}
			}
			return nil, errors.New("invalid temperature set point type")
		case "start-record":
			if ok := atomic.CompareAndSwapInt32(&e.recording, 0, 1); !ok {
				return nil, errors.New("already recording")
			}
			e.Add(1)
			go func() {
				defer e.Done()
				defer close(frameChan)
				defer log.Println("Exiting rng data loop")
				time.Sleep(1 * time.Second) // sleep for a second while laniakea sets up the plugin
				for {
					select {
					case <-e.quitChan:
						return
					default:
						data := []DemoPayload{}
						df := DemoFrame{}
						var cumTemp float64
						for i := 0; i < 96; i++ {
							e.RLock()
							v := (rand.Float64()*1 + e.tempSetPoint)
							e.RUnlock()
							n := fmt.Sprintf("Temperature: %v", i+1)
							data = append(data, DemoPayload{Name: n, Value: v})
							cumTemp += v
						}
						e.Lock()
						e.avgTemp = cumTemp / float64(len(data))
						e.Unlock()
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
		default:
			return nil, errors.New("invalid command")
		}
	default:
		return nil, errors.New("invalid frame type")
	}
}

// Implements the Datasource interface funciton Stop
func (e *DemoController) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&e.recording, 1, 0); !ok {
		return errors.New("already stopped recording")
	}
	close(e.quitChan)
	e.Wait()
	return nil
}

func main() {
	impl := &DemoController{quitChan: make(chan struct{}), tempSetPoint: 25.0}
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
