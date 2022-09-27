/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20
*/

package plugins

import (
	"context"
	"os"
	"testing"

	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/SSSOC-CAN/laniakea/errors"
	"github.com/SSSOC-CAN/laniakea/lanirpc"
	"github.com/rs/zerolog"
)

// TestInstanceStates will test if updating and reading state works as intended
func TestInstanceStates(t *testing.T) {
	instance := PluginInstance{}
	t.Run("ready", func(t *testing.T) {
		if instance.getState() != lanirpc.PluginState_READY {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("busy", func(t *testing.T) {
		instance.setBusy()
		defer instance.setReady()
		if instance.getState() != lanirpc.PluginState_BUSY {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("unknown", func(t *testing.T) {
		instance.setUnknown()
		defer instance.setReady()
		if instance.getState() != lanirpc.PluginState_UNKNOWN {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("stopping", func(t *testing.T) {
		instance.setStopping()
		defer instance.setReady()
		if instance.getState() != lanirpc.PluginState_STOPPING {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("stopped", func(t *testing.T) {
		instance.setStopped()
		defer instance.setReady()
		if instance.getState() != lanirpc.PluginState_STOPPED {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("unresponsive", func(t *testing.T) {
		instance.setUnresponsive()
		defer instance.setReady()
		if instance.getState() != lanirpc.PluginState_UNRESPONSIVE {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("killed", func(t *testing.T) {
		instance.setKilled()
		defer instance.setReady()
		if instance.getState() != lanirpc.PluginState_KILLED {
			t.Errorf("Unexpected instance state: %v", instance.getState())
		}
	})
	t.Run("recording", func(t *testing.T) {
		instance.setRecording()
		defer instance.setNotRecording()
		if instance.getRecordState() != RECORDING {
			t.Errorf("Unexpected instance recording state: %v", instance.getRecordState())
		}
	})
	t.Run("not recording", func(t *testing.T) {
		if instance.getRecordState() != NOTRECORDING {
			t.Errorf("Unexpected instance recording state: %v", instance.getRecordState())
		}
	})
}

// TestTimeoutCount tests if the timeout counter is working properly
func TestTimeoutCount(t *testing.T) {
	instance := PluginInstance{}
	t.Run("default is zero", func(t *testing.T) {
		if instance.getTimeoutCount() != 0 {
			t.Errorf("Unexpected timeout count number: %v", instance.getTimeoutCount())
		}
	})
	t.Run("increment", func(t *testing.T) {
		instance.incrementTimeoutCount()
		if instance.getTimeoutCount() != 1 {
			t.Errorf("Unexpected timeout count number: %v", instance.getTimeoutCount())
		}
	})
	t.Run("reset counter", func(t *testing.T) {
		instance.resetTimeoutCount()
		if instance.getTimeoutCount() != 0 {
			t.Errorf("Unexpected timeout count number: %v", instance.getTimeoutCount())
		}
	})
}

// TestStartRecordNoPlugin tests the plugin instance startRecord function without an actual plugin started
func TestStartRecordNoPlugin(t *testing.T) {
	instance := PluginInstance{}
	t.Run("test when busy", func(t *testing.T) {
		instance.setBusy()
		defer instance.setReady()
		err := instance.startRecord(context.Background())
		if err != ErrPluginNotReady {
			t.Errorf("Unexpected error when calling startRecord: %v", err)
		}
	})
	t.Run("test when already recording", func(t *testing.T) {
		instance.setRecording()
		defer instance.setNotRecording()
		err := instance.startRecord(context.Background())
		if err != errors.ErrAlreadyRecording {
			t.Errorf("Unexpected error when calling startRecord: %v", err)
		}
	})
	t.Run("when client is uninitialized", func(t *testing.T) {
		err := instance.startRecord(context.Background())
		if err != ErrPluginNotStarted {
			t.Errorf("Unexpected error when calling startRecord: %v", err)
		}
	})
}

// TestStopRecordNoPlugin tests the plugin instance stopRecord function without an actual plugin started
func TestStopRecordNoPlugin(t *testing.T) {
	instance := PluginInstance{}
	t.Run("test when busy", func(t *testing.T) {
		instance.setBusy()
		defer instance.setReady()
		err := instance.stopRecord(context.Background())
		if err != ErrPluginNotReady {
			t.Errorf("Unexpected error when calling stopRecord: %v", err)
		}
	})
	t.Run("test when already stopped recording", func(t *testing.T) {
		err := instance.stopRecord(context.Background())
		if err != errors.ErrAlreadyStoppedRecording {
			t.Errorf("Unexpected error when calling stopRecord: %v", err)
		}
	})
	t.Run("when client is uninitialized", func(t *testing.T) {
		instance.setRecording()
		defer instance.setNotRecording()
		err := instance.stopRecord(context.Background())
		if err != ErrPluginNotStarted {
			t.Errorf("Unexpected error when calling startRecord: %v", err)
		}
	})
}

// TestStopNoPlugin tests the plugin instance stop function without an actual plugin started
func TestStopNoPlugin(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	instance := &PluginInstance{cfg: &lanirpc.PluginConfig{Name: "test"}, logger: &logger}
	t.Run("stopping state", func(t *testing.T) {
		instance.setStopping()
		defer instance.setReady()
		err := instance.stop(context.Background())
		if err != errors.ErrServiceAlreadyStopped {
			t.Errorf("Unexpected error when calling stop: %v", err)
		}
	})
	t.Run("stopped state", func(t *testing.T) {
		instance.setStopped()
		defer instance.setReady()
		err := instance.stop(context.Background())
		if err != errors.ErrServiceAlreadyStopped {
			t.Errorf("Unexpected error when calling stop: %v", err)
		}
	})
	t.Run("killed state", func(t *testing.T) {
		instance.setKilled()
		defer instance.setReady()
		err := instance.stop(context.Background())
		if err != errors.ErrServiceAlreadyStopped {
			t.Errorf("Unexpected error when calling stop: %v", err)
		}
	})
}

// TestCommandNoPlugin tests the plugin instance command function without a plugin started
func TestCommandNoPlugin(t *testing.T) {
	instance := &PluginInstance{}
	t.Run("state not ready or unknown", func(t *testing.T) {
		instance.setBusy()
		defer instance.setReady()
		err := instance.command(context.Background(), &proto.Frame{})
		if err != ErrPluginNotReady {
			t.Errorf("Unexpected error when calling command: %v", err)
		}
	})
	t.Run("uninitialized client", func(t *testing.T) {
		err := instance.command(context.Background(), &proto.Frame{})
		if err != ErrPluginNotStarted {
			t.Errorf("Unexpected error when calling command: %v", err)
		}
	})
}

// TestVersionNoPlugin tests the plugin instance pushVersion and getVersion functions without a plugin started
func TestVersionNoPlugin(t *testing.T) {
	instance := &PluginInstance{}
	t.Run("pushVersion state not ready or unknown", func(t *testing.T) {
		instance.setBusy()
		defer instance.setReady()
		err := instance.pushVersion(context.Background())
		if err != ErrPluginNotReady {
			t.Errorf("Unexpected error when calling command: %v", err)
		}
	})
	t.Run("pushVersion uninitialized client", func(t *testing.T) {
		err := instance.pushVersion(context.Background())
		if err != ErrPluginNotStarted {
			t.Errorf("Unexpected error when calling command: %v", err)
		}
	})
	t.Run("getVersion state not ready or unknown", func(t *testing.T) {
		instance.setBusy()
		defer instance.setReady()
		err := instance.getVersion(context.Background())
		if err != ErrPluginNotReady {
			t.Errorf("Unexpected error when calling command: %v", err)
		}
	})
	t.Run("getVersion uninitialized client", func(t *testing.T) {
		err := instance.getVersion(context.Background())
		if err != ErrPluginNotStarted {
			t.Errorf("Unexpected error when calling command: %v", err)
		}
	})
}
