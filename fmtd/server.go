/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10

Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package fmtd

import (
	"sync/atomic"

	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/rs/zerolog"
)

// Server is the object representing the state of the server
type Server struct {
	Active int32 // atomic
	cfg    *Config
	logger *zerolog.Logger
}

// InitServer creates a new instance of the server and returns a pointer to it
func InitServer(config *Config, logger *zerolog.Logger) (*Server, error) {
	return &Server{
		cfg:    config,
		logger: logger,
	}, nil
}

// Start starts the server. Returns an error if any issues occur
func (s *Server) Start() error {
	s.logger.Info().Msg("Starting Daemon...")
	if ok := atomic.CompareAndSwapInt32(&s.Active, 0, 1); !ok {
		return errors.ErrServiceAlreadyStarted
	}
	s.logger.Info().Msgf("Daemon succesfully started. Version: %s", utils.AppVersion)
	return nil
}

// Stop stops the server. Returns an error if any issues occur
func (s *Server) Stop() error {
	s.logger.Info().Msg("Stopping Daemon...")
	if ok := atomic.CompareAndSwapInt32(&s.Active, 1, 0); !ok {
		return errors.ErrServiceAlreadyStopped
	}
	s.logger.Info().Msg("Daemon succesfully stopped.")
	return nil
}
