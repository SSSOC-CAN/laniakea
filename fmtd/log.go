package fmtd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"github.com/mattn/go-colorable"
	color "github.com/mgutz/ansi"
	"github.com/rs/zerolog"
)

// subLogger is a thin-wrapper for the `zerolog.Logger` struct
type subLogger struct {
	SubLogger zerolog.Logger
	Subsystem string
}

// log_level is a mapping of log levels as strings to structs from the zerolog package
var log_level = map[string]zerolog.Level{
	"INFO": zerolog.InfoLevel,
	"PANIC": zerolog.PanicLevel,
	"FATAL": zerolog.FatalLevel,
	"ERROR": zerolog.ErrorLevel,
	"DEBUG": zerolog.DebugLevel,
	"TRACE": zerolog.TraceLevel,
}

// InitLogger creates a new instance of the `zerolog.Logger` type. If `console_out` is true, it will output the logs to the console as well as the logfile
func InitLogger(config *Config) (zerolog.Logger, error) {
	// check to see if log file exists. If not, create one
	var (
		log_file 	*os.File
		err			error
		logger 		zerolog.Logger
	)
	log_file, err = os.OpenFile(config.LogFileDir+"/logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		// try to create the .fmtd dir and try again if log dir is default log dir
		if config.DefaultLogDir {
			err = os.Mkdir(config.LogFileDir, 0775)
			if err != nil {
				return zerolog.Logger{}, err
			}
			log_file, err = os.OpenFile(config.LogFileDir+"/logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
			if err != nil {
				return zerolog.Logger{}, err
			}
		} else {
			return zerolog.Logger{}, err
		}
	}
	
	
	if config.ConsoleOutput {
		output := zerolog.NewConsoleWriter()
		if runtime.GOOS == "windows" {
			output.Out = colorable.NewColorableStdout()
		} else {
			output.Out = os.Stderr
		}
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
		multi := zerolog.MultiLevelWriter(output, log_file)
		logger = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(log_file).With().Timestamp().Logger()
	}
	return logger, nil
}

// NewSubLogger takes a `zerolog.Logger` and string for the name of the subsystem and creates a `subLogger` for this subsystem
func NewSubLogger(l *zerolog.Logger, subsystem string) *subLogger {
	sub := l.With().Str("subsystem", subsystem).Logger()
	s := subLogger{
		SubLogger: sub,
		Subsystem: subsystem,
	}
	return &s
}

// LogWithErrors is a method which takes a log level and message as a string and writes the corresponding log. Returns an error if the log level doesn't exist
func (s subLogger) LogWithErrors(level, msg string) error {
	if lvl, ok := log_level[level]; ok {
		s.SubLogger.WithLevel(lvl).Msg(msg)
		return nil
	} else {
		s.SubLogger.Error().Msg(fmt.Sprintf("Log level %v not found.", level))
		return fmt.Errorf("log: Log level %v not found.", level)
	}
}

// Log is a method which takes a log level and message as a string and writes the corresponding log.
func (s subLogger) Log(level, msg string) {
	_ = s.LogWithErrors(level, msg)
}