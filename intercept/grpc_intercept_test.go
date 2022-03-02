package intercept

import (
	"os"
	"testing"
	"github.com/rs/zerolog"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	defaultTestingPermissions = map[string][]bakery.Op{
		"/fmtrpc.Fmt/StopDaemon": {{
			Entity: "fmtd",
			Action:	"write",
		}},
		"/fmtrpc.Fmt/AdminTest": {{
			Entity: "fmtd",
			Action: "read",
		}},
	}
	duplicatePermissionMethod = "/fmtrpc.Fmt/StopDaemon"
	duplicatePermission = []bakery.Op{
		{
			Entity: "uri",
			Action: "test",
		},
	}
	newPermissionMethod = "/fmtrpc.Fmt/TestCommand"
	newPermission = []bakery.Op{
		{
			Entity: "fmtd",
			Action: "read",
		},
	}
)

// TestAddPermissions tests a few cases for the AddPermissions method
func TestAddPermissions(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	grpcInterceptor := NewGrpcInterceptor(&logger, true)
	// valid add permissions
	t.Run("Valid Add Permissions", func(t *testing.T) {
		err := grpcInterceptor.AddPermissions(defaultTestingPermissions)
		if err != nil {
			t.Errorf("Could not add permissions: %v", err)
		}
	})
	// invalid add permissions
	t.Run("Invalid Add Permissions", func(t *testing.T) {
		err := grpcInterceptor.AddPermissions(defaultTestingPermissions)
		if err == nil {
			t.Errorf("Expected an error and none is present")
		}
	})
	// invalid add permission
	t.Run("Invalid Add Permission", func(t *testing.T) {
		err := grpcInterceptor.AddPermission(duplicatePermissionMethod, duplicatePermission)
		if err == nil {
			t.Errorf("Expected an error and none is present")
		}
	})
	// valid add permission
	t.Run("Valid Add Permission", func(t *testing.T) {
		err := grpcInterceptor.AddPermission(newPermissionMethod, newPermission)
		if err != nil {
			t.Errorf("Could not add permission: %v", err)
		}
	})
}