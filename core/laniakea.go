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
package core

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

	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/cert"
	"github.com/SSSOC-CAN/fmtd/health"
	"github.com/SSSOC-CAN/fmtd/intercept"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/lanirpc"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	"github.com/SSSOC-CAN/fmtd/plugins"
	"github.com/SSSOC-CAN/fmtd/unlocker"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
)

const (
	ErrShutdown = bg.Error("shutting down")
)

var (
	tempPwd                 = []byte("abcdefgh")
	defaultMacTimeout int64 = 60
)

// Main is the true entry point for Laniakea. It's called in a nested manner for proper defer execution
func Main(interceptor *intercept.Interceptor, server *Server) error {
	var services []api.Service
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	registeredPlugins := []string{}
	for _, cfg := range server.cfg.Plugins {
		registeredPlugins = append(registeredPlugins, cfg.Name)
	}

	// Starting main server
	err := server.Start()
	if err != nil {
		server.logger.Fatal().Msg("Could not start server")
		return err
	}
	defer server.Stop()

	// Get TLS config
	server.logger.Info().Msg("Loading TLS configuration...")
	serverOpts, restDialOpts, restListen, cleanUp, err := cert.GetTLSConfig(server.cfg.TLSCertPath, server.cfg.TLSKeyPath, server.cfg.ExtraIPAddr)
	if err != nil {
		server.logger.Error().Msgf("Could not load TLS configuration: %v", err)
		return err
	}
	server.logger.Info().Msg("TLS configuration successfully loaded.")
	defer cleanUp()

	// Instantiating RPC server
	rpcServer, err := NewRpcServer(interceptor, server.cfg, server.logger)
	if err != nil {
		server.logger.Fatal().Msgf("Could not initialize RPC server: %v", err)
		return err
	}
	server.logger.Info().Msg("RPC Server Initialized.")

	// Creating gRPC server and Server options
	grpc_interceptor := intercept.NewGrpcInterceptor(rpcServer.SubLogger, false)
	err = grpc_interceptor.AddPermissions(MainGrpcServerPermissions())
	if err != nil {
		server.logger.Error().Msgf("Could not add permissions to gRPC middleware: %v", err)
		return err
	}
	rpcServerOpts := grpc_interceptor.CreateGrpcOptions()
	serverOpts = append(serverOpts, rpcServerOpts...)
	grpc_server := grpc.NewServer(serverOpts...)
	defer grpc_server.Stop()
	rpcServer.RegisterWithGrpcServer(grpc_server)
	rpcServer.AddGrpcInterceptor(grpc_interceptor)

	// Initialize Health Checker
	healthChecker := health.NewHealthService()
	healthChecker.RegisterWithGrpcServer(grpc_server)
	err = healthChecker.RegisterHealthService("main", rpcServer)
	if err != nil {
		server.logger.Error().Msgf("could not register RPC server with health check server: %v", err)
		return err
	}

	// Initialize Plugin Manager
	server.logger.Info().Msg("Initializing plugins...")
	pluginManager := plugins.NewPluginManager(server.cfg.PluginDir, server.cfg.Plugins, NewSubLogger(server.logger, "PLGN").SubLogger, false)
	pluginManager.RegisterWithGrpcServer(grpc_server)
	err = pluginManager.AddPermissions(StreamingPluginAPIPermission())
	if err != nil {
		server.logger.Error().Msgf("Could not add permissions to plugin manager: %v", err)
		return err
	}
	err = pluginManager.Start(ctx)
	if err != nil {
		server.logger.Error().Msgf("Unable to start plugin manager: %v", err)
		return err
	}
	defer pluginManager.Stop()
	server.logger.Info().Msg("Plugins initialized")
	err = healthChecker.RegisterHealthService("plugins", pluginManager)
	if err != nil {
		server.logger.Error().Msgf("could not register plugin manager with health check server: %v", err)
		return err
	}
	server.logger.Info().Msg("RPC subservices instantiated and registered successfully.")

	// Starting kvdb
	server.logger.Info().Msg("Opening database...")
	db, err := kvdb.NewDB(server.cfg.MacaroonDBPath)
	if err != nil {
		server.logger.Fatal().Msgf("Could not initialize Macaroon DB: %v", err)
		return err
	}
	server.logger.Info().Msg("Database successfully opened.")
	defer db.Close()

	// Instantiate Unlocker Service and register with gRPC server
	server.logger.Info().Msg("Initializing unlocker service...")
	unlockerService, err := unlocker.InitUnlockerService(db, []string{server.cfg.AdminMacPath, server.cfg.TestMacPath})
	if err != nil {
		server.logger.Fatal().Msgf("Could not initialize unlocker service: %v", err)
		return err
	}
	unlockerService.RegisterWithGrpcServer(grpc_server)
	server.logger.Info().Msg("Unlocker service initialized.")

	// Starting RPC server
	err = rpcServer.Start()
	if err != nil {
		server.logger.Fatal().Msgf("Could not start RPC server: %v", err)
		return err
	}
	defer rpcServer.Stop()

	// Start gRPC listening
	err = startGrpcListen(grpc_server, rpcServer.Listener)
	if err != nil {
		rpcServer.SubLogger.Fatal().Msgf("Could not start gRPC listen on %v:%v", rpcServer.Listener.Addr(), err)
		return err
	}
	rpcServer.SubLogger.Info().Msgf("gRPC listening on %v", rpcServer.Listener.Addr())

	// Starting REST proxy
	stopProxy, err := startRestProxy(
		server.cfg,
		rpcServer,
		[]api.RestProxyService{
			rpcServer,
			unlockerService,
			pluginManager,
			healthChecker,
		},
		restDialOpts,
		restListen,
	)
	if err != nil {
		return err
	}
	defer stopProxy()

	// Wait for password
	grpc_interceptor.SetDaemonLocked()
	server.logger.Info().Msg("Waiting for password. Use `lanicli setpassword` to set a password for the first time, " +
		"`lanicli login` to unlock the daemon with an existing password, or `lanicli changepassword` to change the " +
		"existing password and unlock the daemon.")
	pwd, err := waitForPassword(unlockerService, interceptor.ShutdownChannel())
	if err != nil {
		server.logger.Error().Msgf("Error while awaiting password: %v", err)
		return err
	}
	grpc_interceptor.SetDaemonUnlocked()
	server.logger.Info().Msg("Login successful")

	// Instantiating Macaroon Service
	server.logger.Info().Msg("Initiating macaroon service...")
	macaroonService, err := macaroons.InitService(*db, "laniakea", NewSubLogger(server.logger, "BAKE").SubLogger, registeredPlugins)
	if err != nil {
		server.logger.Error().Msgf("Unable to instantiate Macaroon service: %v", err)
		return err
	}
	server.logger.Info().Msg("Macaroon service initialized.")
	defer macaroonService.Close()

	// Unlock Macaroon Store
	server.logger.Info().Msg("Unlocking macaroon store...")
	err = macaroonService.CreateUnlock(&pwd)
	if err != nil {
		server.logger.Error().Msgf("Unable to unlock macaroon store: %v", err)
		return err
	}
	server.logger.Info().Msg("Macaroon store unlocked.")
	// Baking Macaroons
	server.logger.Info().Msg("Baking macaroons...")
	// Delete macaroon files if user requested it
	if server.cfg.RegenerateMacaroons {
		if utils.FileExists(server.cfg.AdminMacPath) {
			err := os.Remove(server.cfg.AdminMacPath)
			if err != nil {
				server.logger.Error().Msgf("Unexpected error when deleting %s: %v", server.cfg.AdminMacPath, err)
			}
		}
		if utils.FileExists(server.cfg.TestMacPath) {
			err := os.Remove(server.cfg.TestMacPath)
			if err != nil {
				server.logger.Error().Msgf("Unexpected error when deleting %s: %v", server.cfg.TestMacPath, err)
			}
		}
	}
	if !utils.FileExists(server.cfg.AdminMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.AdminMacPath, adminPermissions(), 0, []string{"all"},
		)
		if err != nil {
			server.logger.Error().Msgf("Unable to create admin macaroon: %v", err)
			return err
		}
	}
	if !utils.FileExists(server.cfg.TestMacPath) {
		err := genMacaroons(
			ctx, macaroonService, server.cfg.TestMacPath, readPermissions, 120, registeredPlugins,
		)
		if err != nil {
			server.logger.Error().Msgf("Unable to create test macaroon: %v", err)
			return err
		}
	}
	server.logger.Info().Msg("Macaroons baked successfully.")
	grpc_interceptor.AddMacaroonService(macaroonService)
	rpcServer.AddMacaroonService(macaroonService)
	pluginManager.AddMacaroonService(macaroonService)

	// Starting services TODO:SSSOCPaulCote - Start all subservices in go routines and make waitgroup
	for _, s := range services {
		err = s.Start()
		if err != nil {
			server.logger.Error().Msgf("Unable to start %s service: %v", s.Name(), err)
			return err
		}
		defer s.Stop()
	}

	cleanUpServices := func() {
		for i := len(services) - 1; i > -1; i-- {
			err := services[i].Stop()
			if err != nil {
				server.logger.Error().Msgf("Unable to stop %s service: %v", services[i].Name(), err)
			}
		}
	}
	defer cleanUpServices()

	// Change RPC state to active
	grpc_interceptor.SetRPCActive()
	server.logger.Info().Msg("Laniakea started successfully and ready to use")
	<-interceptor.ShutdownChannel()
	return nil
}

// bakeMacaroons is a wrapper function around the NewMacaroon method of the macaroons.Service struct
func bakeMacaroons(ctx context.Context, svc *macaroons.Service, perms []bakery.Op, seconds int64, pluginNames []string) ([]byte, error) {
	caveats := []checkers.Caveat{}
	if seconds != 0 {
		caveats = append(caveats, macaroons.TimeoutCaveat(seconds))
	}
	if len(pluginNames) != 0 {
		caveats = append(caveats, macaroons.PluginCaveat(pluginNames))
	}

	mac, err := svc.NewMacaroon(
		ctx,
		macaroons.DefaultRootKeyID,
		caveats,
		perms...,
	)
	if err != nil {
		return nil, err
	}
	return mac.M().MarshalBinary()
}

// genMacaroons will create the macaroon files specified if not already created
func genMacaroons(ctx context.Context, svc *macaroons.Service, macFile string, perms []bakery.Op, seconds int64, pluginNames []string) error {
	macBytes, err := bakeMacaroons(ctx, svc, perms, seconds, pluginNames)
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
		return nil, ErrShutdown
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
func startRestProxy(cfg *Config, rpcServer *RpcServer, services []api.RestProxyService, restDialOpts []grpc.DialOption, restListen func(net.Addr) (net.Listener, error)) (func(), error) {
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
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
		},
	)
	mux := proxy.NewServeMux(customMarshalerOption, proxy.WithDisablePathLengthFallback())

	err = lanirpc.RegisterUnlockerHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return nil, err
	}
	for _, s := range services {
		err = s.RegisterWithRestProxy(
			ctx, mux, restDialOpts, restProxyDest,
		)
		if err != nil {
			return nil, err
		}
	}
	// Wrap the default grpc-gateway handler with the WebSocket handler.
	restHandler := lanirpc.NewWebSocketProxy(
		mux, rpcServer.SubLogger, cfg.WSPingInterval, cfg.WSPongWait,
	)
	var wg sync.WaitGroup
	restEndpoints, err := utils.NormalizeAddresses([]string{fmt.Sprintf("0.0.0.0:%d", cfg.RestPort)}, strconv.FormatInt(cfg.RestPort, 10), net.ResolveTCPAddr)
	if err != nil {
		rpcServer.SubLogger.Error().Msgf("Unable to normalize address %s: %v", fmt.Sprintf("0.0.0.0:%d", cfg.RestPort), err)
	}
	restEndpoint := restEndpoints[0]
	lis, err := restListen(restEndpoint)
	if err != nil {
		rpcServer.SubLogger.Error().Msgf("REST proxy unable to listen on %s: %v", restEndpoint, err)
	}
	shutdownFuncs = append(shutdownFuncs, func() {
		err := lis.Close()
		if err != nil {
			rpcServer.SubLogger.Error().Msgf("Error closing listerner: %v", err)
		}
	})
	wg.Add(1)
	go func() {
		rpcServer.SubLogger.Info().Msgf("REST proxy started and listening at %s", lis.Addr())
		wg.Done()
		err := http.Serve(lis, restHandler)
		if err != nil && !lanirpc.IsClosedConnError(err) {
			rpcServer.SubLogger.Error().Msg(err.Error())
		}
	}()
	wg.Wait()
	return shutdown, nil
}
