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
package intercept

import (
	"context"
	"fmt"
	"sync"

	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	bg "github.com/SSSOCPaulCote/blunderguard"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

type rpcState uint8

const (
	waitingToStart rpcState = iota
	daemonLocked
	daemonUnlocked
	rpcActive
	ErrWaitingToStart          = bg.Error("waiting to start, RPC services not available")
	ErrDaemonLocked            = bg.Error("daemon locked, unlock it to enable full RPC access")
	ErrDaemonUnlocked          = bg.Error("daemon already unlocked, unlocker service is no longer available")
	ErrRPCStarting             = bg.Error("the RPC server is in the process of starting up, but not yet ready to accept calls")
	ErrInvalidRPCState         = bg.Error("invalid RPC state")
	ErrDuplicateMacConstraints = bg.Error("duplicate macaroon constraints for given path")
	ErrUnknownPermission       = bg.Error("unknown permission for given method")
)

var (
	// List of commands that don't need macaroons
	macaroonWhitelist = map[string]struct{}{
		"/fmtrpc.Fmt/TestCommand":         {},
		"/fmtrpc.Unlocker/Login":          {}, //don't need a macaroon to login because succesful login will create macaroons
		"/fmtrpc.Unlocker/SetPassword":    {},
		"/fmtrpc.Unlocker/ChangePassword": {},
	}
)

// GrpcInteceptor struct is a data structure with attributes relevant to creating the gRPC interceptor
type GrpcInterceptor struct {
	state         rpcState
	noMacaroons   bool
	log           *zerolog.Logger
	permissionMap map[string][]bakery.Op
	svc           *macaroons.Service
	sync.RWMutex
}

// NewGrpcInterceptor instantiates a new GrpcInterceptor struct
func NewGrpcInterceptor(log *zerolog.Logger, noMacaroons bool) *GrpcInterceptor {
	return &GrpcInterceptor{
		state:         waitingToStart,
		noMacaroons:   noMacaroons,
		permissionMap: make(map[string][]bakery.Op),
		log:           log,
	}
}

// SetWaitingToStart changes the rRPC state to waitingToStart
func (i *GrpcInterceptor) SetWaitingToStart() {
	i.Lock()
	defer i.Unlock()
	i.state = waitingToStart
}

// SetDaemonLocked changes the RPC state from waitingToStart to locked
func (i *GrpcInterceptor) SetDaemonLocked() {
	i.Lock()
	defer i.Unlock()
	i.state = daemonLocked
}

// SetDaemonUnlocked changes the RPC state from locked to unlocked
func (i *GrpcInterceptor) SetDaemonUnlocked() {
	i.Lock()
	defer i.Unlock()
	i.state = daemonUnlocked
}

// SetRPCActive changes the RPC state from unlocked to rpcActive
func (i *GrpcInterceptor) SetRPCActive() {
	i.Lock()
	defer i.Unlock()
	i.state = rpcActive
}

// CreateGrpcOptions creates a array of gRPC interceptors
func (i *GrpcInterceptor) CreateGrpcOptions() []grpc.ServerOption {
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var strmInterceptors []grpc.StreamServerInterceptor
	// add the log interceptors
	unaryInterceptors = append(
		unaryInterceptors, logUnaryServerInterceptor(i.log),
	)
	strmInterceptors = append(
		strmInterceptors, logStreamServerInterceptor(i.log),
	)
	// Next we'll add our RPC state check interceptors, that will check
	// whether the attempted call is allowed in the current state.
	unaryInterceptors = append(
		unaryInterceptors, i.rpcStateUnaryServerInterceptor(),
	)
	strmInterceptors = append(
		strmInterceptors, i.rpcStateStreamServerInterceptor(),
	)
	// add macaroon interceptors
	unaryInterceptors = append(
		unaryInterceptors, i.MacaroonUnaryServerInterceptor(),
	)
	strmInterceptors = append(
		strmInterceptors, i.MacaroonStreamServerInterceptor(),
	)
	// Create server options from the interceptors we just set up.
	chainedUnary := grpc_middleware.WithUnaryServerChain(
		unaryInterceptors...,
	)
	chainedStream := grpc_middleware.WithStreamServerChain(
		strmInterceptors...,
	)
	serverOpts := []grpc.ServerOption{chainedUnary, chainedStream}
	return serverOpts
}

// checkRPCState checks whether a call to the given server is allowed in the current RPC state
func (i *GrpcInterceptor) checkRPCState(srv interface{}) error {
	i.RLock()
	state := i.state
	i.RUnlock()

	switch state {
	case waitingToStart:
		return ErrWaitingToStart
	case daemonLocked:
		_, ok := srv.(fmtrpc.UnlockerServer)
		if !ok {
			return ErrDaemonLocked
		}
	case daemonUnlocked, rpcActive:
		_, ok := srv.(fmtrpc.UnlockerServer)
		if ok {
			return ErrDaemonUnlocked
		}
	default:
		return ErrInvalidRPCState
	}
	return nil
}

// rpcStateUnaryServerInterceptor is a GRPC interceptor that checks whether
// calls to the given gGRPC server is allowed in the current rpc state.
func (i *GrpcInterceptor) rpcStateUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if err := i.checkRPCState(info.Server); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// rpcStateStreamServerInterceptor is a GRPC interceptor that checks whether
// calls to the given gGRPC server is allowed in the current rpc state.
func (i *GrpcInterceptor) rpcStateStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		if err := i.checkRPCState(srv); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

// AddPermissions adds the inputted permission to the permissionMap attribute of the GrpcInterceptor struct
func (i *GrpcInterceptor) AddPermissions(perms map[string][]bakery.Op) error {
	for m, ops := range perms {
		err := i.AddPermission(m, ops)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddPermission adds a new macaroon rule for the given method
func (i *GrpcInterceptor) AddPermission(method string, ops []bakery.Op) error {
	if _, ok := i.permissionMap[method]; ok {
		return ErrDuplicateMacConstraints
	}
	i.permissionMap[method] = ops
	return nil
}

// Permissions returns the current set of macaroon permissions
func (i *GrpcInterceptor) Permissions() map[string][]bakery.Op {
	c := make(map[string][]bakery.Op)
	for k, v := range i.permissionMap {
		s := make([]bakery.Op, len(v))
		copy(s, v)
		c[k] = s
	}
	return c
}

// checkMacaroon validates that the context contains the macaroon needed to
// invoke the given RPC method.
func (i *GrpcInterceptor) checkMacaroon(ctx context.Context,
	fullMethod string) error {

	// If noMacaroons is set, we'll always allow the call.
	if i.noMacaroons {
		return nil
	}

	// Check whether the method is whitelisted, if so we'll allow it
	// regardless of macaroons.
	_, ok := macaroonWhitelist[fullMethod]
	if ok {
		return nil
	}
	svc := i.svc

	// If the macaroon service is not yet active, we cannot allow
	// the call.
	if svc == nil {
		return errors.ErrMacSvcNil
	}

	uriPermissions, ok := i.permissionMap[fullMethod]
	if !ok {
		return ErrUnknownPermission
	}

	// Find out if there is an external validator registered for
	// this method. Fall back to the internal one if there isn't.
	validator, ok := svc.ExternalValidators[fullMethod]
	if !ok {
		validator = svc
	}

	// Now that we know what validator to use, let it do its work.
	return validator.ValidateMacaroon(ctx, uriPermissions, fullMethod)
}

// logUnaryServerInterceptor is a simple UnaryServerInterceptor that will
// automatically log any errors that occur when serving a client's unary
// request.
func logUnaryServerInterceptor(log *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("[%v]: %v", info.FullMethod, err))
		}
		return resp, err
	}
}

// logStreamServerInterceptor is a simple StreamServerInterceptor that
// will log any errors that occur while processing a client or server streaming
// RPC.
func logStreamServerInterceptor(log *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("[%v]: %v", info.FullMethod, err))
		}
		return err
	}
}

// MacaroonUnaryServerInterceptor is a GRPC interceptor that checks whether the
// request is authorized by the included macaroons.
func (i *GrpcInterceptor) MacaroonUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		// Check macaroons.
		if err := i.checkMacaroon(ctx, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// MacaroonStreamServerInterceptor is a GRPC interceptor that checks whether
// the request is authorized by the included macaroons.
func (i *GrpcInterceptor) MacaroonStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Check macaroons.
		err := i.checkMacaroon(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

// Adds the macaroon service provided to GrpcInterceptor struct attributes
func (i *GrpcInterceptor) AddMacaroonService(service *macaroons.Service) {
	i.svc = service
}
