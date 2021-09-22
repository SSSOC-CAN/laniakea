package unlocker

import (
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"sync"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

var (
	ErrPasswordAlreadySet = fmt.Errorf("Password has already been set.")
	ErrPasswordNotSet = fmt.Errorf("Password has not been set.")
	ErrWrongPassword = fmt.Errorf("Wrong password.")
	pwdKeyBucketName = []byte("pwdkeys")
	pwdKeyID = []byte("pwd")
)

type LoginMsg struct {
	Password	*[]byte
	Err			error
}

type UnlockerService struct {
	fmtrpc.UnimplementedUnlockerServer
	ps *PasswordStorage
	LoginMsgs chan *LoginMsg
}

type PasswordStorage struct {
	bolt.DB
	passwordMutex	sync.RWMutex
}

// InitUnlockerService instantiates the UnlockerService
func InitUnlockerService(db bolt.DB) (*UnlockerService, error) {
	ps, err := initPasswordStore(db)
	if err != nil {
		return nil, err
	}
	return &UnlockerService{ps: ps, LoginMsgs: make(chan *LoginMsg, 1)}, nil
}

// initPasswordStore initializes the bucket for password storage within the db
func initPasswordStore(db bolt.DB) (*PasswordStorage, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(pwdKeyBucketName)
		return err
	}); err != nil {
		return nil, err
	}
	return &PasswordStorage{
		DB: db,
	}, nil
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (u *UnlockerService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterUnlockerServer(grpcServer, u)
	return nil
}

// setPassword will set the password if one has not already been set
func (u *UnlockerService) setPassword(password *[]byte) error {
	u.ps.passwordMutex.Lock()
	defer u.ps.passwordMutex.Unlock()
	return u.ps.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(pwdKeyBucketName) // get the password bucket
		if bucket == nil {
			return fmt.Errorf("Password bucket not found")
		}
		pwd := bucket.Get(pwdKeyID) //get the password kv pair
		if len(pwd) > 0 {
			return ErrPasswordAlreadySet
		}
		// no pwd has been set so creating a new one
		hash := sha256.Sum256(*password)
		err := bucket.Put(pwdKeyID, hash[:])
		if err != nil {
			return err
		}
		return nil
	})
}

// readPassword will read the password provided and compare to what's in the db
func (u *UnlockerService) readPassword(password *[]byte) error {
	u.ps.passwordMutex.Lock()
	defer u.ps.passwordMutex.Unlock()
	return u.ps.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(pwdKeyBucketName) // get the password bucket
		if bucket == nil {
			return fmt.Errorf("Password bucket not found")
		}
		pwd := bucket.Get(pwdKeyID) //get the password kv pair
		if len(pwd) == 0 {
			return u.setPassword(password)
		}
		// pwd has been set so comparing
		hash := sha256.Sum256(*password)
		if !reflect.DeepEqual(hash[:], pwd) {
			return ErrWrongPassword
		}
		return nil
	})
}



// Login will login a user
func (u *UnlockerService) Login(ctx context.Context, req *fmtrpc.LoginRequest) (*fmtrpc.LoginResponse, error) {
	err := u.setPassword(&req.Password)
	if err != ErrPasswordAlreadySet && err != nil {
		return nil, err
	}
	if err == ErrPasswordAlreadySet {
		err = u.readPassword(&req.Password)
		if err != nil {
			return nil, err
		}
	}
	// We can now send the a LoginMsg through the channel and return a successful loging message
	select {
	case u.LoginMsgs <- &LoginMsg{Password: &req.Password, Err: nil}:
		return &fmtrpc.LoginResponse{Msg: "Login successful"}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("Login timed out.")
	}
	
}