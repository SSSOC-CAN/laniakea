/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var (
	testFileName      = "test.csv"
	expectedFileNames = map[int]string{
		1:  "test (1).csv",
		4:  "test (4).csv",
		11: "test (11).csv",
	}
)

// TestUniqueFileName tests the function UniqueFileName against a series of cases
func TestUniqueFileName(t *testing.T) {
	tmp_dir, err := ioutil.TempDir("", "utils-")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	file_name := filepath.Join(tmp_dir, testFileName)
	for i := 0; i < 12; i++ {
		new_file_name := UniqueFileName(file_name)
		f, err := os.Create(new_file_name)
		defer f.Close()
		if err != nil {
			t.Errorf("Could not create file at %v: %v", file_name, err)
		}
		if expected, ok := expectedFileNames[i]; ok {
			if new_file_name != filepath.Join(tmp_dir, expected) {
				t.Errorf("Expected: %s, recieved: %s", filepath.Join(tmp_dir, expected), new_file_name)
			}
		}
	}
}

// TestFileExists tests the FileExists function
func TestFileExists(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "utils-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	adminMacPath := path.Join(tempDir, "admin.macaroon")
	_, err = os.OpenFile(adminMacPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		t.Fatalf("Could not open/create file: %v", err)
	}
	testMacPath := path.Join(tempDir, "test.macaroon")
	if !FileExists(adminMacPath) {
		t.Fatal("File doesn't exist when it should")
	}
	if FileExists(testMacPath) {
		t.Fatal("File exists when it shouldn't")
	}
}

// TestNormalizeToNDecimalPlace tests the TestNormalizeToNDecimalPlace function
func TestNormalizeToNDecimalPlace(t *testing.T) {
	t.Run("test value greater than 1", func(t *testing.T) {
		_, err := NormalizeToNDecimalPlace(2.0)
		if err != ErrFloatLargerThanOne {
			t.Errorf("Unexpected error when calling NormalizeToNDecimalPlace: %v", err)
		}
	})
	t.Run("test value 1", func(t *testing.T) {
		_, err := NormalizeToNDecimalPlace(1.0)
		if err != ErrFloatLargerThanOne {
			t.Errorf("Unexpected error when calling NormalizeToNDecimalPlace: %v", err)
		}
	})
	t.Run("test value 0.9999", func(t *testing.T) {
		v, err := NormalizeToNDecimalPlace(0.9999)
		if err != nil {
			t.Errorf("Unexpected error when calling NormalizeToNDecimalPlace: %v", err)
		}
		if v != 0.1000 {
			t.Errorf("Unexpected result calling NormalizeToNDecimalPlace: %f", v)
		}
	})
	t.Run("test value -0.9999", func(t *testing.T) {
		v, err := NormalizeToNDecimalPlace(-0.9999)
		if err != nil {
			t.Errorf("Unexpected error when calling NormalizeToNDecimalPlace: %v", err)
		}
		if v != -0.1000 {
			t.Errorf("Unexpected result calling NormalizeToNDecimalPlace: %f", v)
		}
	})
}

// TestVerifyPluginStringFormat tests the VerifyPluginStringFormat function
func TestVerifyPluginStringFormat(t *testing.T) {
	cases := map[string]bool{
		"My-Plugin_124:datasource:plugin.exe":          true,
		"myplugin.test:datasource:plugin.exe":          true,
		"myplugin123:random:plugin.exe":                false,
		"myplugin_-123sdaASD:controller:plugin.exe":    true,
		"myplugin:datasource:pLgste_sdaA-1234.exe_":    false,
		"myplugin:datasource:pLgste_sdaA-1234.ex_e":    false,
		"myplugin:datasource:pLgste_sdaA-1234.ex-e":    false,
		"myplugin:datasource:pLgste_sdaA-1234.ex231e4": true,
	}
	for name, c := range cases {
		t.Run(fmt.Sprintf("test %v", name), func(t *testing.T) {
			if VerifyPluginStringFormat(name) != c {
				t.Errorf("Unexpected result: case %v\texpected %v", name, c)
			}
		})
	}
}
