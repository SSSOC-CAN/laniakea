/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10

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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	e "github.com/pkg/errors"
	"github.com/urfave/cli"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	ErrInvalidServiceName     = bg.Error("invalid service name. service must be either telemetry or rga")
	ErrParsingPermissionTuple = bg.Error("unable to parse permission tuple")
	ErrEmptyAction            = bg.Error("action cannot be empty")
	ErrEmptyEntity            = bg.Error("entity cannot be empty")
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
		Multiline:       true,
		UseProtoNames:   true,
		Indent:          "    ",
		EmitUnpopulated: true,
	}
	jsonStr := jsonMarshaler.Format(resp)
	fmt.Println(jsonStr)
}

var stopCommand = cli.Command{
	Name:  "stop",
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
	Name:  "admin-test",
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
	Name:  "test",
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

var bakeMacaroon = cli.Command{
	Name:      "bake-macaroon",
	Usage:     "Bakes a new macaroon",
	ArgsUsage: "permissions",
	Description: `
	This command bakes a new macaroon based on the provided permissions and optional constraints. Permissions must be of the following format: entity:action
	For specific commands, the format is uri:<command_uri> example: fmtd:write, uri:/fmtrpc.Fmt/Test this would generate a macaroon with write permissions for any commands associated to the fmtd entity as well as the fmtrpc.Fmt.Test command
	`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "timeout_unit",
			Usage: `If set, then the the timeout given will be in the specified units. Possible units are:
			- day
			- hour
			- minute
			- second
			second is the default.`,
		},
		cli.Int64Flag{
			Name:  "timeout",
			Usage: "If set, the macaroon will timeout after the given number of units. The default unit is seconds. The default timeout is 24 hours",
		},
		cli.StringFlag{
			Name:  "save_to",
			Usage: "If set, the macaroon will be saved as a .macaroon file at the specified path. The hex encoded macaroon is printed to the console by default",
		},
		cli.StringFlag{
			Name:  "plugins",
			Usage: "If set, the macaroon will be locked to specified plugins. Enter plugin names, spaced with a colon (e.g. plugin-1:plugin_TWO:PlUgIn-_3Thousand)",
		},
	},
	Action: bakeMac,
}

func bakeMac(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() < 1 || ctx.NumFlags() > 4 {
		return cli.ShowCommandHelp(ctx, "bake-macaroon")
	}
	client, cleanUp := getFmtClient(ctx)
	defer cleanUp()
	permissions := ctx.Args()
	timeoutType := fmtrpc.TimeoutType_SECOND
	timeout := ctx.Int64("timeout")
	if ttype := ctx.String("timeout_type"); ttype != "" {
		switch ttype {
		case "day":
			timeoutType = fmtrpc.TimeoutType_DAY
		case "hour":
			timeoutType = fmtrpc.TimeoutType_HOUR
		case "minute":
			timeoutType = fmtrpc.TimeoutType_MINUTE
		}
	}
	var parsedPerms []*fmtrpc.MacaroonPermission
	for _, perm := range permissions {
		tuple := strings.Split(perm, ":")
		if len(tuple) != 2 {
			return ErrParsingPermissionTuple
		}
		entity, action := tuple[0], tuple[1]
		if entity == "" {
			return e.Wrap(ErrEmptyEntity, "invalid permission")
		}
		if action == "" {
			return e.Wrap(ErrEmptyAction, "invalid permission")
		}
		parsedPerms = append(parsedPerms, &fmtrpc.MacaroonPermission{
			Entity: entity,
			Action: action,
		})
	}
	var parsedPlugins []string
	if pluginNames := ctx.String("plugins"); pluginNames != "" {
		parsedPlugins = strings.Split(pluginNames, ":")
	}

	bakeMacReq := &fmtrpc.BakeMacaroonRequest{
		Timeout:     timeout,
		TimeoutType: timeoutType,
		Permissions: parsedPerms,
		Plugins:     parsedPlugins,
	}
	bakeMacResponse, err := client.BakeMacaroon(ctxc, bakeMacReq)
	if err != nil {
		return err
	}
	if path := ctx.String("save_to"); path != "" {
		macBytes, err := hex.DecodeString(bakeMacResponse.Macaroon)
		if err != nil {
			return err
		}
		if !utils.FileExists(path) {
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			defer file.Close()
		}
		err = ioutil.WriteFile(path, macBytes, 0755)
		if err != nil {
			_ = os.Remove(path)
			return err
		}
	}
	printRespJSON(bakeMacResponse)
	return nil
}
