package main 

import (
	//"fmt"
	//"io/ioutil"
	//"os"
	"os/exec"
	//"path"
	"testing"
	//"github.com/SSSOC-CAN/fmtd/utils"
)

var (
	app = "fmtcli"
	listOfGoodTestCommandCases = [][]string{
		[]string{
			"test",
		},
		[]string{
			"admin-test",
		},
		[]string{
			"start-record",
			"fluke",
		},
		[]string{
			"stop-record",
			"fluke",
		},
	}
	bakeMacaroonFlags = []string{
		`bake-macaroon`,
		`"%v"`,
		"--timeout=60",
		`--save_to="%v"`,
	}
	badTestCommandCase = []string{
		"start-record",
		"rga",
	}
)

// TestGoodFMTCLICommands tests a list of commands where the expectation is that they won't fail
// FMTD must be running and must be logged in for success
func TestGoodFMTCLICommands(t *testing.T) {
	for _, args := range listOfGoodTestCommandCases {
		cmd := exec.Command(app, args...)
		_, err := cmd.Output()
		if err != nil {
			t.Fatalf("Could not execute fmtcli command: %v", err)
		}
	}
}

// TestBadFMTCLICommand tests whether a purposeful bad input will fail
func TestBadFMTCLICommand(t *testing.T) {
	cmd := exec.Command(app, badTestCommandCase...)
	_, err := cmd.Output()
	if err == nil {
		t.Fatalf("fmtcli command successfully executed when expected to fail: %v", badTestCommandCase)
	}
}

// TODO:SSSOCPaulCote - figure out why this isn't working
// // TestBakeMacaroonCommand tests if the bake macaroon command works
// func TestBakeMacaroonCommand(t *testing.T) {
// 	tempDir, err := ioutil.TempDir("", "fmtcli-")
// 	if err != nil {
// 		t.Fatalf("Error creating temporary directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)
// 	tempMacPath := path.Join(tempDir, "test.macaroon")
// 	cmd := exec.Command(
// 		app, 
// 		bakeMacaroonFlags[0], 
// 		fmt.Sprintf(bakeMacaroonFlags[1], "uri:/fmtrpc.Fmt/AdminTest"),
// 		bakeMacaroonFlags[2],
// 		fmt.Sprintf(bakeMacaroonFlags[3], tempMacPath),
// 	)
// 	stdout, err := cmd.Output()
// 	if err != nil {
// 		t.Fatalf("Could not execute fmtcli command: %v\n%v", err, stdout)
// 	}
// 	if !utils.FileExists(tempMacPath) {
// 		t.Fatalf("Could not bake new macaroon at %v", tempMacPath)
// 	}
// }