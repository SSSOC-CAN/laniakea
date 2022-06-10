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
package auth

import(
	"encoding/base64"
	"encoding/hex"
	e "github.com/pkg/errors"
	"fmt"
	"strings"
	"github.com/btcsuite/btcwallet/snacl"
	"github.com/SSSOC-CAN/fmtd/errors"
	macaroon "gopkg.in/macaroon.v2"
)

const (
	encryptionPrefix = "snacl:"
	ErrEmptyMac = errors.Error("macaroon data is empty")
	ErrInvalidMacEncrypt = errors.Error("invalid encrypted macaroon. Format expected: 'snacl:<key_base64>:<encrypted_macaroon_base64>'")
)

type getPasswordFunc func(prompt string) ([]byte, error)

// decryptMacaroon will take a password and derive the priv key using the provided password to then decrypt the macaroon 
func decryptMacaroon(keyBase64, dataBase64 string, pw []byte) ([]byte, error) {
	keyData, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, e.Wrap(err, "could not base64 decode encryption key")
	}
	encryptedMac, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, e.Wrap(err, "could not base64 decode encypted macaroon")	
	}
	key := &snacl.SecretKey{}
	err = key.Unmarshal(keyData)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall encryption key: %v", err)
		return nil, e.Wrap(err, "could not unmarshall encryption key")
	}
	err = key.DeriveKey(&pw)
	if err != nil {
		return nil, e.Wrap(err, "could not derive encryption key")
	}
	macBytes, err := key.Decrypt(encryptedMac)
	if err != nil {
		return nil, e.Wrap(err, "could not decrypt macaroon")
	}
	return macBytes, nil
}

// LoadMacaroon takes a password prompting function and hex encoded macaroon and returns an instantiated macaroon object
func LoadMacaroon(pwCallback getPasswordFunc, macHex string) (*macaroon.Macaroon, error) {
	if len(strings.TrimSpace(macHex)) == 0 {
		return nil, ErrEmptyMac
	}
	var (
		macBytes	[]byte
		err			error
	)
	if strings.HasPrefix(macHex, encryptionPrefix) {
		parts := strings.Split(macHex, ":")
		if len(parts) != 3 {
			return nil, ErrInvalidMacEncrypt
		}
		pw, err := pwCallback("Enter macaroon encryption password: ")
		if err != nil {
			return nil, e.Wrap(err, "could not read password from terminal")
		}
		macBytes, err = decryptMacaroon(parts[1], parts[2], pw)
		if err != nil {
			return nil, err
		}	
	} else {
		macBytes, err = hex.DecodeString(macHex)
		if err != nil {
			return nil, e.Wrap(err, "unable to hex decode macaroon")
		}
	}
	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		return nil, e.Wrap(err, "unable to decode macaroon")
	}
	return mac, nil
}