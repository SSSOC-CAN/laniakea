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

type ControllerGRPCClient struct{ client fmtrpc.ControllerClient }

type ControllerGRPCServer struct {
	fmtrpc.UnimplementedControllerServer
	Impl Controller
}

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

// PushVersion implements the Controller interface method PushVersion
func (c *ControllerGRPCClient) PushVersion(versionNumber string) error {
	_, err := c.client.PushVersion(context.Background(), &fmtrpc.VersionNumber{Version: versionNumber})
	if err != nil {
		return err
	}
	return nil
}

// GetVersion implements the Controller interface method GetVersion
func (c *ControllerGRPCClient) GetVersion() string {
	resp, err := c.client.GetVersion(context.Background(), &fmtrpc.Empty{})
	if err != nil {
		return ""
	}
	return resp.Version
}

// Stop implements the Controller gRPC server interface
func (s *ControllerGRPCServer) Stop(ctx context.Context, _ *fmtrpc.Empty) (*fmtrpc.Empty, error) {
	err := s.Impl.Stop()
	return &fmtrpc.Empty{}, err
}

// Command implements the Controller gRPC server interface
func (s *ControllerGRPCServer) Command(req *fmtrpc.Frame, stream fmtrpc.Controller_CommandServer) error {
	frameChan, err := s.Impl.Command(req)
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

// PushVersion implements the Controller gRPC server interface
func (s *ControllerGRPCServer) PushVersion(ctx context.Context, req *fmtrpc.VersionNumber) (*fmtrpc.Empty, error) {
	err := s.Impl.PushVersion(req.Version)
	return &fmtrpc.Empty{}, err
}

// GetVersion implements the Controller gRPC server interface
func (s *ControllerGRPCServer) GetVersion(ctx context.Context, _ *fmtrpc.Empty) (*fmtrpc.VersionNumber, error) {
	v := s.Impl.GetVersion()
	return &fmtrpc.VersionNumber{Version: v}, nil
}
