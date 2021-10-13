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
	"context"
	"fmt"
	"os"
	"syscall"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	defaultPollingInterval int64 = 10
)

// getContext spins up a go routine to monitor for shutdown requests and returns a context object
func getContext() context.Context {
	shutdownInterceptor, err := intercept.InitInterceptor()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctxc, cancel := context.WithCancel(context.Background())
	go func() {
		<-shutdownInterceptor.ShutdownChannel()
		cancel()
	}()
	return ctxc
}

// printRespJSON will convert a proto response as a string and print it
func printRespJSON(resp proto.Message) {
	jsonMarshaler := &protojson.MarshalOptions{
		Multiline: true,
		UseProtoNames: true,
		Indent: "    ",
	}
	jsonStr := jsonMarshaler.Format(resp)
	// if err != nil {
	// 	fmt.Println("Unable to decode response: ", err)
	// 	return
	// }
	fmt.Println(jsonStr)
}

var stopCommand = cli.Command{
	Name: "stop",
	Usage: "Stop and shutdown the daemon",
	Description: `
	Gracefully stop all daemon subprocesses before stopping the daemon itself. This is equivalent to stopping it using CTRL-C.`,
	Action: stopDaemon,
}

// stopDaemon is the proxy command between fmtcli and gRPC equivalent. 
func stopDaemon(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getFmtClient(ctx) //This command returns the proto generated FmtClient instance
	defer cleanUp()

	_, err := client.StopDaemon(ctxc, &fmtrpc.StopRequest{})
	if err != nil {
		return err
	}
	return nil
}

var adminTestCommand = cli.Command{
	Name: "admin-test",
	Usage: "Test command for the admin macaroon",
	Description: `
	A test command which returns a string only if the admin macaroon is provided.`,
	Action: adminTest,
}

// Proxy command for the fmtcli
func adminTest(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getFmtClient(ctx)
	defer cleanUp()
	testResp, err := client.AdminTest(ctxc, &fmtrpc.AdminTestRequest{})
	if err != nil {
		return err
	}
	printRespJSON(testResp)
	return nil
}

var testCommand = cli.Command{
	Name: "test",
	Usage: "Test command",
	Description: `
	A test command which returns a string for any macaroon provided.`,
	Action: test,
}

// Proxy command for the fmtcli
func test(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getFmtClient(ctx)
	defer cleanUp()
	testResp, err := client.TestCommand(ctxc, &fmtrpc.TestRequest{})
	if err != nil {
		return err
	}
	printRespJSON(testResp)
	return nil
}

// readPasswordFromTerminal will prompt the user on the terminal with the msg string and return the response
func readPasswordFromTerminal(msg string) ([]byte, error) {
	fmt.Print(msg)
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return pw, err
}

var loginCommand = cli.Command{
	Name: "login",
	Usage: "Login into the FMTD",
	Description: `
	When invoked, user will be prompted to enter a password. Either create one or use an existing one.`,
	Action: login,
}

// login is the wrapper around the UnlockerClient.Login method
func login(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getUnlockerClient(ctx)
	defer cleanUp()
	pw, err := readPasswordFromTerminal("Input password: ")
	if err != nil {
		return err
	}
	loginReq := &fmtrpc.LoginRequest{
		Password: pw,
	}
	loginResp, err := client.Login(ctxc, loginReq)
	if err != nil {
		return err
	}
	printRespJSON(loginResp)
	return nil
}

var startRecording = cli.Command{
	Name: "start-record",
	Usage: "Start recording data in realtime.",
	Description: `
	This command starts the process of recording data from the Fluke DAQ into a timestamped csv file.`,
	Flags: []cli.Flag{
		cli.Int64Flag{
			Name: "polling_interval",
			Usage: "If set, then the time between polls to the DAQ will be set to the specified amount in seconds.",
		},
	},
	Action: startRecord,
}

// startRecord is the CLI wrapper around the FlukeService StartRecording method
func startRecord(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getDataCollectorClient(ctx)
	defer cleanUp()
	var polling_interval int64
	if ctx.NumFlags() == 0 {
		polling_interval = defaultPollingInterval
	} else {
		polling_interval = ctx.Int64("polling_interval")
	}
	recordRequest := &fmtrpc.RecordRequest{
		PollingInterval: polling_interval,
	}
	recordResponse, err := client.StartRecording(ctxc, recordRequest)
	if err != nil {
		return err
	}
	printRespJSON(recordResponse)
	return nil
}

var stopRecording = cli.Command{
	Name: "stop-record",
	Usage: "Stop recording data.",
	Description: `
	This command stops the process of recording data from the Fluke DAQ into a timestamped csv file.`,
	Action: stopRecord,
}

// stopRecord is the CLI wrapper around the FlukeService StopRecording method
func stopRecord(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getDataCollectorClient(ctx)
	defer cleanUp()
	recordResponse, err := client.StopRecording(ctxc, &fmtrpc.StopRecRequest{})
	if err != nil {
		return err
	}
	printRespJSON(recordResponse)
	return nil
}