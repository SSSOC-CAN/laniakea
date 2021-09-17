package fmtd

import (
	"fmt"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"google.golang.org/grpc"
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

	// Starting badger kvdb
	db, err := badger.Open(badger.DefaultOptions(server.cfg.MacaroonDBDir))
	defer db.Close()

	// Starting RPC server
	rpcServer, err := NewRpcServer(interceptor, server.cfg, server.logger)
	if err != nil {
		server.logger.Fatal().Msg("Could not initialize RPC server")
		return err
	}
	server.logger.Info().Msg("RPC Server Initialized")

	// Creating gRPC server and Server options
	grpc_interceptor := intercept.NewGrpcInterceptor(rpcServer.SubLogger, true)
	grpc_server := grpc.NewServer(grpc_interceptor.CreateGrpcOptions()...)
	rpcServer.AddGrpcServer(grpc_server)
	defer rpcServer.GrpcServer.Stop()

	//Starting RPC and gRPC Servers
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
