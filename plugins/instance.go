package plugins

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	e "github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/queue"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog"
)

type PluginRecordState int32

const (
	NOTRECORDING PluginRecordState = iota
	RECORDING
)

const (
	ErrPluginNotStarted = bg.Error("plugin not started")
)

type PluginInstance struct {
	cfg           *fmtrpc.PluginConfig
	client        *plugin.Client
	state         fmtrpc.Plugin_PluginState
	timeoutCnt    int
	outgoingQueue *queue.Queue
	cleanUp       func()
	logger        *zerolog.Logger
	recordState   PluginRecordState
	startedAt     time.Time
	stoppedAt     time.Time
	version       string
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

// setKilled changes the plugin instance state to killed
func (i *PluginInstance) setKilled() {
	i.Lock()
	defer i.Unlock()
	i.state = fmtrpc.Plugin_KILLED
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

// incrementTimeoutCount increments the timeout count by one
func (i *PluginInstance) incrementTimeoutCount() {
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

// startRecord is the method that starts the data recording process if this plugin is a datasource
func (i *PluginInstance) startRecord(ctx context.Context) error {
	// check if plugin is ready
	if i.getState() != fmtrpc.Plugin_READY && i.getState() != fmtrpc.Plugin_UNKNOWN {
		return ErrPluginNotReady
	}
	// check if we're already recording
	if i.getRecordState() == RECORDING {
		return e.ErrAlreadyRecording
	}
	if i.client == nil {
		return ErrPluginNotStarted
	}
	// set plugin as busy
	i.setBusy()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		i.setUnknown()
		return err
	}
	raw, err := gRPCClient.Dispense(i.cfg.Name)
	if err != nil {
		i.setUnknown()
		return err
	}
	datasource, ok := raw.(Datasource)
	if !ok {
		i.setUnknown()
		return ErrInvalidPluginType
	}
	// spin up go routine
	go func() {
		dataChan, err := datasource.StartRecord()
		if err != nil {
			i.logger.Error().Msg(fmt.Sprintf("could not start recording: %v", err))
			i.setUnknown()
			return
		}
		// change recording state
		i.setRecording()
		defer i.setNotRecording()
		timer := time.NewTimer(time.Duration(i.cfg.Timeout) * time.Second)
	loop:
		for {
			select {
			case <-timer.C:
				i.logger.Error().Msg(ErrPluginTimeout.Error())
				// only set state to unresponsive if we didn't request to stop recording
				if i.recordState != NOTRECORDING {
					i.setUnresponsive()
				}
				break loop
			case frame := <-dataChan:
				if frame == nil {
					i.logger.Info().Msg(PluginEOF)
					break loop
				}
				i.outgoingQueue.Push(frame)
				timer.Reset(time.Duration(i.cfg.Timeout) * time.Second)
			}
		}
	}()
	i.setReady()
	return nil
}

// stopRecord stops the recording process of the datasource plugin
func (i *PluginInstance) stopRecord(ctx context.Context) error {
	// check if plugin is ready
	if i.getState() != fmtrpc.Plugin_READY && i.getState() != fmtrpc.Plugin_UNKNOWN {
		return ErrPluginNotReady
	}
	// check if plugin is recording
	if i.getRecordState() != RECORDING {
		return e.ErrAlreadyStoppedRecording
	}
	if i.client == nil {
		return ErrPluginNotStarted
	}
	// set plugin as busy
	i.setBusy()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		i.setUnknown()
		return err
	}
	raw, err := gRPCClient.Dispense(i.cfg.Name)
	if err != nil {
		i.setUnknown()
		return err
	}
	datasource, ok := raw.(Datasource)
	if !ok {
		i.setUnknown()
		return ErrInvalidPluginType
	}
	// create timeout context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(i.cfg.Timeout)*time.Second)
	defer cancel()
	errChan := make(chan error)
	go func(ctx context.Context, errChan chan error) {
		err := datasource.StopRecord()
		if err != nil {
			i.logger.Error().Msg(fmt.Sprintf("could not stop recording: %v", err))
			i.setUnknown()
		} else {
			i.setNotRecording()
		}
		errChan <- err
	}(ctx, errChan)
	select {
	case <-ctx.Done():
		i.logger.Error().Msg(ErrPluginTimeout.Error())
		i.setUnresponsive()
		return ctx.Err()
	case err := <-errChan:
		if err == nil {
			i.setReady()
		}
		return err
	}
}

// kill will kill the actual plugin
func (i *PluginInstance) kill() {
	i.Lock()
	defer i.Unlock()
	if i.client == nil {
		return
	}
	i.client.Kill()
	i.client = nil
	i.recordState = NOTRECORDING
	noOpLog := zerolog.Nop()
	i.logger = &noOpLog
	i.stoppedAt = time.Now()
}

// stop will send the signal to kill the plugin
func (i *PluginInstance) stop(ctx context.Context) error {
	if i.getState() == fmtrpc.Plugin_STOPPING || i.getState() == fmtrpc.Plugin_STOPPED || i.getState() == fmtrpc.Plugin_KILLED {
		return e.ErrServiceAlreadyStopped
	}
	if i.client == nil {
		return ErrPluginNotStarted
	}
	i.setStopping()
	defer i.setStopped()
	// try to gracefully stop recording if recording
	var err error
	if i.getRecordState() == RECORDING {
		errChan := make(chan error)
		go func(ctx context.Context, errChan chan error) {
			err := i.stopRecord(ctx)
			errChan <- err
		}(ctx, errChan)
		select {
		case err = <-errChan:
			i.logger.Error().Msg(fmt.Sprintf("error when stopping %v plugin recording: %v", i.cfg.Name, err))
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				i.setUnresponsive()
			} else if err != nil {
				i.setUnknown()
			}
		}
	}
	defer i.kill()
	// we call the plugin Stop function to do a safe plugin stop if possible
	// check plugin type
	gRPCClient, err := i.client.Client()
	if err != nil {
		i.setUnknown()
		return err
	}
	raw, err := gRPCClient.Dispense(i.cfg.Name)
	if err != nil {
		i.setUnknown()
		return err
	}
	// create timeout context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(i.cfg.Timeout)*time.Second)
	defer cancel()
	errChan := make(chan error)

	switch plug := raw.(type) {
	case Datasource:
		plug = plug.(Datasource)
		go func(ctx context.Context, errChan chan error) {
			err := plug.Stop()
			if err != nil {
				i.logger.Error().Msg(fmt.Sprintf("could not safely stop plugin: %v", err))
			}
			errChan <- err
		}(ctx, errChan)
	case Controller:
		plug = plug.(Controller)
		go func(ctx context.Context, errChan chan error) {
			err := plug.Stop()
			if err != nil {
				i.logger.Error().Msg(fmt.Sprintf("could not safely stop plugin: %v", err))
			}
			errChan <- err
		}(ctx, errChan)
	default:
		return ErrInvalidPluginType
	}
	select {
	case <-ctx.Done():
		i.logger.Error().Msg(ErrPluginTimeout.Error())
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// setClient will set a new plugin client for the plugin instace
func (i *PluginInstance) setClient(client *plugin.Client) {
	i.Lock()
	defer i.Unlock()
	i.client = client
}

// setLogger will set the logger for the plugin instance
func (i *PluginInstance) setLogger(logger *zerolog.Logger) {
	i.Lock()
	defer i.Unlock()
	i.logger = logger
}

// command will pass along a command frame to a controller plugin and store the streamed data in the queue
func (i *PluginInstance) command(ctx context.Context, frame *fmtrpc.Frame) error {
	// check if plugin is ready
	if i.getState() != fmtrpc.Plugin_READY && i.getState() != fmtrpc.Plugin_UNKNOWN {
		return ErrPluginNotReady
	}
	if i.client == nil {
		return ErrPluginNotStarted
	}
	// set plugin as busy
	i.setBusy()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		i.setUnknown()
		return err
	}
	raw, err := gRPCClient.Dispense(i.cfg.Name)
	if err != nil {
		i.setUnknown()
		return err
	}
	ctrller, ok := raw.(Controller)
	if !ok {
		i.setUnknown()
		return ErrInvalidPluginType
	}
	// spin up go routine
	go func() {
		dataChan, err := ctrller.Command(frame)
		if err != nil {
			i.logger.Error().Msg(fmt.Sprintf("could not send command: %v", err))
			i.setUnknown()
			return
		}
		timer := time.NewTimer(time.Duration(i.cfg.Timeout) * time.Second)
	loop:
		for {
			select {
			case <-timer.C:
				i.logger.Error().Msg(ErrPluginTimeout.Error())
				// only set to unresponsive if we didn't request to stop recording
				if i.recordState != NOTRECORDING {
					i.setUnresponsive()
				}
				break loop
			case frame := <-dataChan:
				if frame == nil {
					i.logger.Info().Msg(PluginEOF)
					break loop
				}
				i.outgoingQueue.Push(frame)
				timer.Reset(time.Duration(i.cfg.Timeout) * time.Second)
			}
		}
	}()
	i.setReady()
	return nil
}

// pushVersion will push the fmtd/laniakea version to the plugin for compatibility reasons
func (i *PluginInstance) pushVersion(ctx context.Context) error {
	// check if plugin is ready
	if i.getState() != fmtrpc.Plugin_READY && i.getState() != fmtrpc.Plugin_UNKNOWN {
		return ErrPluginNotReady
	}
	if i.client == nil {
		return ErrPluginNotStarted
	}
	// set plugin as busy
	i.setBusy()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		i.setUnknown()
		return err
	}
	raw, err := gRPCClient.Dispense(i.cfg.Name)
	if err != nil {
		i.setUnknown()
		return err
	}
	// setup timeout context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(i.cfg.Timeout)*time.Second)
	defer cancel()

	// spin up go routine
	errChan := make(chan error)
	switch plug := raw.(type) {
	case Datasource:
		plug = plug.(Datasource)
		go func(ctx context.Context, errChan chan error) {
			errChan <- plug.PushVersion(utils.AppVersion)
		}(ctx, errChan)
	case Controller:
		plug = plug.(Controller)
		go func(ctx context.Context, errChan chan error) {
			errChan <- plug.PushVersion(utils.AppVersion)
		}(ctx, errChan)
	default:
		i.setUnknown()
		return ErrInvalidPluginType
	}

	// wait for error response
	select {
	case err = <-errChan:
		if err != nil {
			i.setUnknown()
			return err
		} else {
			i.setReady()
			return nil
		}
	case <-ctx.Done():
		i.setUnresponsive()
		return context.Canceled
	}
}

// setVersion sets the plugin version instance attribute
func (i *PluginInstance) setVersion(v string) {
	i.Lock()
	defer i.Unlock()
	i.version = v
}

// getVersion will get the plugin version for compatibility reasons
func (i *PluginInstance) getVersion(ctx context.Context) error {
	// check if plugin is ready
	if i.getState() != fmtrpc.Plugin_READY && i.getState() != fmtrpc.Plugin_UNKNOWN {
		return ErrPluginNotReady
	}
	if i.client == nil {
		return ErrPluginNotStarted
	}
	// set plugin as busy
	i.setBusy()
	// check if plugin is a Datasource
	gRPCClient, err := i.client.Client()
	if err != nil {
		i.setUnknown()
		return err
	}
	raw, err := gRPCClient.Dispense(i.cfg.Name)
	if err != nil {
		i.setUnknown()
		return err
	}
	// setup timeout context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(i.cfg.Timeout)*time.Second)
	defer cancel()
	// spin up go routine
	respChan := make(chan struct {
		Version string
		Err     error
	})
	switch plug := raw.(type) {
	case Datasource:
		plug = plug.(Datasource)
		go func(ctx context.Context, respChan chan struct {
			Version string
			Err     error
		}) {
			version, err := plug.GetVersion()
			respChan <- struct {
				Version string
				Err     error
			}{
				Version: version,
				Err:     err,
			}
		}(ctx, respChan)
	case Controller:
		plug = plug.(Controller)
		go func(ctx context.Context, respChan chan struct {
			Version string
			Err     error
		}) {
			version, err := plug.GetVersion()
			respChan <- struct {
				Version string
				Err     error
			}{
				Version: version,
				Err:     err,
			}
		}(ctx, respChan)
	default:
		i.setUnknown()
		return ErrInvalidPluginType
	}

	// wait for error response
	select {
	case resp := <-respChan:
		if resp.Err != nil {
			i.setUnknown()
			return resp.Err
		} else {
			i.setVersion(resp.Version)
			i.setReady()
			return nil
		}
	case <-ctx.Done():
		i.setUnresponsive()
		return context.Canceled
	}
}

// getState is a goroutine safe way to read the current plugin state
func (i *PluginInstance) getState() fmtrpc.Plugin_PluginState {
	i.RLock()
	defer i.RUnlock()
	return i.state
}

// getRecordState is a goroutine safe way to read the current plugin record state
func (i *PluginInstance) getRecordState() PluginRecordState {
	i.RLock()
	defer i.RUnlock()
	return i.recordState
}

// getTimeoutCount returns the current plugin instance timeout count
func (i *PluginInstance) getTimeoutCount() int {
	i.RLock()
	defer i.RUnlock()
	return i.timeoutCnt
}
