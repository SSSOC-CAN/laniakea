package main

import (
	"fmt"
	fmtd "github.com/SSSOC-CAN/fmtd/fmtd"
)

func main() {
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
	err = server.Stop()
	if err != nil {
		log.Fatal().Msg("Could not stop server")
	}
	log.Debug().Msg(fmt.Sprintf("Server active: %v\tServer stopping: %v", server.Active, server.Stopping))
}
