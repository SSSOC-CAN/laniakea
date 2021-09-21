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

// getClient returns the FmtClient instance from the fmtrpc package as well as a cleanup function
func getClient(ctx *cli.Context) (fmtrpc.FmtClient, func()) {
	conn := getClientConn(ctx)
	cleanUp := func() {
		conn.Close()
	}
	return fmtrpc.NewFmtClient(conn), cleanUp
}

// getClientConn returns the grpc Client connection for use in instantiating FmtClient
func getClientConn(ctx *cli.Context) *grpc.ClientConn {
	config, err := fmtd.InitConfig()
	if err != nil {
		fatal(fmt.Errorf("Could not load config: %v", err))
	}
	creds, err := credentials.NewClientTLSFromFile(config.TLSCertPath, "")
	if err != nil {
		fatal(err)
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	adminMac, err := os.ReadFile(config.AdminMacPath)
	if err != nil {
		fatal(fmt.Errorf("Could not read macaroon at %v: %v", config.AdminMacPath, err))
	}
	macHex := hex.EncodeToString(adminMac)
	mac, err := loadMacaroon(readPassword, macHex)
	if err != nil {
		fatal(fmt.Errorf("Could not load macaroon; %v", err))
	}
	macConstraints := []macaroons.Constraint{
		macaroons.TimeoutConstraint(defaultMacaroonTimeout),
	}
	constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
	if err != nil {
		fatal(err)
	}
	cred := macaroons.NewMacaroonCredential(constrainedMac)
	opts = append(opts, grpc.WithPerRPCCredentials(cred))
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
	}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}

func readPassword(text string) ([]byte, error) {
	fmt.Print(text)
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return pw, err
}