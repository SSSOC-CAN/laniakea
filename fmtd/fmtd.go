package fmtd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	readPermissions = []bakery.Op{
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
	if !cert.FileExists(server.cfg.AdminMacPath) {
		err = genMacaroons(
			ctx, macaroonService, server.cfg.AdminMacPath,
		)
		if err != nil {
			server.logger.Error().Msg(fmt.Sprintf("Unable to create macaroons: %v", err))
			return err
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

func bakeMacaroons(ctx context.Context, svc *macaroons.Service, permissions []bakery.Op) ([]byte, error) {
	mac, err := svc.NewMacaroon(ctx, macaroons.DefaultRootKeyID, permissions...)
	if err != nil {
		return nil, err
	}
	return mac.M().MarshalBinary()
}

func genMacaroons(ctx context.Context, svc *macaroons.Service, adminFile string) error {
	adminBytes, err := bakeMacaroons(ctx, svc, adminPermissions())
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(adminFile, adminBytes, 0755)
	if err != nil {
		_ = os.Remove(adminFile)
		return err
	}
	return nil
}

func adminPermissions() []bakery.Op {
	admin := make([]bakery.Op, len(readPermissions)+len(writePermissions))
	copy(admin[:len(readPermissions)], readPermissions)
	copy(admin[len(readPermissions):], writePermissions)
	return admin
}