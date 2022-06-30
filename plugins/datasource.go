/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/30
*/

package plugins

import (
	"context"
	"errors"
	"io"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
)

type DatasourceGRPCClient struct{ client fmtrpc.DatasourceClient }

type DatasourceGRPCServer struct {
	fmtrpc.UnimplementedDatasourceServer
	Impl Datasource
}

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

// StartRecord implements the Datasource gRPC server interface
func (s *DatasourceGRPCServer) StartRecord(_ *fmtrpc.Empty, stream fmtrpc.Datasource_StartRecordServer) error {
	frameChan, err := s.Impl.StartRecord()
	if err != nil {
		return err
	}
	for {
		select {
		case frame := <-frameChan:
			if err := stream.Send(frame); err != nil {
				return err
			}
		case <-stream.Context().Done():
			if errors.Is(stream.Context().Err(), context.Canceled) {
				return nil
			}
			return stream.Context().Err()
		}
	}
}

// StopRecord implements the Datasource gRPC server interface
func (s *DatasourceGRPCServer) StopRecord(ctx context.Context, _ *fmtrpc.Empty) (*fmtrpc.Empty, error) {
	err := s.Impl.StopRecord()
	return &fmtrpc.Empty{}, err
}
