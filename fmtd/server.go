package fmtd

import (
	"sync/atomic"
	"github.com/rs/zerolog"
)

// Server is the object representing the state of the server
type Server struct {
	Active 		int32 // atomic
	Stopping	int32 // atomic
	cfg			*Config
	logger		*zerolog.Logger
}

// InitServer creates a new instance of the server and returns a pointer to it
func InitServer(config *Config, logger *zerolog.Logger) (*Server, error) {
	return &Server{
		cfg: config,
		logger: logger,
	}, nil
}

// Start starts the server. Returns an error if any issues occur
func (s *Server) Start() error {
	s.logger.Info().Msg("Starting Daemon...")
	atomic.StoreInt32(&s.Active, 1)
	return nil
}

// Stop stops the server. Returns an error if any issues occur
func (s *Server) Stop() error {
	s.logger.Info().Msg("Stopping Daemon...")
	atomic.StoreInt32(&s.Stopping, 1)
	return nil
}