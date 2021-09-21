package main

import(
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"github.com/btcsuite/btcwallet/snacl"
	macaroon "gopkg.in/macaroon.v2"
)

const (
	encryptionPrefix = "snacl:"
)

type getPasswordFunc func(prompt string) ([]byte, error)

func decryptMacaroon(keyBase64, dataBase64 string, pw []byte) ([]byte, error) {
	keyData, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("Could not base64 decode encryption key: %v", err)
	}
	encryptedMac, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, fmt.Errorf("Could not base64 decode encrypted macaroon: %v", err)
	}
	key := &snacl.SecretKey{}
	err = key.Unmarshal(keyData)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshall encryption key: %v", err)
	}
	err = key.DeriveKey(&pw)
	if err != nil {
		return nil, fmt.Errorf("Could not derive encryption key possibly due to incorrect password: %v", err)
	}
	macBytes, err := key.Decrypt(encryptedMac)
	if err != nil {
		return nil, fmt.Errorf("Could not decrypt macaroon: %v", err)
	}
	return macBytes, nil
}

func loadMacaroon(pwCallback getPasswordFunc, macHex string) (*macaroon.Macaroon, error) {
	if len(strings.TrimSpace(macHex)) == 0 {
		return nil, fmt.Errorf("macaroon data is empty")
	}
	var (
		macBytes	[]byte
		err			error
	)
	if strings.HasPrefix(macHex, encryptionPrefix) {
		parts := strings.Split(macHex, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("Invalid encrypted macaroon. Format expected: 'snacl:<key_base64>:<encrypted_macaroon_base64>'")
		}
		pw, err := pwCallback("Enter macaroon encryption password: ")
		if err != nil {
			return nil, fmt.Errorf("Could not read password from terminal: %v", err)
		}
		macBytes, err = decryptMacaroon(parts[1], parts[2], pw)
		if err != nil {
			return nil, fmt.Errorf("Unable to decrypt macaroon: %v", err)
		}	
	} else {
		macBytes, err = hex.DecodeString(macHex)
		if err != nil {
			return nil, fmt.Errorf("Unable to hex decode macaroon: %v", err)
		}
	}
	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		return nil, fmt.Errorf("Unable to decode macaroon: %v", err)
	}
	return mac, nil
}
