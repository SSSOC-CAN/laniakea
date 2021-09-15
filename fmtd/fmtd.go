package fmtd

import (
	"fmt"
	"github.com/SSSOC-CAN/fmtd/intercept"
)

// Main is the true entry point for fmtd. It's called in a nested manner for proper defer execution
func Main(interceptor *intercept.Interceptor, server *Server) error {
	// Starting main server
	err := server.Start()
	if err != nil {
		server.logger.Fatal().Msg("Could not start server")
		return err
	}
	server.logger.Debug().Msg(fmt.Sprintf("Server active: %v\tServer stopping: %v", server.Active, server.Stopping))
	defer server.Stop()

	// Starting RPC server
	rpcServer, err := NewRpcServer(interceptor, server.cfg)
	if err != nil {
		server.logger.Fatal().Msg("Could not initialize RPC server")
		return err
	}
	server.logger.Info().Msg("RPC Server Initialized")
	defer rpcServer.Grpc_server.Stop()
	err = rpcServer.Start()
	if err != nil {
		server.logger.Fatal().Msg("Could not start RPC server")
		return err
	}
	server.logger.Info().Msg("RPC Server Started")
	defer rpcServer.Stop()

	<-interceptor.ShutdownChannel()
	return nil
}
