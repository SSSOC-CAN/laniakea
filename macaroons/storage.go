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
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"github.com/btcsuite/btcwallet/snacl"
	bolt "go.etcd.io/bbolt"
)

var (
	rootKeyBucketName = []byte("macrootkeys")
	RootKeyIDContextKey = contextKey{"rootkeyid"}
	RootKeyLen = 32
	DefaultRootKeyID = []byte("0")
	encryptionKeyID = []byte("enckey")
	ErrAlreadyUnlocked = fmt.Errorf("Macaroon store already unlocked")
	ErrContextRootKeyID = fmt.Errorf("Failed to read root key ID from context")
	ErrKeyValueForbidden = fmt.Errorf("Root key ID value is not allowed")
	ErrPasswordRequired = fmt.Errorf("a non-nil password is required")
	ErrStoreLocked = fmt.Errorf("macaroon store is locked")
)

type contextKey struct {
	Name string
}

type RootKeyStorage struct{
	bolt.DB

	encKeyMutex	sync.RWMutex
	encKey *snacl.SecretKey // Don't roll your own crypto
}

// InitRootKeyStorage initializes the top level bucket within the bbolt db for macaroons
func InitRootKeyStorage(db bolt.DB) (*RootKeyStorage, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(rootKeyBucketName)
		return err
	}); err != nil {
		return nil, err
	}
	return &RootKeyStorage{
		DB: db,
		encKey: nil,
	}, nil
}

// Get returns the root key for the given id. If the item is not there, it returns an error
func (r *RootKeyStorage) Get(_ context.Context, id []byte) ([]byte, error) {
	r.encKeyMutex.RLock()
	defer r.encKeyMutex.RUnlock()
	if r.encKey == nil {
		return nil, ErrStoreLocked
	}
	var rootKey []byte
	err := r.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(rootKeyBucketName) // get the rootkey bucket
		if bucket == nil {
			return fmt.Errorf("Root key bucket not found")
		}
		dbKey := bucket.Get(id) //get the encryption key kv pair
		if len(dbKey) == 0 {
			return fmt.Errorf("Root key with id %s doesn't exist", string(id))
		}
		decKey, err := r.encKey.Decrypt(dbKey)
		if err != nil {
			return err
		}
		rootKey = make([]byte, len(decKey))
		copy(rootKey[:], decKey)
		return nil
	})
	if err != nil {
		rootKey = nil
		return nil, err
	}
	return rootKey, nil
}

// RootKeyIDFromContext retrieves the root key ID from context using the key
// RootKeyIDContextKey.
func RootKeyIDFromContext(ctx context.Context) ([]byte, error) {
	id, ok := ctx.Value(RootKeyIDContextKey).([]byte)
	if !ok {
		return nil, ErrContextRootKeyID
	}
	if len(id) == 0 {
		return nil, ErrMissingRootKeyID
	}
	return id, nil
}

// generateAndStoreNewRootKey creates a new random RootKeyLen-byte root key,
// encrypts it with the given encryption key and stores it in the bucket.
// Any previously set key will be overwritten.
func generateAndStoreNewRootKey(bucket *bolt.Bucket, id []byte,
	key *snacl.SecretKey) ([]byte, error) {

	rootKey := make([]byte, RootKeyLen)
	if _, err := io.ReadFull(rand.Reader, rootKey); err != nil {
		return nil, err
	}

	encryptedKey, err := key.Encrypt(rootKey)
	if err != nil {
		return nil, err
	}
	return rootKey, bucket.Put(id, encryptedKey)
}

// Implements RootKey from the bakery.RootKeyStorage interface
func (r* RootKeyStorage) RootKey(ctx context.Context) ([]byte, []byte, error) {
	r.encKeyMutex.RLock()
	defer r.encKeyMutex.RUnlock()
	if r.encKey == nil {
		return nil, nil, ErrStoreLocked
	}
	id, err := RootKeyIDFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(id, encryptionKeyID) {
		return nil, nil, ErrKeyValueForbidden
	}
	var rootKey []byte
	err = r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(rootKeyBucketName) // get the rootkey bucket
		if bucket == nil {
			return fmt.Errorf("Root key bucket not found")
		}
		dbKey := bucket.Get(id) //get the encryption key kv pair
		if len(dbKey) != 0 {
			decKey, err := r.encKey.Decrypt(dbKey)
			if err != nil {
				return err
			}

			rootKey = make([]byte, len(decKey))
			copy(rootKey[:], decKey[:])
			return nil
		}
		newKey, err := generateAndStoreNewRootKey(bucket, id, r.encKey)
		rootKey = newKey
		return err
	})
	if err != nil {
		rootKey = nil
		return nil, nil, err
	}
	return rootKey, id, err
}

// CreateUnlock sets an encryption key if one isn't already set or checks if the password is correct for the existing encryption key.
func (r *RootKeyStorage) CreateUnlock(password *[]byte) error {
	r.encKeyMutex.Lock()
	defer r.encKeyMutex.Unlock()
	if r.encKey != nil {
		return ErrAlreadyUnlocked
	}
	if password == nil {
		return ErrPasswordRequired
	}
	return r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(rootKeyBucketName) // get the rootkey bucket
		if bucket == nil {
			return fmt.Errorf("Root key bucket not found")
		}
		dbKey := bucket.Get(encryptionKeyID) //get the encryption key kv pair
		if len(dbKey) > 0 {
			// dbKey has already been set
			encKey := &snacl.SecretKey{}
			err := encKey.Unmarshal(dbKey)
			if err != nil {
				return err
			}
			err = encKey.DeriveKey(password)
			if err != nil {
				return err
			}
			r.encKey = encKey
			return nil
		}
		// no key has been set so creating a new one 
		encKey, err := snacl.NewSecretKey(
			password, snacl.DefaultN, snacl.DefaultR, snacl.DefaultP,
		)
		if err != nil {
			return err
		}
		err = bucket.Put(encryptionKeyID, encKey.Marshal())
		if err != nil {
			return err
		}
		r.encKey = encKey
		return nil
	})
}

// Close resets the encryption key in memory
func (r *RootKeyStorage) Close() error {
	r.encKeyMutex.Lock()
	defer r.encKeyMutex.Unlock()
	if r.encKey != nil {
		r.encKey.Zero()
		r.encKey = nil
	}
	return nil
}