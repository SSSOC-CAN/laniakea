package plugins

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

type PluginLogger struct {
	zl zerolog.Logger
}

func (l *PluginLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel:
		l.zl.Info().Msg(fmt.Sprintf(msg, args))
	case hclog.Trace:
		l.zl.Trace().Msg(fmt.Sprintf(msg, args))
	case hclog.Debug:
		l.zl.Debug().Msg(fmt.Sprintf(msg, args))
	case hclog.Info:
		l.zl.Info().Msg(fmt.Sprintf(msg, args))
	case hclog.Warn:
		l.zl.Warn().Msg(fmt.Sprintf(msg, args))
	case hclog.Error:
		l.zl.Error().Msg(fmt.Sprintf(msg, args))
	case hclog.Off:
		return
	default:
		l.zl.Debug().Msg(fmt.Sprintf("Unrecognized log level: %v", level))
	}
}

func (l *PluginLogger) Trace(msg string, args ...interface{}) {
	l.zl.Trace().Msg(fmt.Sprintf(msg, args))
}

func (l *PluginLogger) Debug(msg string, args ...interface{}) {
	l.zl.Debug().Msg(fmt.Sprintf(msg, args))
}

func (l *PluginLogger) Info(msg string, args ...interface{}) {
	l.zl.Info().Msg(fmt.Sprintf(msg, args))
}

func (l *PluginLogger) Warn(msg string, args ...interface{}) {
	l.zl.Warn().Msg(fmt.Sprintf(msg, args))
}

func (l *PluginLogger) Error(msg string, args ...interface{}) {
	l.zl.Error().Msg(fmt.Sprintf(msg, args))
}
