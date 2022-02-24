/*
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
package fmtd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/SSSOC-CAN/fmtd/testplan"
	"github.com/SSSOC-CAN/fmtd/unlocker"
	"github.com/SSSOC-CAN/fmtd/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
)

var (
	initialState = struct{rtd fmtrpc.RealTimeData}{
		rtd: drivers.FlukeInitialState,
	}
	tempPwd = []byte("abcdefgh")
)

// Main is the true entry point for fmtd. It's called in a nested manner for proper defer execution
func Main(interceptor *intercept.Interceptor, server *Server) error {
	var services []data.Service
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create State store
	rtdStateStore := state.CreateStore(drivers.FlukeInitialState, drivers.FlukeReducer)

	// Starting main server
	err := server.Start()
	if err != nil {
		server.logger.Fatal().Msg("Could not start server")
		return err
	}
	// server.logger.Debug().Msg(fmt.Sprintf("Server active: %v\tServer stopping: %v", server.Active, server.Stopping))
	defer server.Stop()

	// Get TLS config
	server.logger.Info().Msg("Loading TLS configuration...")
	serverOpts, restDialOpts, restListen, cleanUp, err := cert.GetTLSConfig(server.cfg.TLSCertPath, server.cfg.TLSKeyPath, server.cfg.ExtraIPAddr)
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Could not load TLS configuration: %v", err))
		return err
	}
	server.logger.Info().Msg("TLS configuration successfully loaded.")
	defer cleanUp()

	// Instantiating RPC server
	rpcServer, err := NewRpcServer(interceptor, server.cfg, server.logger)
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not initialize RPC server: %v", err))
		return err
	}
	server.logger.Info().Msg("RPC Server Initialized.")

	// Creating gRPC server and Server options
	grpc_interceptor := intercept.NewGrpcInterceptor(rpcServer.SubLogger, false)
	err = grpc_interceptor.AddPermissions(MainGrpcServerPermissions())
	rpcServerOpts := grpc_interceptor.CreateGrpcOptions()
	serverOpts = append(serverOpts, rpcServerOpts...)
	grpc_server := grpc.NewServer(serverOpts...)
	rpcServer.RegisterWithGrpcServer(grpc_server)

	// Instantiate RTD Service
	server.logger.Info().Msg("Instantiating RTD subservice...")
	rtdService := data.NewRTDService(&NewSubLogger(server.logger, "RTD").SubLogger, server.cfg.TCPAddr, server.cfg.TCPPort, rtdStateStore)
	err = rtdService.RegisterWithGrpcServer(grpc_server)
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to register RTD Service with gRPC server: %v", err))
		return err
	}
	services = append(services, rtdService)
	server.logger.Info().Msg("RTD service instantiated.")

	// Instantiate Fluke service and register with gRPC server but NOT start
	server.logger.Info().Msg("Instantiating RPC subservices and registering with gRPC server...")
	flukeService, err := drivers.NewFlukeService(&NewSubLogger(server.logger, "FLUKE").SubLogger, server.cfg.DataOutputDir, rtdStateStore)
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to instantiate Fluke service: %v", err))
		return err
	}
	flukeService.RegisterWithRTDService(rtdService)
	services = append(services, flukeService)

	// Instantiate RGA service and register with gRPC server and Fluke Service but NOT start
	rgaService, err := drivers.NewRGAService(&NewSubLogger(server.logger, "RGA").SubLogger, server.cfg.DataOutputDir, rtdStateStore)
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to instantiate RGA service: %v", err))
		return err
	}
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to register Fluke Service with gRPC server: %v", err))
		return err
	}
	rgaService.RegisterWithRTDService(rtdService)
	services = append(services, rgaService)

	// Instantiate Test Plan Executor and register with gRPC server but NOT start
	testPlanExecutor := testplan.NewTestPlanService(
		&NewSubLogger(server.logger, "TPEX").SubLogger,
		server.cfg.TLSCertPath,
		server.cfg.AdminMacPath,
		server.cfg.GrpcPort,
	)
	testPlanExecutor.RegisterWithGrpcServer(grpc_server)
	services = append(services, testPlanExecutor)

	server.logger.Info().Msg("RPC subservices instantiated and registered successfully.")

	// Starting kvdb
	server.logger.Info().Msg("Opening database...")
	db, err := kvdb.NewDB(server.cfg.MacaroonDBPath)
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not initialize Macaroon DB: %v", err))
		return err
	}
	server.logger.Info().Msg("Database successfully opened.")
	defer db.Close()

	// Instantiate Unlocker Service and register with gRPC server
	server.logger.Info().Msg("Initializing unlocker service...")
	unlockerService, err := unlocker.InitUnlockerService(db, []string{server.cfg.AdminMacPath, server.cfg.TestMacPath})
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not initialize unlocker service: %v", err))
		return err
	}
	unlockerService.RegisterWithGrpcServer(grpc_server)
	server.logger.Info().Msg("Unlocker service initialized.")

	// Starting RPC server
	err = rpcServer.Start()
	if err != nil {
		server.logger.Fatal().Msg(fmt.Sprintf("Could not start RPC server: %v", err))
		return err
	}
	defer rpcServer.Stop()
	
	// Start gRPC listening
	err = startGrpcListen(grpc_server, rpcServer.Listener)
	if err != nil {
		rpcServer.SubLogger.Fatal().Msg(fmt.Sprintf("Could not start gRPC listen on %v:%v", rpcServer.Listener.Addr(), err))
		return err
	}
	rpcServer.SubLogger.Info().Msg(fmt.Sprintf("gRPC listening on %v", rpcServer.Listener.Addr()))

	// Starting REST proxy
	stopProxy, err := startRestProxy(
		server.cfg, rpcServer, restDialOpts, restListen,
	)
	if err != nil {
		return err
	}
	defer stopProxy()

	// Wait for password
	grpc_interceptor.SetDaemonLocked()
	server.logger.Info().Msg("Waiting for password. Use `fmtcli setpassword` to set a password for the first time, " +
	"`fmtcli login` to unlock the daemon with an existing password, or `fmtcli changepassword` to change the " +
	"existing password and unlock the daemon.")
	pwd, err := waitForPassword(unlockerService, interceptor.ShutdownChannel())
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Error while awaiting password: %v", err))
	}
	grpc_interceptor.SetDaemonUnlocked()
	server.logger.Info().Msg("Login successful")

	// Instantiating Macaroon Service
	server.logger.Info().Msg("Initiating macaroon service...")
	macaroonService, err := macaroons.InitService(*db, "fmtd")
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to instantiate Macaroon service: %v", err))
		return err
	}
	server.logger.Info().Msg("Macaroon service initialized.")
	defer macaroonService.Close()

	// Unlock Macaroon Store
	server.logger.Info().Msg("Unlocking macaroon store...")
	err = macaroonService.CreateUnlock(&pwd)
	if err != nil {
		server.logger.Error().Msg(fmt.Sprintf("Unable to unlock macaroon store: %v", err))
		return err
	}
	server.logger.Info().Msg("Macaroon store unlocked.")
	// Baking Macaroons
	server.logger.Info().Msg("Baking macaroons...")
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
			return err
		}
	}
	server.logger.Info().Msg("Macaroons baked successfully.")
	grpc_interceptor.AddMacaroonService(macaroonService)
	rpcServer.AddMacaroonService(macaroonService)

	// Starting services TODO:SSSOCPaulCote - Start all subservices in go routines and make waitgroup
	for _, s := range services {
		err = s.Start()
		if err != nil {
			server.logger.Error().Msg(fmt.Sprintf("Unable to start %S service: %v", s.Name(), err))
			return err
		}
		defer s.Stop()
	}

	// Change RPC state to active
	grpc_interceptor.SetRPCActive()
	
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

// waitForPassword hangs until a password is provided or a shutdown request is receieved
func waitForPassword(u *unlocker.UnlockerService, shutdownChan <-chan struct{}) ([]byte, error) {
	select {
	case msg := <-u.PasswordMsgs:
		if msg.Err != nil {
			return nil, msg.Err
		}
		return msg.Password, nil
	case <-shutdownChan:
		return nil, fmt.Errorf("Shutting Down")
	}
}

// startGrpcListen starts the gRPC listening on given ports
func startGrpcListen(grpcServer *grpc.Server, listener net.Listener) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(lis net.Listener) {
		wg.Done()
		_ = grpcServer.Serve(lis)
	}(listener)
	wg.Wait()
	return nil
}

// startRestProxy starts the given REST proxy on the listeners found in the config.
func startRestProxy(cfg *Config, rpcServer *RpcServer, restDialOpts []grpc.DialOption, restListen func(net.Addr) (net.Listener, error)) (func(), error) {
	restProxyDestNet, err := utils.NormalizeAddresses([]string{fmt.Sprintf("localhost:%d", cfg.GrpcPort)}, strconv.FormatInt(cfg.GrpcPort, 10), net.ResolveTCPAddr)
	if err != nil {
		return nil, err
	}
	restProxyDest := restProxyDestNet[0].String()
	switch {
	case strings.Contains(restProxyDest, "0.0.0.0"):
		restProxyDest = strings.Replace(restProxyDest, "0.0.0.0", "127.0.0.1", 1)
	case strings.Contains(restProxyDest, "[::]"):
		restProxyDest = strings.Replace(restProxyDest, "[::]", "[::1]", 1)
	}
	var shutdownFuncs []func()
	shutdown := func() {
		for _, shutdownFn := range shutdownFuncs {
			shutdownFn()
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	shutdownFuncs = append(shutdownFuncs, cancel)

	customMarshalerOption := proxy.WithMarshalerOption(
		proxy.MIMEWildcard, &proxy.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
				EmitUnpopulated: true,
			},
		},
	)
	mux := proxy.NewServeMux(customMarshalerOption)

	err = fmtrpc.RegisterUnlockerHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return nil, err
	}
	err = rpcServer.RegisterWithRestProxy(
		ctx, mux, restDialOpts, restProxyDest,
	)
	if err != nil {
		return nil, err
	}
	// Wrap the default grpc-gateway handler with the WebSocket handler.
	restHandler := fmtrpc.NewWebSocketProxy(
		mux, rpcServer.SubLogger, cfg.WSPingInterval, cfg.WSPongWait,
	)
	var wg sync.WaitGroup
	restEndpoints, err := utils.NormalizeAddresses([]string{fmt.Sprintf("localhost:%d", cfg.RestPort)}, strconv.FormatInt(cfg.RestPort, 10), net.ResolveTCPAddr)
	if err != nil {
		rpcServer.SubLogger.Error().Msg(fmt.Sprintf("Unable to normalize address %s: %v", fmt.Sprintf("localhost:%d", cfg.RestPort), err))
	}
	restEndpoint := restEndpoints[0]
	lis, err := restListen(restEndpoint)
	if err != nil {
		rpcServer.SubLogger.Error().Msg(fmt.Sprintf("gRPC proxy unable to listen on %s: %v", restEndpoint, err))
	}
	shutdownFuncs = append(shutdownFuncs, func() {
		err := lis.Close()
		if err != nil {
			rpcServer.SubLogger.Error().Msg(fmt.Sprintf("Error closing listerner: %v", err))
		}
	})
	wg.Add(1)
	go func() {
		rpcServer.SubLogger.Info().Msg(fmt.Sprintf("gRPC proxy started and listening at %s", lis.Addr()))
		wg.Done()
		err := http.Serve(lis, restHandler)
		if err != nil && !fmtrpc.IsClosedConnError(err) {
			rpcServer.SubLogger.Error().Msg(fmt.Sprintf("%v", err))
		}
	}()
	wg.Wait()
	return shutdown, nil
}