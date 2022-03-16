package utils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var (
	testFileName = "test.csv"
	expectedFileNames = map[int]string {
		1: "test (1).csv",
		4: "test (4).csv",
		11: "test (11).csv",
	}
)
// TestUniqueFileName tests the function UniqueFileName against a series of cases
func TestUniqueFileName(t *testing.T) {
	tmp_dir, err := ioutil.TempDir(AppDataDir("fmtd", false), "fluke_test")
	if err != nil {
		t.Errorf("Could not create a temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp_dir)
	file_name := filepath.Join(tmp_dir, testFileName)
	for i := 0; i < 12; i++{
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