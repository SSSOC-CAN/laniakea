package plugins

import (
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

const (
	MissingKey = "EXTRA_VALUE_AT_END"
)

type PluginLogger struct {
	zl zerolog.Logger
	implied []interface{}
	name string
}

func NewPluginLogger(name string, logger zerolog.Logger) hclog.Logger {
	return &PluginLogger{
		zl: logger,
		name: name,
	}
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

// IsTrace checks if the current logger level is set to Trace
// TODO:SSSOCPaulCote - Make new config option for log level and integrate into pluggin logger as well
func (l *PluginLogger) IsTrace() bool {
	return false
}

// IsDebug checks if the current logger level is set to Debug
// TODO:SSSOCPaulCote - Make new config option for log level and integrate into pluggin logger as well
func (l *PluginLogger) IsDebug() bool {
	return false
}

// IsInfo checks if the current logger level is set to Info
// TODO:SSSOCPaulCote - Make new config option for log level and integrate into pluggin logger as well
func (l *PluginLogger) IsInfo() bool {
	return false
}

// IsWarn checks if the current logger level is set to Warn
// TODO:SSSOCPaulCote - Make new config option for log level and integrate into pluggin logger as well
func (l *PluginLogger) IsWarn() bool {
	return false
}

// IsError checks if the current logger level is set to Error
// TODO:SSSOCPaulCote - Make new config option for log level and integrate into pluggin logger as well
func (l *PluginLogger) IsError() bool {
	return false
}

// ImpliedArgs returns the loggers implied args
func (l *PluginLogger) ImpliedArgs() []interace{} {
	return l.implied
}

// With returns a new sublogger using new key/value pairs in the log output
func (l *PluginLogger) With(args ...interface{}) hclog.Logger {
	var (
		extra interface{}
		sl    zerolog.Logger
	)
	if len(args)%2 != 0 {
		extra = args[len(args)-1]
		args = args[:len(args)-1]
	}
	for i := 0; i < len(l.implied); i += 2 {
		if i == 0 {
			sl = l.zl.With().Str(l.implied[i].(string), fmt.Sprintf("%v", l.implied[i+1])).Logger()
		} else {
			sl = sl.With().Str(l.implied[i].(string), fmt.Sprintf("%v", l.implied[i+1])).Logger()
		}
	}
	for i := 0; i < len(args); i += 2 {
		sl = sl.With().Str(args[i].(string), fmt.Sprintf("%v", args[i+1])).Logger()
	}
	return &PluginLogger{
		zl: sl,
		implied: l.implied[:],
		name: l.name,
	}
}

// Name returns the name of the Logger
func (l *PluginLogger) Name() string {
	return l.name
}

// Creates a new sub-logger with added name
func (l *PluginLogger) Named(name string) hclog.Logger {
	newName := l.name+"."+name
	return &PluginLogger{
		zl: l.zl.With().Logger(),
		implied: l.implied[:],
		name: newName,
	}
}

// ResetNamed returns a new sub-logger with the given name
func (l *PluginLogger) ResetNamed(name string) hclog.Logger {
	return &PluginLogger{
		zl: l.zl.With().Logger(),
		implied: l.implied[:],
		name: name,
	}
}

// SetLevel doesn't do anything right now, just satisfies the interface
func (l *PluginLogger) SetLevel(level hclog.Level) hclog.Logger {
	//TODO:SSSOCPaulCote - Implement this 
	return
}

// StandardLogger returns a go standard library logger
func (l *PluginLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	if opts == nil {
		opts = &hclog.StandardLoggerOptions{}
	}

	return log.New(l.StandardWriter(opts), "", 0)
}

// StandardWriter returns an io package standard writer
func (l *PluginLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	newLog := *l
	return 
}