// Package intercept defines objects and related functions to monitor requests to shutdown the application
package intercept

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"github.com/rs/zerolog"
)

var (
	started int32
)

// Interceptor is the object controlling application shutdown requests
type Interceptor struct {
	interruptChannel chan os.Signal
	Logger *zerolog.Logger
	shutdownChannel chan struct{}
	shutdownRequestChannel chan struct{}
	quit chan struct{}
}

// mainInterruptHandler listens for SIGINT (Ctrl+C) signals on the interruptChannel and shutdown requests on the 
// shutdownRequestChannel. 
func (interceptor *Interceptor) mainInterruptHandler() {
	defer atomic.StoreInt32(&started, 0)
	var isShutdown bool
	shutdown := func() {
		if isShutdown {
			if interceptor.Logger != nil {
				interceptor.Logger.Info().Msg("Already shutting down...")
			} else {
				log.Println("Already shutting down...")
			}
			return
		}
		isShutdown = true
		if interceptor.Logger != nil {
			interceptor.Logger.Info().Msg("Shutting down...")
		} else {
			log.Println("Shutting down...")
		}
		close(interceptor.quit)
	}
	for {
		select {
		case signal := <-interceptor.interruptChannel:
			if interceptor.Logger != nil {
				interceptor.Logger.Info().Msg(fmt.Sprintf("Received %v", signal))
			} else {
				log.Printf("Received %v", signal)
			}
			shutdown()
		case <-interceptor.shutdownRequestChannel:
			if interceptor.Logger != nil {
				interceptor.Logger.Info().Msg("Received shutdown request.")
			} else {
				log.Println("Received shutdown request.")
			}
			shutdown()
		case <-interceptor.quit:
			if interceptor.Logger != nil {
				interceptor.Logger.Info().Msg("Gracefully shutting down.")
			} else {
				log.Println("Gracefully shutting down.")
			}
			close(interceptor.shutdownChannel)
			signal.Stop(interceptor.interruptChannel)
			return
		}
	}
}

// RequestShutdown initiates a graceful shutdown from the application.
func (interceptor *Interceptor) RequestShutdown() {
	select {
	case interceptor.shutdownRequestChannel <- struct{}{}:
	case <-interceptor.quit:
	}
}

// ShutdownChannel returns the channel that will be closed once the main
// interrupt handler has exited.
func (c *Interceptor) ShutdownChannel() <-chan struct{} {
	return c.shutdownChannel
}

// InitInterceptor initializes the shutdown and interrupt interceptor
func InitInterceptor() (*Interceptor, error) {
	if !atomic.CompareAndSwapInt32(&started, 0, 1) {
		return &Interceptor{}, errors.New("Interceptor already initialized")
	}
	interceptor := Interceptor{
		interruptChannel: 		make(chan os.Signal, 1),
		shutdownChannel:		make(chan struct{}),
		shutdownRequestChannel:	make(chan struct{}),
		quit:					make(chan struct{}),
	}
	signalsToCatch := []os.Signal{
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}
	signal.Notify(interceptor.interruptChannel, signalsToCatch...)
	go interceptor.mainInterruptHandler()
	return &interceptor, nil
}