package auth

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"github.com/SSSOC-CAN/fmtd/fmtd"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)
)

// GetClientConn returns the grpc Client connection for use in instantiating FmtClient
func GetClientConn(skipMacaroons bool, macaroon_timeout int64) (*grpc.ClientConn, error) {
	config, err := fmtd.InitConfig()
	if err != nil {
		return nil, fmt.Errorf("Could not load config: %v", err)
	}
	//get TLS credentials from TLS certificate file
	creds, err := credentials.NewClientTLSFromFile(config.TLSCertPath, "")
	if err != nil {
		return nil, err
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	if !skipMacaroons {
		// grab Macaroon data and load it into macaroon.Macaroon struct
		adminMac, err := os.ReadFile(config.AdminMacPath)
		if err != nil {
			return nil, fmt.Errorf("Could not read macaroon at %v: %v", config.AdminMacPath, err)
		}
		macHex := hex.EncodeToString(adminMac)
		mac, err := loadMacaroon(readPassword, macHex)
		if err != nil {
			return nil, fmt.Errorf("Could not load macaroon; %v", err)
		}
		// Add constraints to our macaroon
		macConstraints := []macaroons.Constraint{
			macaroons.TimeoutConstraint(macaroon_timeout), // prevent a replay attack
		}
		constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
		if err != nil {
			return nil, err
		}
		cred := macaroons.NewMacaroonCredential(constrainedMac)
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	genericDialer := utils.ClientAddressDialer(strconv.FormatInt(config.GrpcPort, 10))
	opts = append(opts, grpc.WithContextDialer(genericDialer))
	opts = append(opts, grpc.WithDefaultCallOptions(maxMsgRecvSize))
	conn, err := grpc.Dial("localhost:"+strconv.FormatInt(config.GrpcPort, 10), opts...)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to RPC server: %v", err)
	}
	return conn, nil
}