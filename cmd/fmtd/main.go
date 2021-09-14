package main

import (
	"fmt"
	"github.com/SSSOC-CAN/fmtd/fmtd"
	"github.com/SSSOC-CAN/fmtd/intercept"
)

func main() {
	shutdownInterceptor, err := intercept.InitInterceptor()
	config := fmtd.InitConfig()
	log := fmtd.InitLogger(&config)
	server, err := fmtd.InitServer(&config, &log)
	if err != nil {
		log.Fatal().Msg("Could not initialize server")
	}
	err = server.Start()
	if err != nil {
		log.Fatal().Msg("Could not start server")
	}
	log.Debug().Msg(fmt.Sprintf("Server active: %v\tServer stopping: %v", server.Active, server.Stopping))
	defer server.Stop()
	rpcServer, err := fmtd.NewRpcServer(shutdownInterceptor, &config)
	if err != nil {
		log.Fatal().Msg("Could not initialize RPC server")
	}
	log.Info().Msg("RPC Server Initialized")
	defer rpcServer.Grpc_server.Stop()
	err = rpcServer.Start()
	if err != nil {
		log.Fatal().Msg("Could not start RPC server")
	}
	log.Info().Msg("RPC Server Started")
	defer rpcServer.Stop()
	<-shutdownInterceptor.ShutdownChannel()
}