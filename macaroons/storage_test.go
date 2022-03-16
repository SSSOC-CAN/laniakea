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
package macaroons

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/btcsuite/btcwallet/snacl"
)

var (
	defaultRootKeyIDContext = ContextWithRootKeyId(
		context.Background(), DefaultRootKeyID,
	)
)

// createDummyRootKeyStore returns a temporary directory, a cleanup function and an instantiated RootKeyStorage
func createDummyRootKeyStorage(t *testing.T) (string, func(), *RootKeyStorage) {
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
	cleanUp := func() {
		_ = rks.Close()
		_ = db.Close()
	}
	return tempDir, cleanUp, rks
}

// TestStore tests the normal use cases of the store like creating, unlocking,
// reading keys and closing it.
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
	key, id, err := rks.RootKey(defaultRootKeyIDContext)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	key2, err := rks.Get(defaultRootKeyIDContext, id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(key, key2) {
		t.Fatalf("Key mismatch: expected: %v, received: %v: %v", key, key2, err)
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
	db, err := kvdb.NewDB(path.Join(tempDir, "macaroon.db"))
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
	_, _, err = rks.RootKey(defaultRootKeyIDContext)
	if err != ErrStoreLocked {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, err = rks.Get(defaultRootKeyIDContext, nil)
	if err != ErrStoreLocked {
		t.Fatalf("Unexpected error: %v", err)
	}
	err = rks.CreateUnlock(&pw)
	if err != nil {
		t.Fatalf("Could not unlock key store: %v", err)
	}
	key, err = rks.Get(defaultRootKeyIDContext, rootId)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(key2, key) {
		t.Fatalf("The value of key2 %v should be the same as key %v: %v", key2, key, err)
	}
	key, id2, err := rks.RootKey(defaultRootKeyIDContext)
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

// TestGenerateNewRootKey tests that a root key can be replaced with a new one in the store
func TestGenerateNewRootKey(t *testing.T) {
	tempDir, cleanUp, rks := createDummyRootKeyStorage(t)
	defer cleanUp()
	defer os.RemoveAll(tempDir)
	// test that we can't generate new root keys when the store is locked
	err := rks.GenerateNewRootKey()
	if err != ErrStoreLocked {
		t.Errorf("Unexpected error: %v", err)
	}
	// unlock store
	pw := []byte("test")
	err = rks.CreateUnlock(&pw)
	if err != nil {
		t.Errorf("Could not unlock root key store: %v", err)
	}
	// get root key
	oldRootKey, _, err := rks.RootKey(defaultRootKeyIDContext)
	if err != nil {
		t.Errorf("Could not get root key: %v", err)
	}
	// generate new root key
	err = rks.GenerateNewRootKey()
	if err != nil {
		t.Errorf("Could not generate new root key: %v", err)
	}
	// get new root key and compare
	newRootKey, _, err := rks.RootKey(defaultRootKeyIDContext)
	if err != nil {
		t.Errorf("Could not get root key: %v", err)
	}
	if reflect.DeepEqual(oldRootKey, newRootKey) {
		t.Fatalf("Root keys are equal")
	}
}

// TestChangePasswird tests that the password for the store can be changed without changing the root key
func TestChangePassword(t *testing.T) {
	tempDir, cleanUp, rks := createDummyRootKeyStorage(t)
	defer cleanUp()
	defer os.RemoveAll(tempDir)

	// test that the store must first be unlocked
	err := rks.GenerateNewRootKey()
	if err != ErrStoreLocked {
		t.Errorf("Unexpected error: %v", err)
	}
	// unlock store
	pw := []byte("test")
	err = rks.CreateUnlock(&pw)
	if err != nil {
		t.Errorf("Could not unlock root key store: %v", err)
	}
	// get root key
	oldRootKey, _, err := rks.RootKey(defaultRootKeyIDContext)
	if err != nil {
		t.Errorf("Could not get root key: %v", err)
	}
	// test for required password
	err = rks.ChangePassword(nil, nil)
	if err != ErrPasswordRequired {
		t.Errorf("Unexpected error: %v", err)
	}
	// test for correct old password
	wrongPw := []byte("wrong")
	newPw := []byte("newpassword")
	err = rks.ChangePassword(wrongPw, newPw)
	if err != snacl.ErrInvalidPassword {
		t.Errorf("Unexpected error: %v", err)
	}
	// change password
	err = rks.ChangePassword(pw, newPw)
	if err != nil {
		t.Errorf("Could not change password: %v", err)
	}
	// close store and db
	cleanUp()
	// Try reopening
	db, err := kvdb.NewDB(path.Join(tempDir, "macaroon.db"))
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
	err = rks.CreateUnlock(&newPw)
	if err != nil {
		t.Errorf("Error unlocking store with new password: %v", err)
	}
	rootKey, _, err := rks.RootKey(defaultRootKeyIDContext)
	if err != nil {
		t.Errorf("Error getting root key: %v", err)
	}
	if !reflect.DeepEqual(oldRootKey, rootKey) {
		t.Errorf("Root keys are unequal old rk: %v new rk: %v", oldRootKey, rootKey)
	}
}