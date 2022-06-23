package plugins

import (
	"context"
	"io"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
)

type DatasourceGRPCClient struct{ client fmtrpc.DatasourceClient }

// StartRecord implements the Datasource interface method StartRecord
// TODO:SSSOCPaulCote - timeout context
func (c *DatasourceGRPCClient) StartRecord() (chan *fmtrpc.Frame, error) {
	stream, err := c.client.StartRecord(context.Background(), &fmtrpc.Empty{})
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

// StopRecord implements the Datasource interface method StopRecord
// TODO:SSSOCPaulCote - timeout context
func (c *DatasourceGRPCClient) StopRecord() error {
	_, err := c.client.StopRecord(context.Background(), &fmtrpc.Empty{})
	if err != nil {
		return err
	}
	return nil
}
