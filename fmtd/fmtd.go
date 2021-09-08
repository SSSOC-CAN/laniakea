package fmtd

func Test_fmtd() {
	log := InitLogger(true)
	log.Info().Msg("Hello World!")
}
