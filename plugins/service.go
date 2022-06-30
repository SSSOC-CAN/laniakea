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
	"strings"
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
	ErrInvalidPluginType   = bg.Error("invalid plugin type")
	ErrDuplicatePluginName = bg.Error("identical plugin names")
	ErrInvalidPluginName   = bg.Error("given plugin not in plugin registry")
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
	allowedPluginProtocols = []plugin.Protocol{plugin.ProtocolGRPC}
	defaultPluginTimeout   = 20 * time.Second
	defaultMaxTimeouts     = 3
	defaultMaxRestarts     = 3
)

type PluginRecordState int32

const (
	NOTRECORDING PluginRecordState = iota
	RECORDING
)

type PluginInstance struct {
	Name          string
	client        *plugin.Client
	state         fmtrpc.Plugin_PluginState
	timeout       time.Duration
	maxTimeouts   int
	maxRestarts   int
	timeoutCnt    int
	restartCnt    int
	incomingQueue *queue.Queue
	outgoingQueue *queue.Queue
	cleanUp       func()
	logger        *zerolog.Logger
	recordState   PluginRecordState
	sync.RWMutex
}

// setReady changes the plugin instance state to ready
func (i *PluginInstance) setReady() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_READY
}

// setBusy changes the plugin instance state to Busy
func (i *PluginInstance) setBusy() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_BUSY
}

// setStopping changes the plugin instance state to stopping
func (i *PluginInstance) setStopping() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_STOPPING
}

// setStopped changes the plugin instance state to stopped
func (i *PluginInstance) setStopped() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_STOPPED
}

// setUnknown changes the plugin instance state to unknown
func (i *PluginInstance) setUnknown() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_UNKNOWN
}

// setUnresponsive changes the plugin instance state to unresponsive
func (i *PluginInstance) setUnresponsive() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_UNRESPONSIVE
}

// setRecording changes the plugin recording state to recording
func (i *PluginInstance) setRecording() {
	i.Lock()
	defer i.Unlock()
	i.recordState = RECORDING
}

// setNotRecording changes the plugin recording state to not recordings
func (i *PluginInstance) setNotRecording() {
	i.Lock()
	defer i.Unlock()
	i.recordState = NOTRECORDING
}

// incremenetTimeoutCount incremenets the timeout count by one
func (i *PluginInstance) incremenetTimeoutCount() {
	i.Lock()
	defer i.Unlock()
	i.timeoutCnt += 1
}

// resetTimeoutCount resets the timeout count. Used when plugin is successfully restarted
func (i *PluginInstance) resetTimeoutCount() {
	i.Lock()
	defer i.Unlock()
	i.timeoutCnt = 0
}

// incremenetRestartCount incremenets the restart count by one
func (i *PluginInstance) incremenetRestartCount() {
	i.Lock()
	defer i.Unlock()
	i.restartCnt += 1
}

// startRecord is the method that starts the data recording process if this plugin is a datasource
func (i *PluginInstance) startRecord(ctx context.Context) error {
	// check if plugin is ready
	if i.state != fmtrpc.Plugin_READY {
		return ErrPluginNotReady
	}
	// check if we're already recording
	if i.recordState == RECORDING {
		return e.ErrAlreadyRecording
	}
	// set plugin as busy
	i.setBusy()
	defer i.setReady()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		return err
	}
	raw, err := gRPCClient.Dispense(i.Name)
	if err != nil {
		return err
	}
	datasource, ok := raw.(Datasource)
	if !ok {
		return ErrInvalidPluginType
	}
	// spin up go routine
	go func() {
		dataChan, err := datasource.StartRecord()
		if err != nil {
			i.logger.Error().Msg(fmt.Sprintf("could not start recording: %v", err))
			return
		}
		// change recording state
		i.setRecording()
		defer i.setNotRecording()
		timer := time.NewTimer(i.timeout)
	loop:
		for {
			select {
			case <-timer.C:
				i.logger.Error().Msg(ErrPluginTimeout.Error())
				// only increment the timeout count if we didn't request to stop recording
				if i.recordState != NOTRECORDING {
					i.incremenetTimeoutCount()
				}
				break loop
			case frame := <-dataChan:
				if frame == nil {
					i.logger.Info().Msg(PluginEOF)
					break loop
				}
				i.outgoingQueue.Push(frame)
				timer.Reset(i.timeout)
			}
		}
	}()
	return nil
}

// stopRecord stops the recording process of the datasource plugin
func (i *PluginInstance) stopRecord(ctx context.Context) error {
	// check if plugin is ready
	if i.state != fmtrpc.Plugin_READY {
		return ErrPluginNotReady
	}
	// check if plugin is recording
	if i.recordState != RECORDING {
		return e.ErrAlreadyStoppedRecording
	}
	// set plugin as busy
	i.setBusy()
	defer i.setReady()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		return err
	}
	raw, err := gRPCClient.Dispense(i.Name)
	if err != nil {
		return err
	}
	datasource, ok := raw.(Datasource)
	if !ok {
		return ErrInvalidPluginType
	}
	// create timeout context
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()
	errChan := make(chan error)
	go func(ctx context.Context, errChan chan error) {
		err := datasource.StopRecord()
		if err != nil {
			i.logger.Error().Msg(fmt.Sprintf("could not stop recording: %v", err))
		} else {
			i.setNotRecording()
		}
		errChan <- err
	}(ctx, errChan)
	select {
	case <-ctx.Done():
		i.logger.Error().Msg(ErrPluginTimeout.Error())
		i.incremenetTimeoutCount()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// kill will kill the actual plugin and reset counter information
func (i *PluginInstance) kill() {
	i.resetTimeoutCount()
	i.Lock()
	defer i.Unlock()
	i.client.Kill()
	i.client = nil
	i.recordState = NOTRECORDING
	noOpLog := i.logger.Nop()
	i.logger = &noOpLog
}

// stop will send the signal to kill the plugin
func (i *PluginInstance) stop(ctx context.Context) error {
	i.setStopping()
	defer i.setStopped()
	// try to gracefully stop recording if recording
	var err error
	if i.recordState == RECORDING {
		errChan := make(chan error)
		go func(ctx context.Context, errChan chan error) {
			err := i.stopRecord(ctx)
			errChan <- err
		}(ctx, errChan)
		select {
		case err = <-errChan:
			i.logger.Error().Msg(fmt.Sprintf("error when stopping %v plugin recording: %v", i.Name, err))
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				i.setUnresponsive()
			} else if err != nil {
				i.setUnknown()
			}
		}
	}
	i.kill()
	return err
}

// setClient will set a new plugin client for the plugin instace
func (i *PluginInstance) setClient(client *plugin.Client) {
	i.Lock()
	defer i.Unlock()
	i.client = client
}

func (i *PluginInstance) setLogger(logger *zerolog.Logger) {
	i.Lock()
	defer i.Unlock()
	i.logger = logger
}

type PluginManager struct {
	fmtrpc.UnimplementedPluginAPIServer
	pluginDir      string
	pluginMap      map[string]plugin.Plugin
	pluginExecs    map[string]string
	logger         *PluginLogger
	pluginRegistry map[string]*PluginInstance
	queueManager   *queue.QueueManager
}

// Compile time check to ensure RpcServer implements api.RestProxyService
var _ api.RestProxyService = (*PluginManager)(nil)

// NewPluginManager takes a list of plugins, parses those plugin strings and instantiates a PluginManager
func NewPluginManager(pluginDir string, listOfPlugins []string, zl zerolog.Logger) (*PluginManager, error) {
	plugins := make(map[string]plugin.Plugin)
	execs := make(map[string]string)
	var err error
loop:
	for _, pluginStr := range listOfPlugins {
		split, err := parsePluginString(pluginStr)
		if err != nil {
			break loop
		}
		if _, ok := plugins[split[0]]; ok {
			err = ErrDuplicatePluginName
			break loop
		}
		var plugType plugin.Plugin
		switch split[1] {
		case DATASOURCE_STR:
			plugType = &DatasourcePlugin{}
		case CONTROLLER_STR:
			plugType = &ControllerPlugin{}
		default:
			err = ErrInvalidPluginType
			break loop
		}
		plugins[split[0]] = plugType
		execs[split[0]] = split[2]
	}
	if err != nil {
		return nil, err
	}
	return &PluginManager{
		pluginDir:    pluginDir,
		pluginMap:    plugins,
		pluginExecs:  execs,
		logger:       NewPluginLogger("PLGN", zl),
		queueManager: &queue.QueueManager{Logger: zl.With().Str("subsystem", "QMGR").Logger()},
	}, nil
}

//parsePluginString parses a given plugin string and returns an error if there are any
func parsePluginString(pluginStr string) ([]string, error) {
	split := strings.Split(pluginStr, ":")
	if len(split) == 0 {
		return nil, ErrInvalidPluginString
	}
	return split, nil
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
func (p *PluginManager) Start() error {
	instances := make(map[string]*PluginInstance)
	for name, plug := range p.pluginMap {
		newInstance, err := p.createPluginInstance(name, plug)
		if err != nil {
			return err
		}
		instances[name] = newInstance
	}
	p.pluginRegistry = instances
	return nil
}

// createPluginInstance creates a new plugin instance from the given plugin name and type
func (p *PluginManager) createPluginInstance(name string, plug plugin.Plugin) (*PluginInstance, error) {
	// create new plugin clients
	zLogger := p.logger.zl.With().Str("plugin", name).Logger()
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{name: plug},
		Cmd:              exec.Command(fmt.Sprintf("%s/%s", p.pluginDir, p.pluginExecs[name])),
		Logger:           NewPluginLogger(name, &zLogger),
		AllowedProtocols: allowedPluginProtocols,
	})
	// register the plugin with the queue manager
	outQ, unregister, err := p.queueManager.RegisterSource(name)
	if err != nil {
		return nil, err
	}
	// create a new plugin instance object to handle plugin
	newInstance := &PluginInstance{
		Name:          name,
		client:        newClient,
		timeout:       defaultPluginTimeout,
		maxTimeouts:   defaultMaxTimeouts,
		maxRestarts:   defaultMaxRestarts,
		outgoingQueue: outQ,
		incomingQueue: queue.NewQueue(),
		logger:        &zLogger,
	}
	cleanUp := func() {
		if newInstance.client != nil {
			newInstance.client.Kill()
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
	return nil
}

// StartRecord is the PluginAPI command which exposes the StartRecord method of all registered datasource plugins
func (p *PluginManager) StartRecord(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.New(codes.InvalidArgument, ErrInvalidPluginName)
	}
	err := instance.startRecord(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err)
	}
	return &fmtrpc.Empty{}, nil
}

// StopRecord is the PluginAPI command which exposes the StopRecord method of all registered datasource plugins
func (p *PluginManager) StopRecord(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.New(codes.InvalidArgument, ErrInvalidPluginName)
	}
	err := instance.stopRecord(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err)
	}
	return &fmtrpc.Empty{}, nil
}

// Subscribe is the PluginAPI command which exposes the data stream of any datasource plugin
func (p *PluginManager) Subscribe(req *fmtrpc.PluginRequest, stream fmtrpc.PluginAPI_SubscribeServer) error {
	// get the plugin instance from the registry
	_, ok := p.pluginRegistry[req.Name]
	if !ok {
		return status.New(codes.InvalidArgument, ErrInvalidPluginName)
	}
	rand.Seed(time.Now().UnixNano())
	subscriberName := req.Name + utils.RandSeq(10)
	q, unregister, err := p.queueManager.RegisterListener(req.Name, subscriberName)
	if err != nil {
		return status.New(codes.Internal, err)
	}
	defer unregister()
	// subscribe to queue
	sigChan, unsub, err := q.Subscribe(subscriberName)
	if err != nil {
		return status.New(codes.Internal, err)
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
		return nil, status.New(codes.InvalidArgument, ErrInvalidPluginName)
	}
	// check if plugin is stopped or killed
	if instance.state != fmtrpc.Plugin_STOPPED && instance.state != fmtrpc.Plugin_KILLED {
		return nil, status.New(codes.FailedPrecondition, e.ErrServiceAlreadyStarted)
	}
	// create new plugin client
	zLogger := p.logger.zl.With().Str("plugin", instance.Name).Logger()
	newClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{instance.Name: p.pluginMap[instance.Name]},
		Cmd:              exec.Command(fmt.Sprintf("%s/%s", p.pluginDir, p.pluginExecs[instance.Name])),
		Logger:           NewPluginLogger(instance.Name, &zLogger),
		AllowedProtocols: allowedPluginProtocols,
	})
	instance.setClient(newClient)
	instance.setLogger(zLogger)
	return &fmtrpc.Empty{}, nil
}

// StopPlugin is the PluginAPI command to stop a given plugin gracefully
func (p *PluginManager) StopPlugin(ctx context.Context, req *fmtrpc.PluginRequest) (*fmtrpc.Empty, error) {
	// get the plugin instance from the registry
	instance, ok := p.pluginRegistry[req.Name]
	if !ok {
		return nil, status.New(codes.InvalidArgument, ErrInvalidPluginName)
	}
	// check if plugin is already stopped, stopping or killed
	if instance.state == fmtrpc.Plugin_STOPPING || instance.state == fmtrpc.Plugin_STOPPED || instance.state == fmtrpc.Plugin_KILLED {
		return nil, status.New(codes.FailedPrecondition, e.ErrServiceAlreadyStopped)
	}
	err := instance.stop(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err)
	}
	return &fmtrpc.Empty{}, nil
}

// AddPlugin is the PluginAPI command to add a new plugin from a formatted plugin string
func (p *PluginManager) AddPlugin(ctx context.Context, req *fmtrpc.AddPluginRequest) (*fmtrpc.Empty, error) {
	// verify plugin string adheres to proper format
	if !utils.VerifyPluginStringFormat(req.PluginString) {
		return nil, ErrInvalidPluginString
	}
	split, err := parsePluginString(req.PluginString)
	if err != nil {
		return nil, err
	}
	if _, ok := p.pluginMap[split[0]]; ok {
		return nil, ErrDuplicatePluginName
	}
	var plugType plugin.Plugin
	switch split[1] {
	case DATASOURCE_STR:
		plugType = &DatasourcePlugin{}
	case CONTROLLER_STR:
		plugType = &ControllerPlugin{}
	default:
		return nil, ErrInvalidPluginType
	}
	p.pluginMap[split[0]] = plugType
	p.pluginExecs[split[0]] = split[2]
	// now create new instance and add to registry
	newInstance, err := p.createPluginInstance(split[0], plugType)
	if err != nil {
		return nil, err
	}
	p.pluginRegistry[split[0]] = newInstance
	return &fmtrpc.Empty{}, nil
}

// ListPlugins is the PluginAPI command for listing all plugins in the plugin registry along with pertinent information about each one
func (p *PluginManager) ListPlugins(ctx context.Context, _ *fmtrpc.Empty) (*fmtrpc.PluginsList, error) {

}

// Command is the PluginAPI command for sending an arbitrary amount of data to a controller service
func (p *PluginManager) Command(req *fmtrpc.AddPluginRequest, stream fmtrpc.PluginAPI_CommandServer) error {

}
