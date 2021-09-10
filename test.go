package main

import (
	fmtd "github.com/SSSOC-CAN/fmtd/fmtd"
)

func main() {
	// fmtd.Test_fmtd()
	config := fmtd.InitConfig()
	log := fmtd.InitLogger(&config)
	tcp_log := fmtd.NewSubLogger(&log, "TCPS")
	tcp_log.Log("INFO", "Hello World")
	tcp_log.Log("TEST", "Hello World")
}
