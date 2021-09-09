package main

import (
	fmtd "github.com/SSSOC-CAN/fmtd/fmtd"
	//"github.com/rs/zerolog"
	//"os"
)

func main() {
	// fmtd.Test_fmtd()
	config := fmtd.InitConfig()
	log := fmtd.InitLogger(true, config)
	tcp_log := fmtd.NewSubLogger(log, "TCPS")
	tcp_log.Log("INFO", "Hello World")
	tcp_log.Log("TEST", "Hello World")
	//logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
