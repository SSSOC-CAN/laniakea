package macaroons

import (
	"fmt"
	"strings"
	"testing"
	"time"
	macaroon "gopkg.in/macaroon.v2"
)

var (
	testRootKey = []byte("dummyRootKey")
	testID = []byte("dummyId")
	testLocation = "fmtd"
	testVersion = macaroon.LatestVersion
	expectedTimeCaveatSubstring = fmt.Sprintf("time-before %d", time.Now().Year())
)

// createDummyMacaroon creates a dummy macaroon with the settings in the above global variables
func createDummyMacaroon(t *testing.T) *macaroon.Macaroon {
	dummyMacaroon, err := macaroon.New(
		testRootKey, testID, testLocation, testVersion,
	)
	if err != nil {
		t.Fatalf("Error creating initial macaroon: %v", err)
	}
	return dummyMacaroon
}

// TestAddConstraints adds a timeout constraint to a dummy mac and tests whether this constraint is properly applied
func TestAddConstraints(t *testing.T) {
	mac := createDummyMacaroon(t)
	moddedMac, err := AddConstraints(mac, TimeoutConstraint(1))
	if err != nil {
		t.Fatalf("Error adding timeout constraint: %v", err)
	}
	if &moddedMac == &mac {
		t.Fatalf("Old and new macaroon are identical, expected differences")
	}
	if len(mac.Caveats()) == len(moddedMac.Caveats()) {
		t.Fatalf("Number of caveats for the old and new macaroon are the same, expected a difference")
	}
}

// TestTimeoutConstraint tests that a caveat for the lifetime of a macaroon is created.
func TestTimeoutConstraint(t *testing.T) {
	timeoutFunction := TimeoutConstraint(3)
	mac := createDummyMacaroon(t)
	err := timeoutFunction(mac)
	if err != nil {
		t.Fatalf("Error applying timeout constraint to dummy macaroon: %v", err)
	}
	if !strings.HasPrefix(string(mac.Caveats()[0].Id), expectedTimeCaveatSubstring) {
		t.Fatalf("Added caveat '%s' does not meet expectations", mac.Caveats()[0].Id)
	}
}