/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20
*/

package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"

	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
)

var (
	defaultTestGrpcAddr = "localhost"
	defaultTestGrpcPort = "5678"
	defaultTestPassword = []byte("test")
)

// TestGetClientConnNoMac tests whether we can produce a gRPC client without macaroons
func TestGetClientConnNoMac(t *testing.T) {
	// let's make a temporary directory for a test TLS cert/key pair
	tempDir, err := ioutil.TempDir("", "tlsstuff-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempTLSCertPath := path.Join(tempDir, "tls.cert")
	tempTLSKeyPath := path.Join(tempDir, "tls.key")
	// generate TLS config for our temp gRPC server
	serverOpts, _, _, cleanUp, err := cert.GetTLSConfig(
		tempTLSCertPath,
		tempTLSKeyPath,
		make([]string, 0),
	)
	if err != nil {
		t.Fatalf("Error creating temporary TLS configuration: %v", err)
	}
	defer cleanUp()
	// grpcServer and net.Listener
	grpcServer := grpc.NewServer(serverOpts...)
	defer grpcServer.Stop()
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", defaultTestGrpcAddr, defaultTestGrpcPort))
	if err != nil {
		t.Fatalf("Failed to listen at %v:%v: %v", defaultTestGrpcAddr, defaultTestGrpcPort, err)
	}
	// goroutine for gRPC serve
	go func(listener net.Listener) {
		_ = grpcServer.Serve(listener)
	}(lis)
	// Now we get gRPC Client
	grpcClient, err := GetClientConn(
		defaultTestGrpcAddr,
		defaultTestGrpcPort,
		tempTLSCertPath,
		path.Join(tempDir, "admin.macaroon"),
		true,
		int64(0),
	)
	defer grpcClient.Close()
	if err != nil {
		t.Fatalf("Unable to connect to server at %v:%v: %v", defaultTestGrpcAddr, defaultTestGrpcPort, err)
	}
}

// TestGetClientConnWithMac tests whether we can produce a gRPC client with macaroons
func TestGetClientConnWithMac(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// let's make a temporary directory for a test TLS cert/key pair
	tempDir, err := ioutil.TempDir("", "tlsstuff-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempTLSCertPath := path.Join(tempDir, "tls.cert")
	tempTLSKeyPath := path.Join(tempDir, "tls.key")
	// generate TLS config for our temp gRPC server
	serverOpts, _, _, cleanUp, err := cert.GetTLSConfig(
		tempTLSCertPath,
		tempTLSKeyPath,
		make([]string, 0),
	)
	if err != nil {
		t.Fatalf("Error creating temporary TLS configuration: %v", err)
	}
	defer cleanUp()
	// macaroons
	tempAdminMacPath := path.Join(tempDir, "admin.macaroon")
	tempMacDBPath := path.Join(tempDir, "macaroon.db")
	// new kvdb
	db, err := kvdb.NewDB(tempMacDBPath)
	if err != nil {
		t.Fatalf("Error creating macaroon db: %v", err)
	}
	defer db.Close()
	// macaroon service
	macaroonService, err := macaroons.InitService(*db, "laniakea", zerolog.New(os.Stderr).With().Timestamp().Logger(), []string{})
	if err != nil {
		t.Fatalf("Error creating macaroon service: %v", err)
	}
	defer macaroonService.Close()
	// unlock macaroon store
	err = macaroonService.CreateUnlock(&defaultTestPassword)
	if err != nil {
		t.Fatalf("Error unlocking macaroon store: %v", err)
	}
	// bake macaroon
	mac, err := macaroonService.NewMacaroon(
		ctx,
		macaroons.DefaultRootKeyID,
		[]checkers.Caveat{macaroons.TimeoutCaveat(int64(0))},
		[]bakery.Op{
			{
				Entity: "laniakea",
				Action: "read",
			},
		}...,
	)
	if err != nil {
		t.Fatalf("Could not bake new macaroon: %v", err)
	}
	macBytes, err := mac.M().MarshalBinary()
	if err != nil {
		t.Fatalf("Could not bake new macaroon: %v", err)
	}
	err = ioutil.WriteFile(tempAdminMacPath, macBytes, 0755)
	if err != nil {
		_ = os.Remove(tempAdminMacPath)
		t.Fatalf("Could not bake new macaroon: %v", err)
	}
	// grpcServer and net.Listener
	grpcServer := grpc.NewServer(serverOpts...)
	defer grpcServer.Stop()
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", defaultTestGrpcAddr, defaultTestGrpcPort))
	if err != nil {
		t.Fatalf("Failed to listen at %v:%v: %v", defaultTestGrpcAddr, defaultTestGrpcPort, err)
	}
	// goroutine for gRPC serve
	go func(listener net.Listener) {
		_ = grpcServer.Serve(listener)
	}(lis)
	// Now we get gRPC Client
	grpcClient, err := GetClientConn(
		defaultTestGrpcAddr,
		defaultTestGrpcPort,
		tempTLSCertPath,
		tempAdminMacPath,
		false,
		int64(30),
	)
	defer grpcClient.Close()
	if err != nil {
		t.Fatalf("Unable to connect to server at %v:%v: %v", defaultTestGrpcAddr, defaultTestGrpcPort, err)
	}
}
