/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20
*/

package main

import (
	"io"
	"time"

	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/SSSOC-CAN/laniakea/lanirpc"
	"github.com/urfave/cli"
)

var pluginStartRecordCmd = cli.Command{
	Name:      "plugin-startrecord",
	Usage:     "Start recording data from a given datasource plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Start the recording process of a given datasource plugin`,
	Action: pluginStartRecord,
}

// pluginStartRecord is the proxy command between lanicli and the gRPC equivalent
func pluginStartRecord(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "plugin-startrecord")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	startRecRequest := &lanirpc.PluginRequest{
		Name: pluginName,
	}
	resp, err := client.StartRecord(ctxc, startRecRequest)
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}

var pluginStopRecordCmd = cli.Command{
	Name:      "plugin-stoprecord",
	Usage:     "Stop recording data from a given datasource plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Stop the recording process of a given datasource plugin`,
	Action: pluginStopRecord,
}

// pluginStopRecord is the proxy command between lanicli and the gRPC equivalent
func pluginStopRecord(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "plugin-stoprecord")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	stopRecRequest := &lanirpc.PluginRequest{
		Name: pluginName,
	}
	resp, err := client.StopRecord(ctxc, stopRecRequest)
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}

var pluginStartCmd = cli.Command{
	Name:      "plugin-start",
	Usage:     "Start a given plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Start a given plugin. The plugin must be registered to start it.`,
	Action: pluginStart,
}

// pluginStart is the proxy command between lanicli and the gRPC equivalent
func pluginStart(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "plugin-start")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	startRequest := &lanirpc.PluginRequest{
		Name: pluginName,
	}
	resp, err := client.StartPlugin(ctxc, startRequest)
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}

var pluginStopCmd = cli.Command{
	Name:      "plugin-stop",
	Usage:     "Stop a given plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Stop a given plugin. The plugin will stop regardless of errors when safely stopping.`,
	Action: pluginStop,
}

// pluginStop is the proxy command between lanicli and the gRPC equivalent
func pluginStop(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "plugin-stop")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	stopRequest := &lanirpc.PluginRequest{
		Name: pluginName,
	}
	resp, err := client.StopPlugin(ctxc, stopRequest)
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}

var pluginCommandCmd = cli.Command{
	Name:      "plugin-command",
	Usage:     "Sends a command to a controller plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Sends a given controller plugin a command. The command is converted into bytes.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "command",
			Usage: "Required. The actual command being sent to the plugin. The command is then converted to bytes",
		},
		cli.StringFlag{
			Name:  "frametype",
			Usage: "The type of data being passed to the plugin. Types are usually MIME like strings (e.g. application/json). Default is application/string",
		},
	},
	Action: pluginCommand,
}

// pluginCommand is the proxy command between lanicli and the gRPC equivalent
func pluginCommand(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 || ctx.NumFlags() == 0 {
		return cli.ShowCommandHelp(ctx, "plugin-command")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	frameType := "application/string"
	if frType := ctx.String("frametype"); frType != "" {
		frameType = frType
	}
	cmdRequest := &lanirpc.ControllerPluginRequest{
		Name: pluginName,
		Frame: &proto.Frame{
			Source:    "lanicli",
			Type:      frameType,
			Timestamp: time.Now().UnixMilli(),
			Payload:   []byte(ctx.String("command")),
		},
	}
	stream, err := client.Command(ctxc, cmdRequest)
	if err != nil {
		return err
	}
	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	printRespJSON(resp)
	for {
		resp, err := stream.Recv()
		if err == io.EOF || resp == nil {
			break
		}
		if err != nil {
			return err
		}
		printRespJSON(resp)
	}
	return nil
}

var addPluginCmd = cli.Command{
	Name:      "plugin-add",
	Usage:     "Registers a new plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Registers and starts a new plugin. The plugin executable must exist in the plugin directory.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type",
			Usage: "Required. Choose from either datasource or controller.",
		},
		cli.StringFlag{
			Name:  "exec",
			Usage: "Required. The file name of the plugin executable",
		},
		cli.Int64Flag{
			Name:  "timeout",
			Usage: "The time, in seconds, for an unresponsive plugin to be considered timed out. Default is 30 seconds.",
		},
		cli.Int64Flag{
			Name:  "maxtimeouts",
			Usage: "The number of times a plugin can timeout before being killed. Default is 3.",
		},
	},
	Action: addPlugin,
}

// addPlugin is the proxy command between lanicli and the gRPC equivalent
func addPlugin(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "plugin-add")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	plugType := ctx.String("type")
	if plugType != "datasource" && plugType != "controller" {
		return cli.ShowCommandHelp(ctx, "plugin-add")
	}
	execName := ctx.String("exec")
	if execName == "" {
		return cli.ShowCommandHelp(ctx, "plugin-add")
	}
	addRequest := &lanirpc.PluginConfig{
		Name:        pluginName,
		Type:        plugType,
		ExecName:    execName,
		Timeout:     ctx.Int64("timeout"),
		MaxTimeouts: ctx.Int64("maxtimeouts"),
	}
	resp, err := client.AddPlugin(ctxc, addRequest)
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}

var listPluginsCmd = cli.Command{
	Name:  "plugin-list",
	Usage: "Returns a list of all registered plugins",
	Description: `
	Returns a list of all registered plugin which includes plugin state, type, start time, stop time, version number, etc.`,
	Action: listPlugins,
}

// listPlugins is the proxy command between lanicli and the gRPC equivalent
func listPlugins(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	resp, err := client.ListPlugins(ctxc, &proto.Empty{})
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}

var getPluginCmd = cli.Command{
	Name:      "plugin-info",
	Usage:     "Returns the info of a given plugin",
	ArgsUsage: "plugin-name",
	Description: `
	Returns the info of a given plugin which includes plugin state, type, start time, stop time, version number, etc.`,
	Action: getPlugin,
}

// getPlugin is the proxy command between lanicli and the gRPC equivalent
func getPlugin(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "plugin-info")
	}
	client, cleanUp := getPluginAPIClient(ctx)
	defer cleanUp()
	pluginName := ctx.Args().First()
	getRequest := &lanirpc.PluginRequest{
		Name: pluginName,
	}
	resp, err := client.GetPlugin(ctxc, getRequest)
	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}
