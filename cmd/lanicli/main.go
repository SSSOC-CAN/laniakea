/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20

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

	"github.com/SSSOC-CAN/laniakea/auth"
	"github.com/SSSOC-CAN/laniakea/lanirpc"
	"github.com/SSSOC-CAN/laniakea/utils"
	"github.com/urfave/cli"
)

var (
	defaultRPCAddr               = "localhost"
	defaultRPCPort               = "7777"
	defaultTLSCertFilename       = "tls.cert"
	defaultFmtdDir               = utils.AppDataDir("laniakea", false)
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
	fmt.Fprintf(os.Stderr, "[lanicli] %v\n", err)
	os.Exit(1)
}

// getLaniClient returns the LaniClient instance from the lanirpc package as well as a cleanup function
func getLaniClient(ctx *cli.Context) (lanirpc.LaniClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return lanirpc.NewLaniClient(conn), cleanUp
}

//getUnlockerClient returns the UnlockerClient instance from the lanirpc package as well as a cleanup function
func getUnlockerClient(ctx *cli.Context) (lanirpc.UnlockerClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, true, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return lanirpc.NewUnlockerClient(conn), cleanUp
}

// getPluginAPIClient returns the PluginAPIClient instance and a cleanup function
func getPluginAPIClient(ctx *cli.Context) (lanirpc.PluginAPIClient, func()) {
	args := extractArgs(ctx)
	conn, err := auth.GetClientConn(args.RPCAddr, args.RPCPort, args.TLSCertPath, args.AdminMacPath, false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return lanirpc.NewPluginAPIClient(conn), cleanUp
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

// main is the entrypoint for lanicli
func main() {
	app := cli.NewApp()
	app.Name = "lanicli"
	app.Usage = "Control panel for the Laniakea daemon (laniakea)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpc_addr",
			Value: defaultRPCAddr,
			Usage: "The host address of the Laniakea daemon (exclude the port)",
		},
		cli.StringFlag{
			Name:  "rpc_port",
			Value: defaultRPCPort,
			Usage: "The host port of the Laniakea daemon",
		},
		cli.StringFlag{
			Name:      "tlscertpath",
			Value:     defaultTLSCertPath,
			Usage:     "The path to Laniakea's TLS certificate.",
			TakesFile: true,
		},
		cli.StringFlag{
			Name:      "macaroonpath",
			Value:     defaultMacPath,
			Usage:     "The path to Laniakea's macaroons.",
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
		bakeMacaroon,
		pluginStartRecordCmd,
		pluginStopRecordCmd,
		pluginStartCmd,
		pluginStopCmd,
		pluginCommandCmd,
		addPluginCmd,
		listPluginsCmd,
		getPluginCmd,
	}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}
