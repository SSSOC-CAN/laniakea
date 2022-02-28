package cert

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"github.com/SSSOC-CAN/fmtd/utils"
)

var (
	defaultTestOrgString = "test org"
)

// TestGenCertPair tests whether we can successfully generate a TLS certificate/key pair
func TestGenCertPair(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "tlsstuff-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempTLSCertPath := path.Join(tempDir, "tls.cert")
	tempTLSKeyPath := path.Join(tempDir, "tls.key")
	err = GenCertPair(defaultTestOrgString, tempTLSCertPath, tempTLSKeyPath, defaultTLSCertDuration, make([]string, 0))
	if err != nil {
		t.Fatalf("Could not generate certificate/key pair: %v", err)
	}
	if !utils.FileExists(tempTLSCertPath) || !utils.FileExists(tempTLSKeyPath) {
		t.Fatalf("TLS certificate/key pair files not found.")
	}
}

// TestLoadCertificate tests if we can load a TLS certificate/key from file
func TestLoadCertificate(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "tlsstuff-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempTLSCertPath := path.Join(tempDir, "tls.cert")
	tempTLSKeyPath := path.Join(tempDir, "tls.key")
	err = GenCertPair(defaultTestOrgString, tempTLSCertPath, tempTLSKeyPath, defaultTLSCertDuration, make([]string, 0))
	if err != nil {
		t.Fatalf("Could not generate certificate/key pair: %v", err)
	}
	if !utils.FileExists(tempTLSCertPath) || !utils.FileExists(tempTLSKeyPath) {
		t.Fatalf("TLS certificate/key pair files not found.")
	}
	_, _, err = LoadCertificate(tempTLSCertPath, tempTLSKeyPath)
	if err != nil {
		t.Fatalf("Could not load certificate/key pair: %v", err)
	}
}

// TestGetTLSConfig tests whether we can load TLS config from existing TLS cert/key pair
// and if no file is provided, if a TLS certificate/key pair will be generated
func TestGetTLSConfig(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "tlsstuff-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempTLSCertPath := path.Join(tempDir, "tls.cert")
	tempTLSKeyPath := path.Join(tempDir, "tls.key")
	if utils.FileExists(tempTLSCertPath) || utils.FileExists(tempTLSKeyPath) {
		t.Fatalf("TLS certificate/key pair found when none were expected")
	}
	_, _, _, cleanUp, err := GetTLSConfig(tempTLSCertPath, tempTLSKeyPath, make([]string, 0))
	if err != nil {
		t.Fatalf("Could not getTLS config: %v", err)
	}
	if !utils.FileExists(tempTLSCertPath) || !utils.FileExists(tempTLSKeyPath) {
		t.Fatalf("TLS certificate/key pair files not found.")
	}
	cleanUp()
	_, _, _, cleanUp, err = GetTLSConfig(tempTLSCertPath, tempTLSKeyPath, make([]string, 0))
	if err != nil {
		t.Fatalf("Could not getTLS config: %v", err)
	}
	cleanUp()
}