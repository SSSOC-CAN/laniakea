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
		Multiline:     true,
		UseProtoNames: true,
		Indent:        "    ",
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

var startRecording = cli.Command{
	Name:      "start-record",
	Usage:     "Start recording data in realtime.",
	ArgsUsage: "service-name",
	Description: `
	This command starts the process of recording data for the given service into a timestamped csv file. Service-name must be one the following:
	    - telemetry
		- rga`,
	Flags: []cli.Flag{
		cli.Int64Flag{
			Name:  "polling_interval",
			Usage: "If set, then the time between polls to the DAQ will be set to the specified amount in seconds.",
		},
		cli.StringFlag{
			Name:  "org_name",
			Usage: "Required. Name of the influx organization",
		},
		cli.StringFlag{
			Name:  "record_name",
			Usage: "Required for telemetry. Name given to the recording",
		},
	},
	Action: startRecord,
}

// startRecord is the CLI wrapper around the TelemetryService StartRecording method
func startRecord(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 || ctx.NumFlags() > 3 || ctx.String("org_name") == "" {
		return cli.ShowCommandHelp(ctx, "start-record")
	}
	client, cleanUp := getDataCollectorClient(ctx)
	defer cleanUp()
	var polling_interval int64
	serviceNameStr := ctx.Args().First()
	if serviceNameStr == "telemetry" && ctx.String("record_name") == "" {
		return cli.ShowCommandHelp(ctx, "start-record")
	}
	var serviceName fmtrpc.RecordService
	if ctx.NumFlags() != 0 {
		polling_interval = ctx.Int64("polling_interval")
	}
	switch serviceNameStr {
	case "telemetry":
		serviceName = fmtrpc.RecordService_TELEMETRY
	case "rga":
		serviceName = fmtrpc.RecordService_RGA
	default:
		return ErrInvalidServiceName
	}
	recordRequest := &fmtrpc.RecordRequest{
		PollingInterval: polling_interval,
		Type:            serviceName,
		OrgName:         ctx.String("org_name"),
		BucketName:      ctx.String("record_name"),
	}
	recordResponse, err := client.StartRecording(ctxc, recordRequest)
	if err != nil {
		return err
	}
	printRespJSON(recordResponse)
	return nil
}

var stopRecording = cli.Command{
	Name:      "stop-record",
	Usage:     "Stop recording data.",
	ArgsUsage: "service-name",
	Description: `
	This command stops the process of recording data from the given service into a timestamped csv file. Service-name must be one the following:
	- telemetry
	- rga`,
	Action: stopRecord,
}

// stopRecord is the CLI wrapper around the TelemetryService StopRecording method
func stopRecord(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 || ctx.NumFlags() > 1 {
		return cli.ShowCommandHelp(ctx, "stop-record")
	}
	client, cleanUp := getDataCollectorClient(ctx)
	defer cleanUp()
	serviceNameStr := ctx.Args().First()
	var serviceName fmtrpc.RecordService
	switch serviceNameStr {
	case "telemetry":
		serviceName = fmtrpc.RecordService_TELEMETRY
	case "rga":
		serviceName = fmtrpc.RecordService_RGA
	default:
		return ErrInvalidServiceName
	}
	recordResponse, err := client.StopRecording(ctxc, &fmtrpc.StopRecRequest{Type: serviceName})
	if err != nil {
		return err
	}
	printRespJSON(recordResponse)
	return nil
}

var loadTestPlan = cli.Command{
	Name:      "load-testplan",
	Usage:     "Load a test plan file.",
	ArgsUsage: "path-to-testplan",
	Description: `
	This command loads a given test plan file. The absolute path must be given. Ex: C:\Users\User\Downloads\testplan.yaml`,
	Action: loadPlan,
}

// loadPlan is the CLI wrapper for the Test Plan Executor method LoadTestPlan
func loadPlan(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "load-testplan")
	}
	client, cleanUp := getTestPlanExecutorClient(ctx)
	defer cleanUp()
	testPlanFilePath := ctx.Args().First()
	loadPlanReq := &fmtrpc.LoadTestPlanRequest{
		PathToFile: testPlanFilePath,
	}
	loadPlanResponse, err := client.LoadTestPlan(ctxc, loadPlanReq)
	if err != nil {
		return err
	}
	printRespJSON(loadPlanResponse)
	return nil
}

var startTestPlan = cli.Command{
	Name:  "start-testplan",
	Usage: "Start a loaded test plan file.",
	Description: `
	This command starts a loaded test plan file. If no test plan has been loaded, an error will be raised.`,
	Action: startPlan,
}

// startPlan is the CLI wrapper for the Test Plan Executor method StartTestPlan
func startPlan(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getTestPlanExecutorClient(ctx)
	defer cleanUp()
	startPlanResponse, err := client.StartTestPlan(ctxc, &fmtrpc.StartTestPlanRequest{})
	if err != nil {
		return err
	}
	printRespJSON(startPlanResponse)
	return nil
}

var stopTestPlan = cli.Command{
	Name:  "stop-testplan",
	Usage: "Stops a currently executing test plan",
	Description: `
	This command stops a running test plan. If no test plan is being executed, an error will be raised.`,
	Action: stopPlan,
}

// stopPlan is the CLI wrapper for the Test Plan Executor method StopTestPlan
func stopPlan(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getTestPlanExecutorClient(ctx)
	defer cleanUp()
	stopPlanResponse, err := client.StopTestPlan(ctxc, &fmtrpc.StopTestPlanRequest{})
	if err != nil {
		return err
	}
	printRespJSON(stopPlanResponse)
	return nil
}

var insertROIMarker = cli.Command{
	Name:      "insert-roi",
	Usage:     "Inserts a Region of Interest marker in the test plan report",
	ArgsUsage: "message",
	Description: `
	This command inserts a Region of Interest marker in the test plan report. If no test plan is currently running, an error is raised.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "report_lvl",
			Usage: `If set, then the entries level of importance in the report file will be marked accordingly. Must be one of the following:
			- INFO
			- ERROR
			- DEBUG
			- WARN
			- FATAL
			INFO by default.`,
		},
		cli.StringFlag{
			Name:  "author",
			Usage: "If set, the author of the entry will be marked accordingly. FMT by default.",
		},
	},
	Action: insertROI,
}

// insertROI is the CLI wrapper for the Test Plan Executor method InsertROIMarker
func insertROI(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 || ctx.NumFlags() > 2 {
		return cli.ShowCommandHelp(ctx, "insert-roi")
	}
	client, cleanUp := getTestPlanExecutorClient(ctx)
	defer cleanUp()
	markerText := ctx.Args().First()
	reportLvl := fmtrpc.ReportLvl_INFO
	author := "FMT"
	if ctx.String("report_lvl") != "" {
		switch ctx.String("report_lvl") {
		case "INFO":
			reportLvl = fmtrpc.ReportLvl_INFO
		case "ERROR":
			reportLvl = fmtrpc.ReportLvl_ERROR
		case "DEBUG":
			reportLvl = fmtrpc.ReportLvl_DEBUG
		case "WARN":
			reportLvl = fmtrpc.ReportLvl_WARN
		case "FATAL":
			reportLvl = fmtrpc.ReportLvl_FATAL
		default:
			return cli.ShowCommandHelp(ctx, "insert-roi")
		}
	}
	if ctx.String("author") != "" {
		author = ctx.String("author")
	}
	insertROIReq := &fmtrpc.InsertROIRequest{
		Text:      markerText,
		ReportLvl: reportLvl,
		Author:    author,
	}
	insertROIResponse, err := client.InsertROIMarker(ctxc, insertROIReq)
	if err != nil {
		return err
	}
	printRespJSON(insertROIResponse)
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
	},
	Action: bakeMac,
}

func bakeMac(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() < 1 || ctx.NumFlags() > 3 {
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
	bakeMacReq := &fmtrpc.BakeMacaroonRequest{
		Timeout:     timeout,
		TimeoutType: timeoutType,
		Permissions: parsedPerms,
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
