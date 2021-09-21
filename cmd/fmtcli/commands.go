package main

import (
	"context"
	"fmt"
	"os"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/urfave/cli"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	client, cleanUp := getClient(ctx) //This command returns the proto generated FmtClient instance
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
	client, cleanUp := getClient(ctx)
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
	client, cleanUp := getClient(ctx)
	defer cleanUp()
	testResp, err := client.TestCommand(ctxc, &fmtrpc.TestRequest{})
	if err != nil {
		return err
	}
	printRespJSON(testResp)
	return nil
}