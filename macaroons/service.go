/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10

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
package macaroons

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/utils"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/metadata"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
	macaroon "gopkg.in/macaroon.v2"
)

const (
	PermissionEntityCustomURI           = "uri"
	ErrMissingRootKeyID                 = bg.Error("missing root key ID")
	ErrValidatorNil                     = bg.Error("validator cannot be nil")
	ErrValidatorMethodAlreadyRegistered = bg.Error("external validator for method already registered")
	ErrMetadataFromContext              = bg.Error("unable to get metadata from context")
	ErrUnexpectedMacNumber              = bg.Error("unexpected number of macaroons")
	ErrKeyNotInContext                  = bg.Error("key is not in the context")
	PluginContextKey                    = "plugin"
)

type MacaroonValidator interface {
	ValidateMacaroon(ctx context.Context, requiredPermissions []bakery.Op, fullMethod string) error
}

type Service struct {
	bakery.Bakery

	rks                *RootKeyStorage
	ExternalValidators map[string]MacaroonValidator
	registeredPlugins  []string
	pluginMethods      []string
}

// InitService returns initializes the rootkeystorage for the Macaroon service and returns the initialized service
func InitService(db kvdb.DB, location string, logger zerolog.Logger, pluginNames, pluginMethods []string, checks ...Checker) (*Service, error) {
	rks, err := InitRootKeyStorage(db)
	if err != nil {
		return nil, err
	}
	bakeryParams := bakery.BakeryParams{
		Location:     location,
		RootKeyStore: rks,
		Locator:      nil,
		Key:          nil,
		Logger:       &MacLogger{Logger: logger},
	}
	service := bakery.New(bakeryParams)
	checker := service.Checker.FirstPartyCaveatChecker.(*checkers.Checker)
	for _, check := range checks {
		cond, fun := check()
		if !isRegistered(checker, cond) {
			checker.Register(cond, "std", fun)
		}
	}
	return &Service{
		Bakery:             *service,
		rks:                rks,
		ExternalValidators: make(map[string]MacaroonValidator),
		registeredPlugins:  pluginNames,
		pluginMethods:      pluginMethods,
	}, nil
}

// isRegistered checks to see if the required checker has already been registered to avoid duplicates
func isRegistered(c *checkers.Checker, name string) bool {
	if c == nil {
		return false
	}
	for _, info := range c.Info() {
		if info.Name == name && info.Prefix == "" && info.Namespace == "std" {
			return true
		}
	}
	return false
}

// RegisterExternalValidator registers a custom, external macaroon validator for
// the specified absolute gRPC URI. That validator is then fully responsible to
// make sure any macaroon passed for a request to that URI is valid and
// satisfies all conditions.
func (svc *Service) RegisterExternalValidator(fullMethod string,
	validator MacaroonValidator) error {

	if validator == nil {
		return ErrValidatorNil
	}

	_, ok := svc.ExternalValidators[fullMethod]
	if ok {
		return ErrValidatorMethodAlreadyRegistered
	}

	svc.ExternalValidators[fullMethod] = validator
	return nil
}

// ValidateMacaroon validates the capabilities of a given request given a
// bakery service, context, and uri. Within the passed context.Context, we
// expect a macaroon to be encoded as request metadata using the key
// "macaroon".
func (svc *Service) ValidateMacaroon(ctx context.Context,
	requiredPermissions []bakery.Op, fullMethod string) error {

	// Get macaroon bytes from context and unmarshal into macaroon.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ErrMetadataFromContext
	}
	if len(md["macaroon"]) != 1 {
		return ErrUnexpectedMacNumber
	}

	// With the macaroon obtained, we'll now decode the hex-string
	// encoding, then unmarshal it from binary into its concrete struct
	// representation.
	macBytes, err := hex.DecodeString(md["macaroon"][0])
	if err != nil {
		return err
	}
	mac := &macaroon.Macaroon{}
	err = mac.UnmarshalBinary(macBytes)
	if err != nil {
		return err
	}

	// If the full method is a plugin method, then do the following
	if len(svc.pluginMethods) != 0 && utils.StrInStrSlice(svc.pluginMethods, fullMethod) {
		reqPlugName, ok := ctx.Value(PluginContextKey).(string)
		if !ok || reqPlugName == "" {
			return ErrKeyNotInContext
		}
		if reqPlugName != "all" {
			// Check to see if the plugins in the macaroon match the list of registered plugins
			if len(svc.registeredPlugins) != 0 {
				var okP bool
				for _, caveat := range mac.Caveats() {
					split := strings.Split(string(caveat.Id), " ") // assuming "declared plugins plugin1:plugin-2"
					if len(split) >= 2 && split[1] == "plugins" {
						if strings.Contains(split[2], ":") {
							plugins := strings.Split(split[2], ":")
							if utils.StrInStrSlice(plugins, reqPlugName) {
								okP = true
							}
						} else {
							if strings.Contains(reqPlugName, split[2]) {
								okP = true
							}
						}
					}
				}
				if !okP {
					return ErrUnauthorizedPluginAction
				}
			}
		}
	}

	// Check the method being called against the permitted operation, the
	// expiration time and IP address and return the result.
	authChecker := svc.Checker.Auth(macaroon.Slice{mac})
	_, err = authChecker.Allow(ctx, requiredPermissions...)

	// If the macaroon contains broad permissions and checks out, we're
	// done.
	if err == nil {
		return nil
	}

	// To also allow the special permission of "uri:<FullMethod>" to be a
	// valid permission, we need to check it manually in case there is no
	// broader scope permission defined.
	_, err = authChecker.Allow(ctx, bakery.Op{
		Entity: PermissionEntityCustomURI,
		Action: fullMethod,
	})
	return err
}

// Close closes the rootkeystorage of the macaroon service
func (s *Service) Close() error {
	return s.rks.Close()
}

// Thin-wrapper for the CreateUnlock function of the RootKeyStorage attribute of the Service
func (s *Service) CreateUnlock(password *[]byte) error {
	return s.rks.CreateUnlock(password)
}

// ContextWithRootKeyId passes the root key ID value to context
func ContextWithRootKeyId(ctx context.Context, value interface{}) context.Context {
	return context.WithValue(ctx, RootKeyIDContextKey, value)
}

// NewMacaroon is a wrapper around the Oven.NewMacaroon method and returns a freshly baked macaroon
func (s *Service) NewMacaroon(ctx context.Context, rootKeyId []byte, cav []checkers.Caveat, ops ...bakery.Op) (*bakery.Macaroon, error) {
	if len(rootKeyId) == 0 {
		return nil, ErrMissingRootKeyID
	}
	ctx = ContextWithRootKeyId(ctx, rootKeyId)
	if len(cav) == 0 {
		return s.Oven.NewMacaroon(ctx, bakery.LatestVersion, nil, ops...)
	}
	return s.Oven.NewMacaroon(ctx, bakery.LatestVersion, cav, ops...)
}

// ChangePassword calls the underlying root key store's ChangePassword and returns the result.
func (svc *Service) ChangePassword(oldPw, newPw []byte) error {
	return svc.rks.ChangePassword(oldPw, newPw)
}

// ListMacaroonIDs returns all the root key ID values except the value of
// encryptedKeyID.
func (svc *Service) ListMacaroonIDs(ctxt context.Context) ([][]byte, error) {
	return svc.rks.ListMacaroonIDs(ctxt)
}

// DeleteMacaroonID removes one specific root key ID. If the root key ID is
// found and deleted, it will be returned.
func (svc *Service) DeleteMacaroonID(ctxt context.Context,
	rootKeyID []byte) ([]byte, error) {
	return svc.rks.DeleteMacaroonID(ctxt, rootKeyID)
}

// SafeCopyMacaroon creates a copy of a macaroon that is safe to be used and
// modified. This is necessary because the macaroon library's own Clone() method
// is unsafe for certain edge cases, resulting in both the cloned and the
// original macaroons to be modified.
func SafeCopyMacaroon(mac *macaroon.Macaroon) (*macaroon.Macaroon, error) {
	macBytes, err := mac.MarshalBinary()
	if err != nil {
		return nil, err
	}

	newMac := &macaroon.Macaroon{}
	if err := newMac.UnmarshalBinary(macBytes); err != nil {
		return nil, err
	}

	return newMac, nil
}
