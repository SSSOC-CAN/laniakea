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
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/metadata"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/checkers"
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
func createDummyRootKeyStore(t *testing.T) (string, *kvdb.DB) {
	tempDir, err := ioutil.TempDir("", "macaroonstore-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	db, err := kvdb.NewDB(path.Join(tempDir, "macaroon.db"))
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
	service, err := InitService(*db, "laniakea", zerolog.Nop(), []string{})
	if err != nil {
		t.Fatalf("Error creating new service: %v", err)
	}
	defer service.Close()
	err = service.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Could not unlock rks: %v", err)
	}
	// Test for missing root key Id
	_, err = service.NewMacaroon(context.TODO(), nil, nil, testOp)
	if err != ErrMissingRootKeyID {
		t.Fatalf("Received %v instead of ErrMissingRootKeyID", err)
	}

	// Test we can actually make a macaroon
	mac, err := service.NewMacaroon(context.TODO(), DefaultRootKeyID, nil, testOp)
	if err != nil {
		t.Fatalf("Error creating macaroon: %v", err)
	}
	// Check the macaroon isn't deffective
	if mac.Namespace().String() != "std:" {
		t.Fatalf("The macaroon has an invalid namespace: %s", mac.Namespace().String())
	}
}

var (
	valPluginNames = []string{"plugin-1", "plugin_two", "PlUgIn_ThR33"}
)

// TODO:SSSOCPaulCote - Update tests for new plugin validation
// TestValidateMacaroon creates a dummy macaroon from a dummy service and validates it against test parameters
func TestValidateMacaroon(t *testing.T) {
	tempDir, db := createDummyRootKeyStore(t)
	defer db.Close()
	defer os.RemoveAll(tempDir)
	service, err := InitService(*db, "laniakea", zerolog.Nop(), valPluginNames)
	if err != nil {
		t.Fatalf("Error creating new service: %v", err)
	}
	defer service.Close()
	err = service.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Could not unlock rks: %v", err)
	}
	mac, err := service.NewMacaroon(context.TODO(), DefaultRootKeyID, []checkers.Caveat{PluginCaveat(valPluginNames)}, testOp, testOpURI)
	if err != nil {
		t.Fatalf("Could not bake new macaroon: %v", err)
	}
	macBinary, err := mac.M().MarshalBinary()
	if err != nil {
		t.Fatalf("Could not serialize macaroon: %v", err)
	}
	t.Run("error-no metadata", func(t *testing.T) {
		err = service.ValidateMacaroon(context.Background(), []bakery.Op{testOp}, "Foo")
		if err != ErrMetadataFromContext {
			t.Errorf("Unexpected error when calling ValidateMacaroon: %v", err)
		}
	})
	t.Run("error-empty mac data", func(t *testing.T) {
		md := metadata.New(map[string]string{"macaroon": ""})
		dummyContext := metadata.NewIncomingContext(context.Background(), md)
		err = service.ValidateMacaroon(dummyContext, []bakery.Op{testOp}, "Foo")
		if err.Error() != "empty macaroon data" {
			t.Errorf("Unexpected error when calling ValidateMacaroon: %v", err)
		}
	})
	md := metadata.New(map[string]string{"macaroon": hex.EncodeToString(macBinary)})
	dummyContext := metadata.NewIncomingContext(context.Background(), md)
	t.Run("valid-test ops", func(t *testing.T) {
		err = service.ValidateMacaroon(dummyContext, []bakery.Op{testOp}, "Foo")
		if err != nil {
			t.Fatalf("Could not validate macaroon: %v", err)
		}
		err = service.ValidateMacaroon(dummyContext, []bakery.Op{{Entity: "Yikes"}}, "Test")
		if err != nil {
			t.Fatalf("Could not validate macaroon: %v", err)
		}
	})
	t.Run("error-plugin key not in context", func(t *testing.T) {
		err = service.ValidateMacaroon(dummyContext, []bakery.Op{testOp}, knownPluginMethods[0])
		if err != ErrKeyNotInContext {
			t.Errorf("Unexpected error when calling ValidateMacaroon: %v", err)
		}
	})
	t.Run("error-unauthorized plugin action", func(t *testing.T) {
		pluginValues := []string{"not-a-valid-name", "all"}
		for _, v := range pluginValues {
			ctx := context.WithValue(dummyContext, PluginContextKey, v)
			err = service.ValidateMacaroon(ctx, []bakery.Op{testOp}, knownPluginMethods[0])
			if err != ErrUnauthorizedPluginAction {
				t.Errorf("Unexpected error when calling ValidateMacaroon: %v", err)
			}
		}
	})
	t.Run("valid-plugin actions", func(t *testing.T) {
		pluginValues := valPluginNames[:]
		for _, v := range pluginValues {
			ctx := context.WithValue(dummyContext, PluginContextKey, v)
			err = service.ValidateMacaroon(ctx, []bakery.Op{testOp}, knownPluginMethods[0])
			if err != nil {
				t.Errorf("Unexpected error when calling ValidateMacaroon: %v", err)
			}
		}
	})
	mac, err = service.NewMacaroon(context.TODO(), DefaultRootKeyID, []checkers.Caveat{PluginCaveat([]string{"all"})}, testOp, testOpURI)
	if err != nil {
		t.Fatalf("Could not bake new macaroon: %v", err)
	}
	macBinary, err = mac.M().MarshalBinary()
	if err != nil {
		t.Fatalf("Could not serialize macaroon: %v", err)
	}
	md = metadata.New(map[string]string{"macaroon": hex.EncodeToString(macBinary)})
	dummyContext = metadata.NewIncomingContext(context.Background(), md)
	t.Run("valid-plugin actions 2", func(t *testing.T) {
		pluginValues := valPluginNames[:]
		pluginValues = append(pluginValues, "all")
		pluginValues = append(pluginValues, "not-a-registered-plugin")
		for _, v := range pluginValues {
			ctx := context.WithValue(dummyContext, PluginContextKey, v)
			err = service.ValidateMacaroon(ctx, []bakery.Op{testOp}, knownPluginMethods[0])
			if err != nil {
				t.Errorf("Unexpected error when calling ValidateMacaroon: %v", err)
			}
		}
	})
}

// TestListMacaroonIDs checks that ListMacaroonIDs returns the expected result
func TestListMacaroonIDs(t *testing.T) {
	tempDir, db := createDummyRootKeyStore(t)
	defer db.Close()
	defer os.RemoveAll(tempDir)
	service, err := InitService(*db, "laniakea", zerolog.Nop(), []string{})
	if err != nil {
		t.Fatalf("Error creating new service: %v", err)
	}
	defer service.Close()
	err = service.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Could not unlock rks: %v", err)
	}

	expectedIDs := [][]byte{{1}, {2}, {3}}
	for _, v := range expectedIDs {
		_, err := service.NewMacaroon(context.TODO(), v, nil, testOp)
		if err != nil {
			t.Errorf("Error creating macaroon from service: %v", err)
		}
	}

	ids, _ := service.ListMacaroonIDs(context.TODO())
	if !reflect.DeepEqual(expectedIDs, ids) {
		t.Errorf("root key IDs mismatch expected: %v actual: %v", expectedIDs, ids)
	}
}

// TestDeleteMacaroonID tests if we can remove the specific root key ID
func TestDeleteMacaroonID(t *testing.T) {
	ctxb := context.Background()
	tempDir, db := createDummyRootKeyStore(t)
	defer db.Close()
	defer os.RemoveAll(tempDir)
	service, err := InitService(*db, "laniakea", zerolog.Nop(), []string{})
	if err != nil {
		t.Fatalf("Error creating new service: %v", err)
	}
	defer service.Close()
	err = service.CreateUnlock(&testPw)
	if err != nil {
		t.Fatalf("Could not unlock rks: %v", err)
	}
	// Checks that removing encryptedKeyID returns an error.
	encryptedKeyID := []byte("enckey")
	_, err = service.DeleteMacaroonID(ctxb, encryptedKeyID)
	if err != ErrDeletionForbidden {
		t.Errorf("Unexpected error: %v", err)
	}
	// Check if removing DefaultRootKeyID returns an error
	_, err = service.DeleteMacaroonID(ctxb, DefaultRootKeyID)
	if err != ErrDeletionForbidden {
		t.Errorf("Unexpected error: %v", err)
	}
	// Check that removing empty key id returns an error
	_, err = service.DeleteMacaroonID(ctxb, []byte{})
	if err != ErrMissingRootKeyID {
		t.Errorf("Unexpected error: %v", err)
	}
	// Check that returning non-existing ids returns nil
	nonExistingID := []byte("doesnt-exist")
	deletedID, err := service.DeleteMacaroonID(ctxb, nonExistingID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if deletedID != nil {
		t.Errorf("Unexpected return value: %v", deletedID)
	}
	// create 3 new macaroons and delete one
	expectedIDs := [][]byte{{1}, {2}, {3}}
	for _, v := range expectedIDs {
		_, err := service.NewMacaroon(context.TODO(), v, nil, testOp)
		if err != nil {
			t.Errorf("Error creating macaroon from service: %v", err)
		}
	}
	deletedID, err = service.DeleteMacaroonID(ctxb, expectedIDs[0])
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expectedIDs[0], deletedID) {
		t.Errorf("Unexpected return value: %v", deletedID)
	}
	// check that ID is deleted
	ids, _ := service.ListMacaroonIDs(context.TODO())
	if !reflect.DeepEqual(expectedIDs[1:], ids) {
		t.Errorf("root key IDs mismatch expected: %v actual: %v", expectedIDs, ids)
	}
}

// TestCloneMacaroons test that macaroons can be cloned correcty and any modifications do not affect the original
func TestCloneMacaroons(t *testing.T) {
	constraintFunc := TimeoutConstraint(3)

	testMac := createDummyMacaroon(t)
	err := constraintFunc(testMac)
	if err != nil {
		t.Errorf("Unexpected error adding constraint to macaroon: %v", err)
	}
	if loc := testMac.Caveats()[0].Location; loc != "" {
		t.Errorf("Expected caveat location to be empty, found: %s", loc)
	}
	newMacCred, err := NewMacaroonCredential(testMac)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	newMac := newMacCred.Macaroon
	if loc := newMac.Caveats()[0].Location; loc != "" {
		t.Errorf("Expected caveat location to be empty, found: %s", loc)
	}
	testMacBytes, err := testMac.MarshalBinary()
	if err != nil {
		t.Errorf("Unexpeted error when getting macaroon bytes: %v", err)
	}
	newMacBytes, err := newMac.MarshalBinary()
	if err != nil {
		t.Errorf("Unexpeted error when getting macaroon bytes: %v", err)
	}
	if !reflect.DeepEqual(testMacBytes, newMacBytes) {
		t.Errorf("Macaroon bytes mismatch, expected: %v actual: %v", testMacBytes, newMacBytes)
	}
	testMac.Caveats()[0].Location = "mars"
	if loc := testMac.Caveats()[0].Location; loc != "mars" {
		t.Errorf("Expected caveat location to be mars, found: %s", loc)
	}
	if loc := newMac.Caveats()[0].Location; loc != "" {
		t.Errorf("Expected caveat location to be empty, found: %s", loc)
	}
}
