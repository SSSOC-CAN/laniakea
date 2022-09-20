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
	"bytes"
	"fmt"

	"github.com/SSSOC-CAN/fmtd/auth"
	"github.com/SSSOC-CAN/fmtd/lanirpc"
	"github.com/urfave/cli"
)

var setPassword = cli.Command{
	Name:  "setpassword",
	Usage: "Sets a password when starting weirwood for the first time.",
	Description: `
	Sets a password used to unlock the macaroon key-store when starting weirwood for the first time.
	There is one required argument for this command, the password.`,
	Action: setPwd,
}

// setPwd is the proxy command between lanicli and gRPC equivalent.
func setPwd(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getUnlockerClient(ctx) //This command returns the proto generated UnlockerClient instance
	defer cleanUp()
	pwd, err := capturePassword(
		"Input new password: ",
	)
	if err != nil {
		return err
	}
	_, err = client.SetPassword(ctxc, &lanirpc.SetPwdRequest{
		Password: pwd,
	})
	if err != nil {
		return err
	}
	fmt.Println("\nDaemon unlocked successfully!")
	return nil
}

var loginCommand = cli.Command{
	Name:  "login",
	Usage: "Login into Laniakea",
	Description: `
	When invoked, user will be prompted to enter a password. Either create one or use an existing one.`,
	Action: login,
}

// login is the wrapper around the UnlockerClient.Login method
func login(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getUnlockerClient(ctx)
	defer cleanUp()
	pwd, err := auth.ReadPassword("Input password: ")
	if err != nil {
		return err
	}
	loginReq := &lanirpc.LoginRequest{
		Password: pwd,
	}
	_, err = client.Login(ctxc, loginReq)
	if err != nil {
		return err
	}
	fmt.Println("\nDaemon unlocked successfully!")
	return nil
}

var changePassword = cli.Command{
	Name:  "changepassword",
	Usage: "Changes the previously set password.",
	Description: `
	Changes the previously set password used to unlock the macaroon key-store.
	There are two required arguments for this command, the old password and the new password.`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "new_mac_root_key",
			Usage: "rotate the macaroon root key resulting in " +
				"all previously created macaroons to be " +
				"invalidated",
		},
	},
	Action: changePwd,
}

// changePwd is the proxy command between lanicli and gRPC equivalent.
func changePwd(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getUnlockerClient(ctx) //This command returns the proto generated UnlockerClient instance
	defer cleanUp()
	oldPwd, err := auth.ReadPassword(
		"Input old password: ",
	)
	if err != nil {
		return err
	}
	newPwd, err := capturePassword(
		"Input new password: ",
	)
	_, err = client.ChangePassword(ctxc, &lanirpc.ChangePwdRequest{
		CurrentPassword:    oldPwd,
		NewPassword:        newPwd,
		NewMacaroonRootKey: ctx.Bool("new_mac_root_key"),
	})
	if err != nil {
		return err
	}
	fmt.Println("\nDaemon unlocked successfully!")
	return nil
}

// capturePassword captures the password from the terminal and a confirmation of the password
func capturePassword(instruction string) ([]byte, error) {
	for {
		password, err := auth.ReadPassword(instruction)
		if err != nil {
			return nil, err
		}
		pwdConfirmed, err := auth.ReadPassword("Confirm new password: ")
		if bytes.Equal(password, pwdConfirmed) {
			return password, nil
		}
		fmt.Println("Passwords don't match, please try again")
		fmt.Println()
	}
}
