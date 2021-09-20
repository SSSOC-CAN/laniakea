package main

import (
	"fmt"
	"os"
	"github.com/urfave/cli"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"google.golang.org/grpc"
)
//TODO:SSSOCPaulCote - Get the TLS certificate and feed it to grpc connection
// fatal exits the process and prints out error information
func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[fmtcli] %v\n", err)
	os.Exit(1)
}

// getClient returns the FmtClient instance from the fmtrpc package as well as a cleanup function
func getClient(ctx *cli.Context) (fmtrpc.FmtClient, func()) {
	conn := getClientConn(ctx)
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewFmtClient(conn), cleanUp
}

// getClientConn returns the grpc Client connection for use in instantiating FmtClient
func getClientConn(ctx *cli.Context) *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(":3567", opts...) // TODO:SSSOCPaulCote - Can't have hardcoded port here. Must somehow link config to this method
	if err != nil {
		fatal(fmt.Errorf("Unable to connect to RPC server: %v", err))
	}
	return conn
}

// main is the entrypoint for fmtcli
func main() {
	app := cli.NewApp()
	app.Name = "fmtcli"
	app.Usage = "Control panel for the Facility Management Tool Daemon (fmtd)"
	app.Commands = []cli.Command{
		stopCommand,
	}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}