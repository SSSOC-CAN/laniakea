package main

import (
	"fmt"
	fmtd "github.com/SSSOC-CAN/fmtd/fmtd"
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
	grpcServer, err := fmtd.NewGrpcServer(shutdownInterceptor)
	if err != nil {
		log.Fatal().Msg("Could not initialize RPC server")
	}
	log.Info().Msg("RPC Server Initialized")
	err = grpcServer.Start("7777")
	if err != nil {
		log.Fatal().Msg("Could not start RPC server")
	}
	log.Info().Msg("RPC Server Started")
	// err = server.Stop()
	// if err != nil {
	// 	log.Fatal().Msg("Could not stop server")
	// }
	// log.Debug().Msg(fmt.Sprintf("Server active: %v\tServer stopping: %v", server.Active, server.Stopping))
}
