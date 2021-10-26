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
	"syscall"
	"github.com/urfave/cli"
	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	defaultMacaroonTimeout int64 = 60
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)
)

// fatal exits the process and prints out error information
func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[fmtcli] %v\n", err)
	os.Exit(1)
}

// getFmtClient returns the FmtClient instance from the fmtrpc package as well as a cleanup function
func getFmtClient() (fmtrpc.FmtClient, func()) {
	conn, err := auth.GetClientConn(false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewFmtClient(conn), cleanUp
}

//getDataCollectorClient returns the DataCollectorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getDataCollectorClient() (fmtrpc.DataCollectorClient, func()) {
	conn, err := auth.GetClientConn(false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewDataCollectorClient(conn), cleanUp
}

//getTestPlanExecutorClient returns the TestPlanExecutorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getTestPlanExecutorClient() (fmtrpc.TestPlanExecutorClient, func()) {
	conn, err := auth.GetClientConn(false, defaultMacaroonTimeout)
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewTestPlanExecutorClient(conn), cleanUp
}

//getUnlockerClient returns the UnlockerClient instance from the fmtrpc package as well as a cleanup function
func getUnlockerClient() (fmtrpc.UnlockerClient, func()) {
	conn, err := auth.GetClientConn(true, int64(0))
	if err != nil {
		fatal(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewUnlockerClient(conn), cleanUp
}

// main is the entrypoint for fmtcli
func main() {
	app := cli.NewApp()
	app.Name = "fmtcli"
	app.Usage = "Control panel for the Facility Management Tool Daemon (fmtd)"
	app.Commands = []cli.Command{
		stopCommand,
		adminTestCommand,
		testCommand,
		loginCommand,
		startRecording,
		stopRecording,
	}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}

// readPassword prompts the user for a password in the command line
func readPassword(text string) ([]byte, error) {
	fmt.Print(text)
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return pw, err
}