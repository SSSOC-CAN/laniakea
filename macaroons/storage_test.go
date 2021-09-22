package macaroons

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"github.com/btcsuite/btcwallet/snacl"
	bolt "go.etcd.io/bbolt"
)

// createDummyRootKeyStore returns a temporary directory, a cleanup function and an instantiated RootKeyStorage
func createDummyRootKeyStorage(t *testing.T) (string, func(), *RootKeyStorage) {
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
	cleanUp := func() {
		_ = rks.Close()
		_ = db.Close()
	}
	return tempDir, cleanUp, rks
}

func TestStorage(t *testing.T) {
	tempDir, cleanUp, rks := createDummyRootKeyStorage(t)
	defer cleanUp()
	defer os.RemoveAll(tempDir)

	_, _, err := rks.RootKey(context.TODO())
	if err != ErrStoreLocked {
		t.Fatalf("Unexpected error: %v", err)
	}
	pw := []byte("password")
	err = rks.CreateUnlock(&pw)
	if err != nil {
		t.Fatalf("Could not unlock key store: %v", err)
	}

	// Context with no root key id
	_, _, err = rks.RootKey(context.TODO())
	if err != ErrContextRootKeyID {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Try empty key ID
	emptyKeyID := make([]byte, 0)
	ctx := ContextWithRootKeyId(context.TODO(), emptyKeyID)
	_, _, err = rks.RootKey(ctx)
	if err != ErrMissingRootKeyID {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Try with bad key ID
	encryptedKeyID := []byte("enckey")
	ctx = ContextWithRootKeyId(context.TODO(), encryptedKeyID)
	_, _, err = rks.RootKey(ctx)
	if err != ErrKeyValueForbidden {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Now the real deal
	key, id, err := rks.RootKey(ContextWithRootKeyId(context.Background(), DefaultRootKeyID))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	rootId := id
	if !reflect.DeepEqual(rootId, DefaultRootKeyID) {
		t.Fatalf("The value of id %v should be the same as DefaultRootKeyId %v: %v", id, DefaultRootKeyID, err)
	}

	// Already unlocked test
	dummyPw := []byte("abcdefgh")
	err = rks.CreateUnlock(&dummyPw)
	if err != ErrAlreadyUnlocked {
		t.Fatalf("Unexpected error: %v", err)
	}

	cleanUp()

	// Try reopening
	db, err := bolt.Open(path.Join(tempDir, "macaroon.db"), 0755, nil)
	if err != nil {
		t.Fatalf("Could not create/open macaroon.db in temporary directory: %v", err)
	}
	rks, err = InitRootKeyStorage(*db)
	if err != nil {
		t.Fatalf("Could not instantiate RootKeyStorage: %v", err)
	}
	defer func() {
		db.Close()
		rks.Close()
	}()
	err = rks.CreateUnlock(&dummyPw)
	if err != snacl.ErrInvalidPassword {
		t.Fatalf("Unexpected error: %v", err)
	}
	err = rks.CreateUnlock(nil)
	if err != ErrPasswordRequired {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, _, err = rks.RootKey(ContextWithRootKeyId(context.Background(), DefaultRootKeyID))
	if err != ErrStoreLocked {
		t.Fatalf("Unexpected error: %v", err)
	}
	err = rks.CreateUnlock(&pw)
	if err != nil {
		t.Fatalf("Could not unlock key store: %v", err)
	}
	key2, id2, err := rks.RootKey(ContextWithRootKeyId(context.Background(), DefaultRootKeyID))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(key2, key) {
		t.Fatalf("The value of key2 %v should be the same as key %v: %v", key2, key, err)
	}
	if !reflect.DeepEqual(id2, rootId) {
		t.Fatalf("The value of id2 %v should be the same as id %v: %v", id2, rootId, err)
	}	
}