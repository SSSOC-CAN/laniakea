package fmtd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/utils"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
)

var (
	readPermissions = []bakery.Op{
		{
			Entity: "fmtd",
			Action: "read",
		},
		{
			Entity: "macaroon",
			Action: "read",
		},
	}
	writePermissions = []bakery.Op{
		{
			Entity: "fmtd",
			Action: "write",
		},
		{
			Entity: "macaroon",
			Action: "generate",
		},
		{
			Entity: "macaroon",
			Action: "write",
		},
	}
)

// Main is the true entry point for fmtd. It's called in a nested manner for proper defer execution
func Main(interceptor *intercept.Interceptor, server *Server) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Starting main server
	err := server.Start()
	if err != nil {
		server.logger.Fatal().Msg("Could not start server")
		return err
	}
	server.logger.Debug().Msg(fmt.Sprintf("Server active: %v\tServer stopping: %v", server.Active, server.Stopping))
	defer server.Stop()

	// Get TLS config
	server.logger.Info().Msg("Loading TLS configuration...")
	serverOpts, _, _, cleanUp, err := cert.GetTLSConfig(server.cfg.TLSCertPath, server.cfg.TLSKeyPath)
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Could not load TLS configuration: %v", err))
		return err
	}
	server.logger.Info().Msg("TLS configuration successfully loaded.")
	defer cleanUp()

	// Starting RPC server
	rpcServer, err := NewRpcServer(interceptor, server.cfg, server.logger)
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not initialize RPC server: %v", err))
		return err
	}
	server.logger.Info().Msg("RPC Server Initialized")

	// Creating gRPC server and Server options
	grpc_interceptor := intercept.NewGrpcInterceptor(rpcServer.SubLogger, false)
	err = grpc_interceptor.AddPermissions(intercept.MainGrpcServerPermissions())
	rpcServerOpts := grpc_interceptor.CreateGrpcOptions()
	serverOpts = append(serverOpts, rpcServerOpts...)
	grpc_server := grpc.NewServer(serverOpts...)
	rpcServer.AddGrpcServer(grpc_server)
	defer rpcServer.GrpcServer.Stop()

	// Starting bbolt kvdb
	db, err := bolt.Open(server.cfg.MacaroonDBPath, 0755, nil)
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not initialize Macaroon DB: %v", err))
		return err
	}
	defer db.Close()

	// Instantiating Macaroon Service
	macaroonService, err := macaroons.InitService(*db, "fmtd")
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to instantiate Macaroon service: %v", err))
		return err
	}
	defer macaroonService.Close()

	// Baking Macaroons
	if !utils.FileExists(server.cfg.AdminMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.AdminMacPath, adminPermissions(), false, 0,
		)
		if err != nil {
			server.logger.Error().Msg(fmt.Sprintf("Unable to create admin macaroon: %v", err))
			return err
		}
	}
	if !utils.FileExists(server.cfg.TestMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.TestMacPath, readPermissions, true, 120,
		)
		if err != nil {
			server.logger.Error().Msg(fmt.Sprintf("Unable to create test macaroon: %v", err))
		}
	}
	grpc_interceptor.AddMacaroonService(macaroonService)

	//Starting RPC and gRPC Servers
	err = rpcServer.Start()
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not start RPC server: %v", err))
		return err
	}
	server.logger.Info().Msg("RPC Server Started")
	defer rpcServer.Stop()

	<-interceptor.ShutdownChannel()
	return nil
}

// bakeMacaroons is a wrapper function around the NewMacaroon method of the macaroons.Service struct
func bakeMacaroons(ctx context.Context, svc *macaroons.Service, perms []bakery.Op, noTimeOutCaveat bool, seconds int64) ([]byte, error) {
	mac, err := svc.NewMacaroon(
		ctx,
		macaroons.DefaultRootKeyID,
		noTimeOutCaveat,
		[]checkers.Caveat{macaroons.TimeoutCaveat(seconds)},
		perms...,
	)
	if err != nil {
		return nil, err
	}
	return mac.M().MarshalBinary()
}

// genMacaroons will create the macaroon files specified if not already created
func genMacaroons(ctx context.Context, svc *macaroons.Service, macFile string, perms []bakery.Op, noTimeOutCaveat bool, seconds int64) error {
	macBytes, err := bakeMacaroons(ctx, svc, perms, noTimeOutCaveat, seconds)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(macFile, macBytes, 0755)
	if err != nil {
		_ = os.Remove(macFile)
		return err
	}
	return nil
}

// adminPermissions returns the permissions associated with the admin macaroon
func adminPermissions() []bakery.Op {
	admin := make([]bakery.Op, len(readPermissions)+len(writePermissions))
	copy(admin[:len(readPermissions)], readPermissions)
	copy(admin[len(readPermissions):], writePermissions)
	return admin
}