package plugins

import (
	"context"
	"io"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
)

type ControllerGRPCClient struct{ client fmtrpc.ControllerClient }

// Stop implements the Controller interface method Stop
// TODO:SSSOCPaulCote - add timeout context
func (c *ControllerGRPCClient) Stop() error {
	_, err := c.client.Stop(context.Background(), &fmtrpc.Empty{})
	if err != nil {
		return err
	}
	return nil
}

// Command implements the Controller interface method Command
// TODO:SSSOCPaulCote - add timeout context
func (c *ControllerGRPCClient) Command(f *fmtrpc.Frame) (chan *fmtrpc.Frame, error) {
	stream, err := c.client.Command(context.Background(), f)
	if err != nil {
		return nil, err
	}
	frameChan := make(chan *fmtrpc.Frame)
	go func() {
		defer close(frameChan)
		for {
			frame, err := stream.Recv()
			if frame == nil || err == io.EOF {
				return
			}
			if err != nil {
				break
			}
			frameChan <- frame
		}
	}()
	return frameChan, nil
}
