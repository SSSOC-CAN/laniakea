package main

import (
	"fmt"
	"os"
	"github.com/SSSOC-CAN/fmtd/fmtd"
	"github.com/SSSOC-CAN/fmtd/intercept"
)

// main is the entry point for fmt daemon.
// TODO:SSSOCPaulCote - Refactor some of this into server.Start()
func main() {
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	config, err := fmtd.InitConfig()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	log, err := fmtd.InitLogger(&config)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	shutdownInterceptor.Logger = &log
	server, err := fmtd.InitServer(&config, &log)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = fmtd.Main(shutdownInterceptor, server); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	
}