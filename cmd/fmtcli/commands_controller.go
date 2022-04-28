// +build !demo

package main

import (
	"github.com/urfave/cli"
)

var setTempCommand = cli.Command{
	Name:  "set-temperature",
	Usage: "Set a temperature set point",
	ArgsUsage: "temp-set-point",
	Description: `
	Set a temperature set point (in degrees Celsius) to change the current average temperature. This feature is not currently active.`,
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
	return cli.ShowCommandHelp(ctx, "set-temperature")
}

var setPresCommand = cli.Command{
	Name:  "set-pressure",
	Usage: "Set a pressure set point",
	ArgsUsage: "pres-set-point",
	Description: `
	Set a pressure set point (in Torr) to change the current pressure. This feature is not currently active.`,
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
	return cli.ShowCommandHelp(ctx, "set-pressure")
}