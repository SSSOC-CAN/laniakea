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
package unlocker

import (
	"context"
	"crypto/sha256"
	"os"
	"reflect"

	"github.com/SSSOC-CAN/fmtd/api"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/kvdb"
	"github.com/SSSOC-CAN/fmtd/macaroons"
	bg "github.com/SSSOCPaulCote/blunderguard"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	e "github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrPasswordAlreadySet = bg.Error("password has already been set")
	ErrPasswordNotSet     = bg.Error("password has not been set")
	ErrWrongPassword      = bg.Error("wrong password")
	ErrUnlockTimeout      = bg.Error("login timed out")
	ErrBucketNotFound     = bg.Error("bucket not found")
	pwdKeyBucketName      = []byte("pwdkeys")
	pwdKeyID              = []byte("pwd")
)

type PasswordMsg struct {
	Password []byte
	Err      error
}

type UnlockerService struct {
	fmtrpc.UnimplementedUnlockerServer
	ps            *kvdb.DB
	PasswordMsgs  chan *PasswordMsg
	macaroonFiles []string
}

// Compile time check to ensure UnlockerService implements api.RestProxyService
var _ api.RestProxyService = (*UnlockerService)(nil)

// InitUnlockerService instantiates the UnlockerService
func InitUnlockerService(db *kvdb.DB, macaroonFiles []string) (*UnlockerService, error) {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(pwdKeyBucketName)
		return err
	}); err != nil {
		return nil, err
	}
	return &UnlockerService{ps: db, PasswordMsgs: make(chan *PasswordMsg, 1), macaroonFiles: macaroonFiles}, nil
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (u *UnlockerService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterUnlockerServer(grpcServer, u)
	return nil
}

// RegisterWithRestProxy registers the UnlockerService with the REST proxy
func (u *UnlockerService) RegisterWithRestProxy(ctx context.Context, mux *proxy.ServeMux, restDialOpts []grpc.DialOption, restProxyDest string) error {
	err := fmtrpc.RegisterUnlockerHandlerFromEndpoint(
		ctx, mux, restProxyDest, restDialOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

// setPassword will set the password if one has not already been set
func (u *UnlockerService) setPassword(password []byte, overwrite bool) error {
	u.ps.Mutex.Lock()
	defer u.ps.Mutex.Unlock()
	return u.ps.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(pwdKeyBucketName) // get the password bucket
		if bucket == nil {
			return ErrBucketNotFound
		}
		pwd := bucket.Get(pwdKeyID) //get the password kv pair
		if len(pwd) > 0 && !overwrite {
			return ErrPasswordAlreadySet
		}
		// no pwd has been set so creating a new one
		hash := sha256.Sum256(password)
		err := bucket.Put(pwdKeyID, hash[:])
		if err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})
}

// readPassword will read the password provided and compare to what's in the db
func (u *UnlockerService) readPassword(password []byte) error {
	u.ps.Mutex.Lock()
	defer u.ps.Mutex.Unlock()
	return u.ps.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(pwdKeyBucketName) // get the password bucket
		if bucket == nil {
			return ErrBucketNotFound
		}
		pwd := bucket.Get(pwdKeyID) //get the password kv pair
		if len(pwd) == 0 {
			return ErrPasswordNotSet
		}
		// pwd has been set so comparing
		hash := sha256.Sum256(password)
		if !reflect.DeepEqual(hash[:], pwd) {
			return ErrWrongPassword
		}
		return nil
	})
}

// Login will login a user
func (u *UnlockerService) Login(ctx context.Context, req *fmtrpc.LoginRequest) (*fmtrpc.LoginResponse, error) {
	err := u.readPassword(req.Password)
	if err == ErrPasswordNotSet {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	} else if err == ErrWrongPassword {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// We can now send the a PasswordMsg through the channel and return a successful loging message
	select {
	case u.PasswordMsgs <- &PasswordMsg{Password: req.Password, Err: nil}:
		return &fmtrpc.LoginResponse{}, nil
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, ErrUnlockTimeout.Error())
	}
}

// SetPassword will set the password of the kvdb if none has been set
func (u *UnlockerService) SetPassword(ctx context.Context, req *fmtrpc.SetPwdRequest) (*fmtrpc.SetPwdResponse, error) {
	err := u.setPassword(req.Password, false)
	if err == ErrPasswordAlreadySet {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// We can now send the SetPasswordMsg through the channel
	select {
	case u.PasswordMsgs <- &PasswordMsg{Password: req.Password, Err: nil}:
		return &fmtrpc.SetPwdResponse{}, nil

	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, ErrUnlockTimeout.Error())
	}
}

// ChangePassword takes the old password, validates it and sets the new password from the inputted new password only if a previous password has been set
func (u *UnlockerService) ChangePassword(ctx context.Context, req *fmtrpc.ChangePwdRequest) (*fmtrpc.ChangePwdResponse, error) {
	// first we check the validaty of the old password
	err := u.readPassword(req.CurrentPassword)
	if err == ErrPasswordNotSet {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	} else if err == ErrWrongPassword {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Next we set the new password
	err = u.setPassword(req.NewPassword, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if req.NewMacaroonRootKey {
		for _, file := range u.macaroonFiles {
			err := os.Remove(file)
			if err != nil {
				return nil, status.Error(codes.Internal, e.Wrap(err, "could not remove macaroon file").Error())
			}
		}
	}
	// Then we have to load the macaroon key-store, unlock it, change the old password and then shut it down
	macaroonService, err := macaroons.InitService(*u.ps, "fmtd")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = macaroonService.CreateUnlock(&req.CurrentPassword)
	if err != nil {
		closeErr := macaroonService.Close()
		if closeErr != nil {
			return nil, status.Error(codes.Internal, e.Wrap(closeErr, "error when closing macaroon service").Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = macaroonService.ChangePassword(req.CurrentPassword, req.NewPassword)
	if err != nil {
		closeErr := macaroonService.Close()
		if closeErr != nil {
			return nil, status.Error(codes.Internal, e.Wrap(closeErr, "error when closing macaroon service").Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = macaroonService.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, e.Wrap(err, "error when closing macaroon service").Error())
	}

	// We can now send the UnlockMsg through the channel
	select {
	case u.PasswordMsgs <- &PasswordMsg{Password: req.NewPassword, Err: nil}:
		return &fmtrpc.ChangePwdResponse{}, nil
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, ErrUnlockTimeout.Error())
	}
}
