/*
Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/urfave/cli"
	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
)

var (
	defaultRPCAddr               = "localhost"
	defaultRPCPort               = "3567"
	defaultTLSCertFilename       = "tls.cert"
	defaultFmtdDir           	 = utils.AppDataDir("fmtd", false)
	defaultTLSCertPath           = filepath.Join(defaultFmtdDir, defaultTLSCertFilename)
	defaultMacaroonTimeout int64 = 60
	defaultAdminMacName          = "admin.macaroon"
	defaultMacPath               = filepath.Join(defaultFmtdDir, defaultAdminMacName)
)

type Args struct {
	RPCAddr      string
	RPCPort      string
	TLSCertPath  string
	AdminMacPath string
}

// fatal exits the process and prints out error information
func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[fmtcli] %v\n", err)
	os.Exit(1)
}

// getFmtClient returns the FmtClient instance from the fmtrpc package as well as a cleanup function
func getFmtClient(ctx *cli.Context) (fmtrpc.FmtClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewFmtClient(conn), cleanUp
}

//getDataCollectorClient returns the DataCollectorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getDataCollectorClient(ctx *cli.Context) (fmtrpc.DataCollectorClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewDataCollectorClient(conn), cleanUp
}

//getTestPlanExecutorClient returns the TestPlanExecutorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getTestPlanExecutorClient(ctx *cli.Context) (fmtrpc.TestPlanExecutorClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewTestPlanExecutorClient(conn), cleanUp
}

//getUnlockerClient returns the UnlockerClient instance from the fmtrpc package as well as a cleanup function
func getUnlockerClient(ctx *cli.Context) (fmtrpc.UnlockerClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, true, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewUnlockerClient(conn), cleanUp
}

// extractArgs extracts the arguments inputted to the heartcli command
func extractArgs(ctx *cli.Context) *Args {
	return &Args{
		RPCAddr:      ctx.GlobalString("rpc_addr"),
		RPCPort:      ctx.GlobalString("rpc_port"),
		TLSCertPath:  ctx.GlobalString("tlscertpath"),
		AdminMacPath: ctx.GlobalString("macaroonpath"),
	}
}

// main is the entrypoint for fmtcli
func main() {
	app := cli.NewApp()
	app.Name = "fmtcli"
	app.Usage = "Control panel for the Facility Management Tool daemon (fmtd)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpc_addr",
			Value: defaultRPCAddr,
			Usage: "The host address of the Facility Management Tool daemon (exclude the port)",
		},
		cli.StringFlag{
			Name:  "rpc_port",
			Value: defaultRPCPort,
			Usage: "The host port of the Facility Management Tool daemon",
		},
		cli.StringFlag{
			Name:      "tlscertpath",
			Value:     defaultTLSCertPath,
			Usage:     "The path to fmtd's TLS certificate.",
			TakesFile: true,
		},
		cli.StringFlag{
			Name:      "macaroonpath",
			Value:     defaultMacPath,
			Usage:     "The path to fmtd's macaroons.",
			TakesFile: true,
		},
	}
	app.Commands = []cli.Command{
		stopCommand,
		adminTestCommand,
		testCommand,
		loginCommand,
		setPassword,
		changePassword,
		startRecording,
		stopRecording,
		loadTestPlan,
		startTestPlan,
		stopTestPlan,
		insertROIMarker,
		bakeMacaroon,
	}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}