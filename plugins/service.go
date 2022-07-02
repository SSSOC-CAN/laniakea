/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/30
*/

package plugins

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/SSSOC-CAN/fmtd/api"
	e "github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/queue"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DATASOURCE_STR         = "datasource"
	CONTROLLER_STR         = "controller"
	ErrInvalidPluginString = bg.Error("invalid plugin string")
	ErrDuplicatePluginName = bg.Error("identical plugin names")
	ErrUnregsiteredPlugin  = bg.Error("given plugin not in plugin registry")
	ErrPluginNotReady      = bg.Error("plugin not ready")
	ErrPluginTimeout       = bg.Error("plugin timed out")
	PluginEOF              = "plugin EOF"
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "LANIAKEA_PLUGIN_MAGIC_COOKIE",
		MagicCookieValue: "a56e5daaa516e17d3d4b3d4685df9f8ca59c62c2d818cd5a7df13c039f134e16",
	}
	allowedPluginProtocols               = []plugin.Protocol{plugin.ProtocolGRPC}
	timeoutChecker         time.Duration = 10 * time.Second // time for when to check if plugins are unresponsive
)

type PluginManager struct {
	fmtrpc.UnimplementedPluginAPIServer
	pluginDir      string
	pluginCfgs     []*fmtrpc.PluginConfig
	logger         *PluginLogger
	pluginRegistry map[string]*PluginInstance
	queueManager   *queue.QueueManager
	quitChan       chan struct{}
	sync.WaitGroup
}

// Compile time check to ensure RpcServer implements api.RestProxyService
var _ api.RestProxyService = (*PluginManager)(nil)

// NewPluginManager takes a list of plugins, parses those plugin strings and instantiates a PluginManager
func NewPluginManager(pluginDir string, pluginCfgs []*fmtrpc.PluginConfig, zl zerolog.Logger) *PluginManager {
	l := NewPluginLogger("PLGN", &zl)
	return &PluginManager{
		pluginDir:    pluginDir,
		pluginCfgs:   pluginCfgs,
		logger:       l,
		queueManager: &queue.QueueManager{Logger: zl.With().Str("subsystem", "QMGR").Logger()},
		quitChan:     make(chan struct{}),
	}
}

// RegisterWithGrpcServer registers the PluginManager with the root gRPC server.
func (p *PluginManager) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterPluginAPIServer(grpcServer, p)
	return nil
}

// RegisterWithRestProxy registers the RPC Server with the REST proxy
func (p *PluginManager) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := fmtrpc.RegisterPluginAPIHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

// Start creates the connection to all registered plugins and establishes the connection to the API service
func (p *PluginManager) Start(ctx context.Context) error {
	instances := make(map[string]*PluginInstance)
	for _, cfg := range p.pluginCfgs {
		if _, ok := instances[cfg.Name]; ok {
			return ErrDuplicatePluginName
		}
		plug, err := getPluginFromType(cfg.Type)
		if err != nil {
			return err
		}
		newInstance, err := p.createPluginInstance(ctx, cfg, plug)
		if err != nil {
			return err
		}
		instances[cfg.Name] = newInstance
	}
	p.pluginRegistry = instances
	go p.monitorPlugins(ctx)
	return nil
}

// monitorPlugins is a goroutine for monitoring the plugin instance unresponsive state.
// If the plugin is in an unresponsive state, it will increment the timeout counter and restart the plugin.
// If the plugin timeout counter reaches the maximum timeouts, it will kill the plugin and set its state to killed
func (p *PluginManager) monitorPlugins(ctx context.Context) {
	defer p.Done()
	ticker := time.NewTicker(timeoutChecker)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, i := range p.pluginRegistry {
				// if the plugin is in the unresponsive state, we will try to restart
				if i.getState() == fmtrpc.Plugin_UNRESPONSIVE {
					// check if the plugin has exceeded the maximum number of times it can timeout
					if int64(i.getTimeoutCount()) >= i.cfg.MaxTimeouts {
						p.logger.zl.Error().Msg(fmt.Sprintf("Maximum timeouts reached for %v plugin, stopping plugin...", i.cfg.Name))
						err := i.stop(ctx)
						if err != nil {
							p.logger.zl.Error().Msg(fmt.Sprintf("could not gracefully stop the %v plugin: %v", i.cfg.Name, err))
							i.kill()
						}
						i.setKilled()
						continue
					}
					i.incrementTimeoutCount()
					p.logger.zl.Error().Msg(fmt.Sprintf("%v plugin timeout, stopping plugin...", i.cfg.Name))
					err := i.stop(ctx)
					if err != nil {
						p.logger.zl.Error().Msg(fmt.Sprintf("could not gracefully stop the %v plugin: %v", i.cfg.Name, err))
						i.kill()
					}
					p.logger.zl.Info().Msg(fmt.Sprintf("Restarting %v plugin...", i.cfg.Name))
					_, err = p.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: i.cfg.Name})
					if err != nil {
						p.logger.zl.Error().Msg(fmt.Sprintf("could not restart the %v plugin: %v", i.cfg.Name, err))
						i.setUnresponsive()
					}
				}
			}
		case <-p.quitChan:
			return
		}
	}
}

// getPluginFromType gets the plugin instance based on the type
func getPluginFromType(typeStr string) (plugin.Plugin, error) {
	var (
		plug plugin.Plugin
		err  error
	)
	switch typeStr {
	case DATASOURCE_STR:
		plug = DatasourcePlugin{}
	case CONTROLLER_STR:
		plug = ControllerPlugin{}
	default:
		err = ErrInvalidPluginType
	}
	return plug, err
}

// getPluginCodeFromType gets the rpc plugin code from the type
func getPluginCodeFromType(typeStr string) (fmtrpc.Plugin_PluginType, error) {
	var (
		plug fmtrpc.Plugin_PluginType
		err  error
	)
	switch typeStr {
	case DATASOURCE_STR:
		plug = fmtrpc.Plugin_DATASOURCE
	case CONTROLLER_STR:
		plug = fmtrpc.Plugin_CONTROLLER
	default:
		err = ErrInvalidPluginType
	}
	return plug, err
}

// createPluginInstance creates a new plugin instance from the given plugin name and type
func (p *PluginManager) createPluginInstance(ctx context.Context, cfg *fmtrpc.PluginConfig, plug plugin.Plugin) (*PluginInstance, error) {
	// create new plugin clients
	zLogger := p.logger.zl.With().Str("plugin", cfg.Name).Logger()
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{cfg.Name: plug},
		Cmd:              exec.Command(filepath.Join(p.pluginDir, cfg.ExecName)),
		Logger:           NewPluginLogger(cfg.Name, &zLogger),
		AllowedProtocols: allowedPluginProtocols,
	})
	// register the plugin with the queue manager
	outQ, unregister, err := p.queueManager.RegisterSource(cfg.Name)
	if err != nil {
		return nil, err
	}
	// create a new plugin instance object to handle plugin
	newInstance := &PluginInstance{
		cfg:           cfg,
		client:        newClient,
		outgoingQueue: outQ,
		logger:        &zLogger,
		startedAt:     time.Now(),
	}
	// push fmtd/laniakea version and get plugin version
	err = newInstance.pushVersion(ctx)
	if err != nil {
		return nil, err
	}
	err = newInstance.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	cleanUp := func() {
		if newInstance.client != nil {
			err = newInstance.stop(ctx)
			if err != nil {
				newInstance.logger.Error().Msg(fmt.Sprintf("could not safely stop %v plugin: %v", newInstance.cfg.Name, err))
				newInstance.kill()
			}
		}
		unregister()
	}
	newInstance.cleanUp = cleanUp
	return newInstance, nil
}

// Stop kills all plugin sub-processes safely
func (p *PluginManager) Stop() error {
	for _, instance := range p.pluginRegistry {
		instance.cleanUp()
	}
	close(p.quitChan)
	p.Wait()
	return nil
}

// StartRecord is the PluginAPI command which exposes the StartRecord method of all registered datasource plugins
func (p *PluginManager) StartRecord(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	err := instance.startRecord(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &fmtrpc.Empty{}, nil
}

// StopRecord is the PluginAPI command which exposes the StopRecord method of all registered datasource plugins
func (p *PluginManager) StopRecord(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	err := instance.stopRecord(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &fmtrpc.Empty{}, nil
}

// Subscribe is the PluginAPI command which exposes the data stream of any datasource plugin
func (p *PluginManager) Subscribe(req *fmtrpc.PluginRequest, stream fmtrpc.PluginAPI_SubscribeServer) error {
	// get the plugin instance from the registry
	_, ok := p.pluginRegistry[req.Name]
	if !ok {
		return status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	rand.Seed(time.Now().UnixNano())
	subscriberName := req.Name + utils.RandSeq(10)
	q, unregister, err := p.queueManager.RegisterListener(req.Name, subscriberName)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer unregister()
	// subscribe to queue
	sigChan, unsub, err := q.Subscribe(subscriberName)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer unsub()
	for {
		select {
		case qLength := <-sigChan:
			if qLength == 0 {
				return bg.Error(PluginEOF)
			}
			for i := 0; i < qLength-2; i++ {
				frame := q.Pop()
				if err := stream.Send(frame); err != nil {
					return err
				}
			}
		case <-stream.Context().Done():
			if errors.Is(stream.Context().Err(), context.Canceled) {
				return nil
			}
			return stream.Context().Err()
		}
	}
}

// StartPlugin is the PluginAPI command to start an existiing plugin
func (p *PluginManager) StartPlugin(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	// check if plugin is stopped or killed
	if instance.getState() != fmtrpc.Plugin_STOPPED && instance.getState() != fmtrpc.Plugin_KILLED {
		return nil, status.Error(codes.FailedPrecondition, e.ErrServiceAlreadyStarted.Error())
	}
	// create new plugin client
	zLogger := p.logger.zl.With().Str("plugin", instance.cfg.Name).Logger()
	plug, err := getPluginFromType(instance.cfg.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{instance.cfg.Name: plug},
		Cmd:              exec.Command(filepath.Join(p.pluginDir, instance.cfg.ExecName)),
		Logger:           NewPluginLogger(instance.cfg.Name, &zLogger),
		AllowedProtocols: allowedPluginProtocols,
	})
	instance.setClient(newClient)
	instance.setLogger(&zLogger)
	instance.setReady()
	return &fmtrpc.Empty{}, nil
}

// StopPlugin is the PluginAPI command to stop a given plugin gracefully
func (p *PluginManager) StopPlugin(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	err := instance.stop(ctx)
	if err == e.ErrServiceAlreadyStopped {
		return nil, status.Error(codes.FailedPrecondition, e.ErrServiceAlreadyStopped.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &fmtrpc.Empty{}, nil
}

// AddPlugin is the PluginAPI command to add a new plugin from a formatted plugin string
func (p *PluginManager) AddPlugin(ctx context.Context, req *fmtrpc.PluginConfig) (*fmtrpc.Plugin, error) {
	// validate plugin config
	err := ValidatePluginConfig(req, p.pluginDir)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, ok := p.pluginRegistry[req.Name]; ok {
		return nil, status.Error(codes.AlreadyExists, ErrDuplicatePluginName.Error())
	}
	plugType, err := getPluginFromType(req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	rpcPlugType, err := getPluginCodeFromType(req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// now create new instance and add to registry
	newInstance, err := p.createPluginInstance(ctx, req, plugType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	p.pluginRegistry[req.Name] = newInstance
	return &fmtrpc.Plugin{
		Name:      newInstance.cfg.Name,
		Type:      rpcPlugType,
		State:     newInstance.getState(),
		StartedAt: newInstance.startedAt.UnixMilli(),
		Version:   newInstance.version,
	}, nil
}

// ListPlugins is the PluginAPI command for listing all plugins in the plugin registry along with pertinent information about each one
func (p *PluginManager) ListPlugins(ctx context.Context, _ *fmtrpc.Empty) (*fmtrpc.PluginsList, error) {
	var plugins []*fmtrpc.Plugin
	for _, plug := range p.pluginRegistry {
		plugType, err := getPluginCodeFromType(plug.cfg.Type)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		plugins = append(plugins, &fmtrpc.Plugin{
			Name:      plug.cfg.Name,
			Type:      plugType,
			State:     plug.getState(),
			StartedAt: plug.startedAt.UnixMilli(),
			StoppedAt: plug.stoppedAt.UnixMilli(),
			Version:   plug.version,
		})
	}
	return &fmtrpc.PluginsList{Plugins: plugins}, nil
}

// Command is the PluginAPI command for sending an arbitrary amount of data to a controller service
func (p *PluginManager) Command(req *fmtrpc.ControllerPluginRequest, stream fmtrpc.PluginAPI_CommandServer) error {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	err := instance.command(context.Background(), req.Frame)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	rand.Seed(time.Now().UnixNano())
	subscriberName := req.Name + utils.RandSeq(10)
	q, unregister, err := p.queueManager.RegisterListener(req.Name, subscriberName)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer unregister()
	// subscribe to queue
	sigChan, unsub, err := q.Subscribe(subscriberName)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer unsub()
	for {
		select {
		case qLength := <-sigChan:
			if qLength == 0 {
				return status.Error(codes.OK, bg.Error(PluginEOF).Error())
			}
			for i := 0; i < qLength-2; i++ {
				frame := q.Pop()
				if err := stream.Send(frame); err != nil {
					return err
				}
			}
		case <-stream.Context().Done():
			if errors.Is(stream.Context().Err(), context.Canceled) {
				return nil
			}
			return stream.Context().Err()
		}
	}
}

// GetPlugin will retrieve the plugin information for a given plugin if it exists
func (p *PluginManager) GetPlugin(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Plugin, error) {
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	plugType, err := getPluginCodeFromType(instance.cfg.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &fmtrpc.Plugin{
		Name:      instance.cfg.Name,
		Type:      plugType,
		State:     instance.getState(),
		StartedAt: instance.startedAt.UnixMilli(),
		StoppedAt: instance.stoppedAt.UnixMilli(),
		Version:   instance.version,
	}, nil
}
