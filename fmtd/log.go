package fmtd

import (
	"os"

	log "github.com/rs/zerolog"
)

func InitLogger() log.Logger {
	logger := log.New(os.Stderr).With().Timestamp().Logger().Output(log.ConsoleWriter{Out: os.Stderr})
	return logger
}