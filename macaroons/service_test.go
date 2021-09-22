package macaroons

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"testing"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/metadata"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	testOp = bakery.Op{
		Entity: "test",
		Action: "read",
	}
	testOpURI = bakery.Op{
		Entity: PermissionEntityCustomURI,
		Action: "Test",
	}
	testPw = []byte("hello") 
)

// createDummyRootKeyStore creates a dummy RootKeyStorage from the test password in a temporary directory
func createDummyRootKeyStore(t *testing.T) (string, *bolt.DB) {
	tempDir, err := ioutil.TempDir("", "macaroonstore-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	db, err := bolt.Open(path.Join(tempDir, "macaroon.db"), 0755, nil)
	if err != nil {
		t.Fatalf("Could not create macaroon.db in temporary directory: %v", err)
	}
	rks, err := InitRootKeyStorage(*db)
	if err != nil {
		t.Fatalf("Could not instantiate RootKeyStorage: %v", err)
	}
	defer rks.Close()
	err = rks.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Error creating unlock: %v", err)
	}
	return tempDir, db
}

// TestNewService instantiates a dummy service from a dummy RootKeyStorage and tests whether it functions as expected
func TestNewService(t *testing.T) {
	tempDir, db := createDummyRootKeyStore(t)
	defer db.Close()
	defer os.RemoveAll(tempDir)
	service, err := InitService(*db, "fmtd")
	if err != nil {
		t.Fatalf("Error creating new service: %v", err)
	}
	defer service.Close()
	err = service.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Could not unlock rks: %v", err)
	}
	// Test for missing root key Id
	_, err = service.NewMacaroon(context.TODO(), nil, false, nil, testOp)
	if err != ErrMissingRootKeyID {
		t.Fatalf("Received %v instead of ErrMissingRootKeyID", err)
	}

	// Test we can actually make a macaroon
	mac, err := service.NewMacaroon(context.TODO(), DefaultRootKeyID, false, nil, testOp)
	if err != nil {
		t.Fatalf("Error creating macaroon: %v", err)
	}
	// Check the macaroon isn't deffective
	if mac.Namespace().String() != "std:" {
		t.Fatalf("The macaroon has an invalid namespace: %s", mac.Namespace().String())
	}
}

// TestValidateMacaroon creates a dummy macaroon from a dummy service and validates it against test parameters
func TestValidateMacaroon(t *testing.T) {
	tempDir, db := createDummyRootKeyStore(t)
	defer db.Close()
	defer os.RemoveAll(tempDir)
	service, err := InitService(*db, "fmtd")
	if err != nil {
		t.Fatalf("Error creating new service: %v", err)
	}
	defer service.Close()
	err = service.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Could not unlock rks: %v", err)
	}
	mac, err := service.NewMacaroon(context.TODO(), DefaultRootKeyID, false, nil, testOp, testOpURI,)
	if err != nil {
		t.Fatalf("Could not bake new macaroon: %v", err)
	}
	macBinary, err := mac.M().MarshalBinary()
	if err != nil {
		t.Fatalf("Could not serialize macaroon: %v", err)
	}
	md := metadata.New(map[string]string{"macaroon": hex.EncodeToString(macBinary)})
	dummyContext := metadata.NewIncomingContext(context.Background(), md)
	err = service.ValidateMacaroon(dummyContext, []bakery.Op{testOp}, "Foo")
	if err != nil {
		t.Fatalf("Could not validate macaroon: %v", err)
	}
	err = service.ValidateMacaroon(dummyContext, []bakery.Op{{Entity: "Yikes"}}, "Test")
	if err != nil {
		t.Fatalf("Could not validate macaroon: %v", err)
	}
}