/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/30
*/

package plugins

import (
	"context"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Datasource interface describes an interface for plugins which will only produce streams of data
type Datasource interface {
	StartRecord() (chan *fmtrpc.Frame, error)
	StopRecord() error
	Stop() error
	PushVersion(versionNumber string) error // This method pushes the version of Laniakea to the plugin. Plugin can then specify a minimum version of laniakea to run properly
	GetVersion() string                     // This method gets the version number from the plugin. Needed if plugins rely on other plugins and specific versions are needed
}

// Controller interface describes an interface for plugins which produce data but also act as controllers
type Controller interface {
	Stop() error
	Command(*fmtrpc.Frame) (chan *fmtrpc.Frame, error)
	PushVersion(versionNumber string) error // This method pushes the version of Laniakea to the plugin. Plugin can then specify a minimum version of laniakea to run properly
	GetVersion() string                     // This method gets the version number from the plugin. Needed if plugins rely on other plugins and specific versions are needed
}

type DatasourcePlugin struct {
	plugin.Plugin
	Impl Datasource
}

type ControllerPlugin struct {
	plugin.Plugin
	Impl Controller
}

// GRPCServer implements the ** interface in the go-plugin package
func (p *DatasourcePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	fmtrpc.RegisterDatasourceServer(s, &DatasourceGRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient implements the ** interface in the go-plugin package
func (p *DatasourcePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &DatasourceGRPCClient{client: fmtrpc.NewDatasourceClient(c)}, nil
}

// GRPCServer implements the ** interface in the go-plugin package
func (p *ControllerPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	fmtrpc.RegisterControllerServer(s, &ControllerGRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient implements the ** interface in the go-plugin package
func (p *ControllerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ControllerGRPCClient{client: fmtrpc.NewControllerClient(c)}, nil
}
