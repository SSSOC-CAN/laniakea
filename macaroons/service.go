package macaroons

import (
	"context"
	bolt "go.etcd.io/bbolt"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

type Service struct {
	bakery.Bakery

	rks *RootKeyStorage
}

// InitService returns initializes the rootkeystorage for the Macaroon service and returns the initialized service
func InitService(db *bolt.DB, location string) (*Service, error) {
	rks, err := InitRootKeyStorage(db)
	if err != nil {
		return nil, err
	}
	bakeryParams := bakery.BakeryParams{
		Location: location,
		RootKeyStore: rks,
		Locator: nil,
		Key: nil,
	}
	service := bakery.New(bakeryParams)
	return &Service{
		Bakery: *service,
		rks: rks,
	}
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
func (s *Service) NewMacaroon(ctx context.Context, rootKeyId []byte, ops ...bakery.Op) (*bakery.Macaroon, error) {
	if len(rootKeyId) == 0 {
		return nil, fmt.Errorf("Missing root key ID")
	}
	ctx = ContextWithRootKeyId(ctx, rootKeyID)
	return s.Oven.NewMacaroon(ctx, bakery.LatestVersion, nil, ops...)
}