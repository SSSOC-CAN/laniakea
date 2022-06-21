/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package fmtd

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	
	"github.com/mattn/go-colorable"
	color "github.com/mgutz/ansi"
	"github.com/rs/zerolog"
	"github.com/SSSOC-CAN/fmtd/utils"
)

const (
	logFileRoot = "logfile"
	logFileExt  = "log"
	logFileName = "logfile.log"
)

// subLogger is a thin-wrapper for the `zerolog.Logger` struct
type subLogger struct {
	SubLogger zerolog.Logger
	Subsystem string
}

type moddedFileWriter struct {
    File *os.File
    maxFileSize int64 // MB
    maxFiles int64
    fileNameRoot string // must include path if not in current directory
    fileExt string
	pathToFile string
}

// Write Implements the io.Writer interface
func (w *moddedFileWriter) Write(p []byte) (n int, err error) {
    stat, err := w.File.Stat()
    if err != nil {
        return 0, err
    }
    // Check if maximum file size if exceeded
    if stat.Size() >= w.maxFileSize || stat.Size()+int64(len(p)) >= w.maxFileSize {
        // get current file name number
        r, err := regexp.Compile(fmt.Sprintf("%s([0-9]+)", w.fileNameRoot)) // may need file extension
        if err != nil {
            return 0, err
        }
        matches := r.FindStringSubmatch(stat.Name())
        var (
            fileNum int64
        )
        if len(matches) > 1 {
            fileNum, err = strconv.ParseInt(matches[1], 10, 64)
            if err != nil {
                return 0, err
            }
        }
        // Close current file and delete new file if it already exists
        w.File.Close()
        var newFileName string
        if fileNum >= w.maxFiles-1 {
            newFileName = fmt.Sprintf("%s.%s", w.fileNameRoot, w.fileExt)
        } else {
            newFileName = fmt.Sprintf("%s%v.%s", w.fileNameRoot, fileNum+int64(1), w.fileExt)
        }
        if utils.FileExists(fmt.Sprintf("%s/%s", w.pathToFile, newFileName)) {
            err = os.Remove(fmt.Sprintf("%s/%s", w.pathToFile, newFileName))
            if err != nil {
                return 0, err
            }
        }
        newFile, err := os.OpenFile(fmt.Sprintf("%s/%s", w.pathToFile, newFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
        if err != nil {
            return 0, err
        }
        w.File = newFile
    }
    return w.File.Write(p)
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
	log_file, err = os.OpenFile(config.LogFileDir+"/"+logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
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
	// use new modified writer
	modded_file := &moddedFileWriter{
        File: log_file,
        maxFileSize: config.MaxLogFileSize*1000000, // converting to Bytes
        maxFiles: config.MaxLogFiles,
        fileNameRoot: logFileRoot,
        fileExt: logFileExt,
		pathToFile: config.LogFileDir,
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
				case "warn":
					msg = color.Color(strings.ToUpper("["+x+"]"), "yellow")
				case "debug":
					msg = color.Color(strings.ToUpper("["+x+"]"), "yellow")
				case "trace":
					msg = color.Color(strings.ToUpper("["+x+"]"), "magenta")
				}
			}
			return msg+fmt.Sprintf("\t")
		}
		multi := zerolog.MultiLevelWriter(output, modded_file)
		logger = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(modded_file).With().Timestamp().Logger()
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