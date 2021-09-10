package fmtd

func Test_fmtd() {
	config := InitConfig()
	log := InitLogger(&config)
	log.Info().Msg("Hello World!")
}
