// +build demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package main

import (
	"io"
	"strconv"

	"github.com/SSSOC-CAN/fmtd/fmtrpc/demorpc"
	"github.com/urfave/cli"
)

var setTempCommand = cli.Command{
	Name:  "set-temperature",
	Usage: "Set a temperature set point",
	ArgsUsage: "temp-set-point",
	Description: `
	Set a temperature set point (in degrees Celsius) to change the current average temperature.`,
	Flags: []cli.Flag{
		cli.Float64Flag{
			Name:  "temp_change_rate",
			Usage: "The rate at which the temperature changes (degrees Celsius/minute). If not set, defaults to instant change.",
		},
	},
	Action: setTemp,
}

// setTemp is the proxy command between fmtcli and gRPC equivalent.
func setTemp(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "set-temperature")
	}
	client, cleanUp := getControllerClient(ctx) //This command returns the proto generated ControllerClient instance
	defer cleanUp()
	var (
		rateChange float64
	)
	setPoint := ctx.Args().First()
	setPt, err := strconv.ParseFloat(setPoint, 64);
	if err != nil {
		return err
	}
	if ctx.NumFlags() != 0 {
		rateChange = ctx.Float64("temp_change_rate")
	}
	stream, err := client.SetTemperature(ctxc, &demorpc.SetTempRequest{
		TempSetPoint: setPt,
		TempChangeRate: rateChange,
	})
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		printRespJSON(resp)
	}
	return nil
}

var setPresCommand = cli.Command{
	Name:  "set-pressure",
	Usage: "Set a pressure set point",
	ArgsUsage: "pres-set-point",
	Description: `
	Set a pressure set point (in Torr) to change the current pressure.`,
	Flags: []cli.Flag{
		cli.Float64Flag{
			Name:  "pres_change_rate",
			Usage: "The rate at which the pressure changes (Torr/minute). If not set, defaults to instant change.",
		},
	},
	Action: setPres,
}

// setPres is the proxy command between fmtcli and gRPC equivalent.
func setPres(ctx *cli.Context) error {
	ctxc := getContext()
	if ctx.NArg() != 1 {
		return cli.ShowCommandHelp(ctx, "set-pressure")
	}
	client, cleanUp := getControllerClient(ctx) //This command returns the proto generated ControllerClient instance
	defer cleanUp()
	var (
		rateChange float64
	)
	setPoint := ctx.Args().First()
	setPt, err := strconv.ParseFloat(setPoint, 64);
	if err != nil {
		return err
	}
	if ctx.NumFlags() != 0 {
		rateChange = ctx.Float64("pres_change_rate")
	}
	stream, err := client.SetPressure(ctxc, &demorpc.SetPresRequest{
		PressureSetPoint: setPt,
		PressureChangeRate: rateChange,
	})
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		printRespJSON(resp)
	}
	return nil
}