/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/07/07
*/

package plugins

import (
	"context"
	"errors"
	"io"
	"log"
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
	sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/SSSOCPaulCote/gux"
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
	ql := l.zl.With().Str("subsystem", "QMGR").Logger()
	return &PluginManager{
		pluginDir:    pluginDir,
		pluginCfgs:   pluginCfgs,
		logger:       l,
		queueManager: queue.NewQueueManager(&ql),
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
	p.pluginRegistry = make(map[string]*PluginInstance)
	for _, cfg := range p.pluginCfgs {
		if _, ok := p.pluginRegistry[cfg.Name]; ok {
			return ErrDuplicatePluginName
		}
		newInstance, err := p.createPluginInstance(ctx, cfg)
		if err != nil {
			return err
		}
		p.pluginRegistry[cfg.Name] = newInstance
	}
	p.Add(1)
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
				if i.getState() == fmtrpc.PluginState_UNRESPONSIVE {
					// check if the plugin has exceeded the maximum number of times it can timeout
					if int64(i.getTimeoutCount()) >= i.cfg.MaxTimeouts {
						p.logger.zl.Error().Msgf("Maximum timeouts reached for %v plugin, stopping plugin...", i.cfg.Name)
						err := i.stop(ctx)
						if err != nil {
							p.logger.zl.Error().Msgf("could not gracefully stop the %v plugin: %v", i.cfg.Name, err)
							i.kill()
						}
						i.setKilled()
						continue
					}
					i.incrementTimeoutCount()
					p.logger.zl.Error().Msgf("%v plugin timeout, stopping plugin...", i.cfg.Name)
					err := i.stop(ctx)
					if err != nil {
						p.logger.zl.Error().Msgf("could not gracefully stop the %v plugin: %v", i.cfg.Name, err)
						i.kill()
						i.setStopped()
					}
					p.logger.zl.Info().Msgf("Restarting %v plugin...", i.cfg.Name)
					_, err = p.StartPlugin(ctx, &fmtrpc.PluginRequest{Name: i.cfg.Name})
					if err != nil {
						p.logger.zl.Error().Msgf("could not restart the %v plugin: %v", i.cfg.Name, err)
						i.setUnresponsive()
					}
				}
			}
		case <-p.quitChan:
			return
		}
	}
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
func (p *PluginManager) createPluginInstance(ctx context.Context, cfg *fmtrpc.PluginConfig) (*PluginInstance, error) {
	// create new plugin clients
	zLogger := p.logger.zl.With().Str("plugin", cfg.Name).Logger()
	// assert plugin type
	var newClient *plugin.Client
	switch cfg.Type {
	case DATASOURCE_STR:
		plug := &sdk.DatasourcePlugin{}
		newClient = plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  sdk.HandshakeConfig,
			Plugins:          map[string]plugin.Plugin{cfg.Name: plug},
			Cmd:              exec.Command(filepath.Join(p.pluginDir, cfg.ExecName)),
			Logger:           NewPluginLogger(cfg.Name, &zLogger),
			AllowedProtocols: allowedPluginProtocols,
		})
	case CONTROLLER_STR:
		plug := &sdk.ControllerPlugin{}
		newClient = plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  sdk.HandshakeConfig,
			Plugins:          map[string]plugin.Plugin{cfg.Name: plug},
			Cmd:              exec.Command(filepath.Join(p.pluginDir, cfg.ExecName)),
			Logger:           NewPluginLogger(cfg.Name, &zLogger),
			AllowedProtocols: allowedPluginProtocols,
		})
	default:
		return nil, ErrInvalidPluginType
	}
	// register the plugin with the queue manager
	outQ, unregister, err := p.queueManager.RegisterSource(cfg.Name)
	if err != nil {
		newClient.Kill()
		return nil, err
	}
	// create a new plugin instance object to handle plugin
	newInstance := &PluginInstance{
		cfg:           cfg,
		client:        newClient,
		outgoingQueue: outQ,
		logger:        &zLogger,
		startedAt:     time.Now(),
		listeners:     make(map[string]*StateListener),
	}
	// push fmtd/laniakea version and get plugin version
	err = newInstance.pushVersion(ctx)
	if err != nil {
		newInstance.kill()
		return nil, err
	}
	err = newInstance.getVersion(ctx)
	if err != nil {
		newInstance.kill()
		return nil, err
	}
	p.logger.zl.Info().Msgf("registered plugin %s version: %v", newInstance.cfg.Name, newInstance.version)
	cleanUp := func() {
		if newInstance.client != nil {
			err = newInstance.stop(ctx)
			if err != nil {
				newInstance.logger.Error().Msgf("could not safely stop %v plugin: %v", newInstance.cfg.Name, err)
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
	p.queueManager.Shutdown()
	return nil
}

// StartRecord is the PluginAPI command which exposes the StartRecord method of all registered datasource plugins
func (p *PluginManager) StartRecord(ctx context.Context, req *fmtrpc.PluginRequest) (*proto.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	err := instance.startRecord(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.Empty{}, nil
}

// StopRecord is the PluginAPI command which exposes the StopRecord method of all registered datasource plugins
func (p *PluginManager) StopRecord(ctx context.Context, req *fmtrpc.PluginRequest) (*proto.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	err := instance.stopRecord(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.Empty{}, nil
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
				return status.Error(codes.OK, bg.Error(PluginEOF).Error())
			}
			for i := 0; i < qLength-1; i++ {
				frame := q.Pop()
				if frame.Type == io.EOF.Error() {
					return status.Error(codes.OK, bg.Error(PluginEOF).Error())
				}
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
func (p *PluginManager) StartPlugin(ctx context.Context, req *fmtrpc.PluginRequest) (*proto.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	// check if plugin is stopped or killed
	if instance.getState() != fmtrpc.PluginState_STOPPED && instance.getState() != fmtrpc.PluginState_KILLED {
		return nil, status.Error(codes.FailedPrecondition, e.ErrServiceAlreadyStarted.Error())
	}
	// create new plugin client
	zLogger := p.logger.zl.With().Str("plugin", instance.cfg.Name).Logger()
	var newClient *plugin.Client
	switch instance.cfg.Type {
	case DATASOURCE_STR:
		plug := &sdk.DatasourcePlugin{}
		newClient = plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  sdk.HandshakeConfig,
			Plugins:          map[string]plugin.Plugin{instance.cfg.Name: plug},
			Cmd:              exec.Command(filepath.Join(p.pluginDir, instance.cfg.ExecName)),
			Logger:           NewPluginLogger(instance.cfg.Name, &zLogger),
			AllowedProtocols: allowedPluginProtocols,
		})
	case CONTROLLER_STR:
		plug := &sdk.ControllerPlugin{}
		newClient = plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  sdk.HandshakeConfig,
			Plugins:          map[string]plugin.Plugin{instance.cfg.Name: plug},
			Cmd:              exec.Command(filepath.Join(p.pluginDir, instance.cfg.ExecName)),
			Logger:           NewPluginLogger(instance.cfg.Name, &zLogger),
			AllowedProtocols: allowedPluginProtocols,
		})
	default:
		return nil, status.Error(codes.InvalidArgument, ErrInvalidPluginType.Error())
	}
	instance.setClient(newClient)
	instance.setLogger(&zLogger)
	instance.setReady()
	return &proto.Empty{}, nil
}

// StopPlugin is the PluginAPI command to stop a given plugin gracefully
func (p *PluginManager) StopPlugin(ctx context.Context, req *fmtrpc.PluginRequest) (*proto.Empty, error) {
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
	return &proto.Empty{}, nil
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
	rpcPlugType, err := getPluginCodeFromType(req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// now create new instance and add to registry
	newInstance, err := p.createPluginInstance(ctx, req)
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
func (p *PluginManager) ListPlugins(ctx context.Context, _ *proto.Empty) (*fmtrpc.PluginsList, error) {
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
			for i := 0; i < qLength-1; i++ {
				frame := q.Pop()
				if frame.Type == io.EOF.Error() {
					return status.Error(codes.OK, bg.Error(PluginEOF).Error())
				}
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

// subscribeStateLoop is the go routine spun up for listening to state updates
func (p *PluginManager) subscribeStateLoop(pluginName string, wg sync.WaitGroup, stateQ *gux.Queue, quitChan chan struct{}) {
	defer wg.Done()
	instance := p.pluginRegistry[pluginName]
	subscriberName := pluginName + utils.RandSeq(10)
	sigChan, unsub, err := instance.subscribeState(subscriberName)
	if err != nil {
		log.Printf("Unexpected error when subscribing to plugin %v state changes: %v", instance.cfg.Name, err)
		instance.logger.Error().Msgf("Unexpected error when subscribing to plugin state changes: %v", err)
		return
	}
	defer unsub()
	lastState := instance.getState()
	stateQ.Push(&fmtrpc.PluginStateUpdate{
		Name:  pluginName,
		State: lastState,
	})
	for {
		select {
		case currentState := <-sigChan:
			if currentState != lastState {
				stateQ.Push(&fmtrpc.PluginStateUpdate{
					Name:  pluginName,
					State: currentState,
				})
				lastState = currentState
			}
		case <-quitChan:
			return
		}
	}
}

// SubscribePluginState implements the gRPC method of the same name. It subscribes a client to the given plugins state updates
func (p *PluginManager) SubscribePluginState(req *fmtrpc.PluginRequest, stream fmtrpc.PluginAPI_SubscribePluginStateServer) error {
	rand.Seed(time.Now().UnixNano())
	if req.Name == "all" {
		stateQ := gux.NewQueue()
		var wg sync.WaitGroup
		quitChans := []chan struct{}{}
		for name, _ := range p.pluginRegistry {
			quitChan := make(chan struct{})
			quitChans = append(quitChans, quitChan)
			wg.Add(1)
			go p.subscribeStateLoop(name, wg, stateQ, quitChan)
		}
		cleanUp := func() {
			for _, quitChan := range quitChans {
				close(quitChan)
			}
			wg.Wait()
		}
		defer cleanUp()
		subscriberName := req.Name + utils.RandSeq(10)
		sigChan, unsub, err := stateQ.Subscribe(subscriberName)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		defer unsub()
		for {
			select {
			case qLength := <-sigChan:
				if qLength == 0 {
					return status.Error(codes.OK, io.EOF.Error())
				}
				for i := 0; i < qLength-1; i++ {
					inter := stateQ.Pop()
					if inter != nil {
						update := inter.(*fmtrpc.PluginStateUpdate)
						if err := stream.Send(update); err != nil {
							return err
						}
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
	_, ok := p.pluginRegistry[req.Name]
	if !ok {
		return status.Error(codes.InvalidArgument, ErrUnregsiteredPlugin.Error())
	}
	stateQ := gux.NewQueue()
	quitChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go p.subscribeStateLoop(req.Name, wg, stateQ, quitChan)
	defer close(quitChan)
	subscriberName := req.Name + utils.RandSeq(10)
	sigChan, unsub, err := stateQ.Subscribe(subscriberName)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer unsub()
	for {
		select {
		case qLength := <-sigChan:
			if qLength == 0 {
				return status.Error(codes.OK, io.EOF.Error())
			}
			for i := 0; i < qLength-1; i++ {
				inter := stateQ.Pop()
				if inter != nil {
					update := inter.(*fmtrpc.PluginStateUpdate)
					if err := stream.Send(update); err != nil {
						return err
					}
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
