/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/30

MIT License

Copyright (c) 2017 HashiCorp

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package plugins

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

const (
	MissingKey = "EXTRA_VALUE_AT_END"
)

var logTimestampRegexp = regexp.MustCompile(`^[\d\s\:\/\.\+-TZ]*`)

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
		l.Info(msg, args)
	case hclog.Trace:
		l.Trace(msg, args)
	case hclog.Debug:
		l.Debug(msg, args)
	case hclog.Info:
		l.Info(msg, args)
	case hclog.Warn:
		l.Warn(msg, args)
	case hclog.Error:
		l.Error(msg, args)
	case hclog.Off:
		return
	default:
		l.zl.Debug().Msg(fmt.Sprintf("Unrecognized log level: %v", level))
	}
}

func (l *PluginLogger) Trace(msg string, args ...interface{}) {
	if len(args) == 0 {
		l.zl.Trace().Msg(msg)
	} else {
		if len(args)%2 != 0 {
			args = append(args, MissingKey)
		}
		newMsg := msg + ": "
		for i := 0; i < len(args); i += 2 {
			newMsg += fmt.Sprintf("%v=%v ", args[i], args[i+1])
		}
		l.zl.Trace().Msg(newMsg)
	}
}

func (l *PluginLogger) Debug(msg string, args ...interface{}) {
	if len(args) == 0 {
		l.zl.Debug().Msg(msg)
	} else {
		if len(args)%2 != 0 {
			args = append(args, MissingKey)
		}
		newMsg := msg + ": "
		for i := 0; i < len(args); i += 2 {
			newMsg += fmt.Sprintf("%v=%v ", args[i], args[i+1])
		}
		l.zl.Debug().Msg(newMsg)
	}
}

func (l *PluginLogger) Info(msg string, args ...interface{}) {
	if len(args) == 0 {
		l.zl.Info().Msg(msg)
	} else {
		if len(args)%2 != 0 {
			args = append(args, MissingKey)
		}
		newMsg := msg + ": "
		for i := 0; i < len(args); i += 2 {
			newMsg += fmt.Sprintf("%v=%v ", args[i], args[i+1])
		}
		l.zl.Info().Msg(newMsg)
	}
}

func (l *PluginLogger) Warn(msg string, args ...interface{}) {
	if len(args) == 0 {
		l.zl.Warn().Msg(msg)
	} else {
		if len(args)%2 != 0 {
			args = append(args, MissingKey)
		}
		newMsg := msg + ": "
		for i := 0; i < len(args); i += 2 {
			newMsg += fmt.Sprintf("%v=%v ", args[i], args[i+1])
		}
		l.zl.Warn().Msg(newMsg)
	}
}

func (l *PluginLogger) Error(msg string, args ...interface{}) {
	if len(args) == 0 {
		l.zl.Error().Msg(msg)
	} else {
		if len(args)%2 != 0 {
			args = append(args, MissingKey)
		}
		newMsg := msg + ": "
		for i := 0; i < len(args); i += 2 {
			newMsg += fmt.Sprintf("%v=%v ", args[i], args[i+1])
		}
		l.zl.Error().Msg(newMsg)
	}
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
	newImplied := l.implied[:]
	newImplied = append(newImplied, args)
	if extra != nil {
		newImplied = append(newImplied, MissingKey, extra)
	}
	return &PluginLogger{
		zl:      sl,
		implied: newImplied,
		name:    l.name,
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
func (l *PluginLogger) SetLevel(level hclog.Level) {
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
	return &LogWriter{
		log: &l.zl,
		inferLevels: opts.InferLevels,
		inferLevelsWithTimestamp: opts.inferLevelsWithTimestamp,
		forceLevel: opts.ForceLevel,
	}
}

type LogWriter struct {
	log	zerolog.Logger
	inferLevels bool
	inferLevelsWithTimestamp bool
	forceLevel hclog.Level
}

func (w *LogWriter) Write(data []byte) (int, error) {
	str := string(bytes.TrimRight(data, "\n"))
	if w.forceLevel != hclog.NoLevel {
		_, str := w.pickLevel(str)
		w.dispatch(str, w.forceLevel)
	} else if w.inferLevels {
		if w.inferLevelsWithTimestamp {
			str = w.trimTimestamp(str)
		}
		lvl, str := w.pickLevel(str)
		w.dispatch(str, lvl)
	} else {
		w.log.Info().Msg(str)
	}
	return len(data), nil
}

func (w *LogWriter) dispatch(str string, level hclog.Level) {
	switch level {
	case hclog.Trace:
		w.log.Trace().Msg(str)
	case hclog.Debug:
		w.log.Debug().Msg(str)
	case hclog.Info:
		w.log.Info().Msg(str)
	case hclog.Warn:
		w.log.Warn().Msg(str)
	case hclog.Error:
		w.log.Error().Msg(str)
	default:
		w.log.Info().Msg(str)
	}
}

// Detect, based on conventions, what log level this is.
func (w *LogWriter) pickLevel(str string) (hclog.Level, string) {
	switch {
	case strings.HasPrefix(str, "[DEBUG]"):
		return hclog.Debug, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[TRACE]"):
		return hclog.Trace, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[INFO]"):
		return hclog.Info, strings.TrimSpace(str[6:])
	case strings.HasPrefix(str, "[WARN]"):
		return hclog.Warn, strings.TrimSpace(str[6:])
	case strings.HasPrefix(str, "[ERROR]"):
		return hclog.Error, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[ERR]"):
		return hclog.Error, strings.TrimSpace(str[5:])
	case strings.HasPrefix(str, "[FATAL]"):
		return hclog.Error, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[PANIC]"):
		return hclog.Error, strings.TrimSpace(str[7:])
	default:
		return hclog.Info, str
	}
}

func (w *LogWriter) trimTimestamp(str string) string {
	idx := logTimestampRegexp.FindStringIndex(str)
	return str[idx[1]:]
}