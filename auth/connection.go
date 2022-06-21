/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package auth

import (
	"encoding/hex"
	e "github.com/pkg/errors"
	"fmt"
	"os"
	"syscall"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/utils"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)
)

// GetClientConn returns the grpc Client connection for use in instantiating gRPC Clients
func GetClientConn(grpcServerAddr, grpcServerPort, tlsCertPath, adminMacPath string, skipMacaroons bool, macaroon_timeout int64) (*grpc.ClientConn, error) {
	//get TLS credentials from TLS certificate file
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		return nil, e.Wrap(err, "could not get TLS credentials from file")
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	if !skipMacaroons {
		// grab Macaroon data and load it into macaroon.Macaroon struct
		adminMac, err := os.ReadFile(adminMacPath)
		if err != nil {
			return nil, e.Wrapf(err, "Could not read macaroon at %v", adminMacPath)
		}
		macHex := hex.EncodeToString(adminMac)
		mac, err := LoadMacaroon(ReadPassword, macHex)
		if err != nil {
			return nil, err
		}
		// Add constraints to our macaroon
		macConstraints := []macaroons.Constraint{
			macaroons.TimeoutConstraint(macaroon_timeout), // prevent a replay attack
		}
		constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
		if err != nil {
			return nil, err
		}
		cred, err := macaroons.NewMacaroonCredential(constrainedMac)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	genericDialer := utils.ClientAddressDialer(grpcServerPort)
	opts = append(opts, grpc.WithContextDialer(genericDialer))
	opts = append(opts, grpc.WithDefaultCallOptions(maxMsgRecvSize))
	conn, err := grpc.Dial(grpcServerAddr+":"+grpcServerPort, opts...)
	if err != nil {
		return nil, e.Wrap(err, "unable to dial gRPC server")
	}
	return conn, nil
}


// ReadPassword prompts the user for a password in the command line
func ReadPassword(text string) ([]byte, error) {
	fmt.Print(text)
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return pw, err
}