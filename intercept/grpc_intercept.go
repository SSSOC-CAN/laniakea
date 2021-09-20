package intercept

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	macaroonWhitelist = map[string]struct{}{
		"/fmtrpc.Fmt/StopDaemon":        {},
	}
)

// GrpcInteceptor struct is a data structure with attributes relevant to creating the gRPC interceptor
type GrpcInterceptor struct {
	noMacaroons bool
	log *zerolog.Logger
	permissionMap map[string][]bakery.Op
	svc *macaroons.Service
}

// NewGrpcInterceptor instantiates a new GrpcInterceptor struct
func NewGrpcInterceptor(log *zerolog.Logger, noMacaroons bool) *GrpcInterceptor {
	return &GrpcInterceptor{
		noMacaroons: noMacaroons,
		permissionMap: make(map[string][]bakery.Op),
		log: log,
	}
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

// AddPermission adds a new macaroon rule for the given method
func (i *GrpcInterceptor) AddPermission(method string, ops []bakery.Op) error {
	if _, ok := i.permissionMap[method]; ok {
		return fmt.Errorf("Detected duplicate macaroon constraints for path: %v", method)
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

	// r.RLock()
	svc := i.svc
	// r.RUnlock()

	// If the macaroon service is not yet active, we cannot allow
	// the call.
	if svc == nil {
		return fmt.Errorf("unable to determine macaroon permissions")
	}

	// r.RLock()
	uriPermissions, ok := i.permissionMap[fullMethod]
	// r.RUnlock()
	if !ok {
		return fmt.Errorf("%s: unknown permissions required for method",
			fullMethod)
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

