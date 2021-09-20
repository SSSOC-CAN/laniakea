package macaroons

import (
	"context"
	"fmt"
	"sync"
	"github.com/btcsuite/btcwallet/snacl"
	bolt "go.etcd.io/bbolt"
)

var (
	rootKeyBucketName = []byte("macrootkeys")
	RootKeyIDContextKey = contextKey{"rootkeyid"}
	DefaultRootKeyID = []byte("0")
	encryptionKeyID = []byte("enckey")
	ErrAlreadyUnlocked = fmt.Errorf("Macaroon store already unlocked")
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
	var dbKey []byte
	err := r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(rootKeyBucketName) // get the rootkey bucket
		if bucket == nil {
			return fmt.Errorf("Root key bucket not found")
		}
		dbKey = bucket.Get(id) //get the encryption key kv pair
		return nil
	})
	return dbKey, err
}

// Implements RootKey from the bakery.RootKeyStorage interface
func (r* RootKeyStorage) RootKey(_ context.Context) ([]byte, []byte, error) {
	r.encKeyMutex.Lock()
	defer r.encKeyMutex.Unlock()
	if r.encKey != nil {
		return nil, nil, ErrAlreadyUnlocked
	}
	id := encryptionKeyID
	var dbKey []byte
	err := r.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(rootKeyBucketName) // get the rootkey bucket
		if bucket == nil {
			return fmt.Errorf("Root key bucket not found")
		}
		dbKey = bucket.Get(encryptionKeyID) //get the encryption key kv pair
		return nil
	})
	return dbKey, id, err
}

// CreateUnlock sets an encryption key if one isn't already set or checks if the password is correct for the existing encryption key.
func (r *RootKeyStorage) CreateUnlock(password *[]byte) error {
	r.encKeyMutex.Lock()
	defer r.encKeyMutex.Unlock()
	if r.encKey != nil {
		return ErrAlreadyUnlocked
	}
	if password == nil {
		return fmt.Errorf("a non-nil password is required")
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