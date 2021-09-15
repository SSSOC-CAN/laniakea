package fmtd

// main this file may be deleted
func main() {
	config := InitConfig()
	log := InitLogger(&config)
	server, err := InitServer(&config, &log)
	if err != nil {
		log.Fatal().Msg("Could not initialize server")
	}
	err = server.Start()
	if err != nil {
		log.Fatal().Msg("Could not start server")
	}
	err = server.Stop()
	if err != nil {
		log.Fatal().Msg("Could not stop server")
	}
}
