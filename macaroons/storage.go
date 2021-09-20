package macaroons

import (
	"fmt"
	"sync"
	"github.com/btcsuite/btcwallet/snacl"
	bolt "go.etcd.io/bbolt"
)

var (
	rootKeyBucketName = []byte("macrootkeys")
	DefaultRootKeyID = []byte("0")
	encryptionKeyID = []byte("enckey")
	ErrAlreadyUnlocked = fmt.Errorf("Macaroon store already unlocked")
)

type RootKeyStorage struct{
	bolt.DB

	encKeyMutex	sync.RWMutex
	encKey *snacl.SecretKey // Don't roll your own crypto
}

// InitRootKeyStorage initializes the top level bucket within the bbolt db for macaroons
func InitRootKeyStorage(db *bolt.DB) (*RootKeyStorage, error) {
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