package fmtd

import (
	"fmt"
	"os"
	log "github.com/rs/zerolog"
	"strings"
	color "github.com/mgutz/ansi"
)

// subLogger is a thin-wrapper for the `zerolog.Logger` struct
type subLogger struct {
	SubLogger log.Logger
	Subsystem string
}

// log_level is a mapping of log levels as strings to structs from the zerolog package
var log_level = map[string]log.Level{
	"INFO": log.InfoLevel,
	"PANIC": log.PanicLevel,
	"FATAL": log.FatalLevel,
	"ERROR": log.ErrorLevel,
	"DEBUG": log.DebugLevel,
	"TRACE": log.TraceLevel,
}

// InitLogger creates a new instance of the `zerolog.Logger` type. If `console_out` is true, it will output the logs to the console as well as the logfile
func InitLogger(console_output bool) log.Logger {
	var logger log.Logger
	if console_output {
		output := log.ConsoleWriter{Out: os.Stderr}
		output.FormatLevel = func(i interface{}) string {
			var msg string
			switch v := i.(type) {
			default:
				x := fmt.Sprintf("%v", v)
				switch x {
				case "info":
					msg = color.Color(strings.ToUpper("["+x+"]"), "green")
				case "panic":
					msg = color.Color(strings.ToUpper("["+x+"]"), "red")
				case "fatal":
					msg = color.Color(strings.ToUpper("["+x+"]"), "red")
				case "error":
					msg = color.Color(strings.ToUpper("["+x+"]"), "red")
				case "debug":
					msg = color.Color(strings.ToUpper("["+x+"]"), "yellow")
				case "trace":
					msg = color.Color(strings.ToUpper("["+x+"]"), "magenta")
				}
			}
			return msg+fmt.Sprintf("\t")
		}
		multi := log.MultiLevelWriter(output, os.Stdout)
		logger = log.New(multi).With().Timestamp().Logger()
	} else {
		logger = log.New(os.Stderr).With().Timestamp().Logger()
	}
	return logger
}

// NewSubLogger takes a `zerolog.Logger` and string for the name of the subsystem and creates a `subLogger` for this subsystem
func NewSubLogger(l log.Logger, subsystem string) *subLogger {
	sub := l.With().Str("subsystem", subsystem).Logger()
	s := subLogger{
		SubLogger: sub,
		Subsystem: subsystem,
	}
	return &s
}

// Log is a method which takes a log level and message as a string and writes the corresponding log
func (s subLogger) Log(level, msg string) {
	if lvl, ok := log_level[level]; ok {
		s.SubLogger.WithLevel(lvl).Msg(msg)
	} else {
		s.SubLogger.Error().Msg(fmt.Sprintf("Log level %v not found.", level))
	}
}