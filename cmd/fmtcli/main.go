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
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"github.com/urfave/cli"
	"github.com/SSSOC-CAN/fmtd/fmtd"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/utils"
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
func getFmtClient(ctx *cli.Context) (fmtrpc.FmtClient, func()) {
	conn := getClientConn(ctx, false)
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewFmtClient(conn), cleanUp
}

//getDataCollectorClient returns the DataCollectorClient instance from the fmtrpc package with macaroon permissions and a cleanup function
func getDataCollectorClient(ctx *cli.Context) (fmtrpc.DataCollectorClient, func()) {
	conn := getClientConn(ctx, false)
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewDataCollectorClient(conn), cleanUp
}

//getUnlockerClient returns the UnlockerClient instance from the fmtrpc package as well as a cleanup function
func getUnlockerClient(ctx *cli.Context) (fmtrpc.UnlockerClient, func()) {
	conn := getClientConn(ctx, true)
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewUnlockerClient(conn), cleanUp
}

// getClientConn returns the grpc Client connection for use in instantiating FmtClient
func getClientConn(ctx *cli.Context, skipMacaroons bool) *grpc.ClientConn {
	config, err := fmtd.InitConfig()
	if err != nil {
		fatal(fmt.Errorf("Could not load config: %v", err))
	}
	//get TLS credentials from TLS certificate file
	creds, err := credentials.NewClientTLSFromFile(config.TLSCertPath, "")
	if err != nil {
		fatal(err)
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	if !skipMacaroons {
		// grab Macaroon data and load it into macaroon.Macaroon struct
		adminMac, err := os.ReadFile(config.AdminMacPath)
		if err != nil {
			fatal(fmt.Errorf("Could not read macaroon at %v: %v", config.AdminMacPath, err))
		}
		macHex := hex.EncodeToString(adminMac)
		mac, err := loadMacaroon(readPassword, macHex)
		if err != nil {
			fatal(fmt.Errorf("Could not load macaroon; %v", err))
		}
		// Add constraints to our macaroon
		macConstraints := []macaroons.Constraint{
			macaroons.TimeoutConstraint(defaultMacaroonTimeout), // prevent a replay attack
		}
		constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
		if err != nil {
			fatal(err)
		}
		cred := macaroons.NewMacaroonCredential(constrainedMac)
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	genericDialer := utils.ClientAddressDialer(strconv.FormatInt(config.GrpcPort, 10))
	opts = append(opts, grpc.WithContextDialer(genericDialer))
	opts = append(opts, grpc.WithDefaultCallOptions(maxMsgRecvSize))
	conn, err := grpc.Dial("localhost:"+strconv.FormatInt(config.GrpcPort, 10), opts...) // TODO:SSSOCPaulCote - Can't have hardcoded port here. Must somehow link config to this method
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