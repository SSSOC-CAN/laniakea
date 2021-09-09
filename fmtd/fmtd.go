package fmtd

func Test_fmtd() {
	config := InitConfig()
	log := InitLogger(true, &config)
	log.Info().Msg("Hello World!")
}
