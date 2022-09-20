/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/20
*/

package intercept

import (
	"os"
	"testing"

	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/rs/zerolog"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	defaultTestingPermissions = map[string][]bakery.Op{
		"/lanirpc.Lani/StopDaemon": {{
			Entity: "laniakea",
			Action: "write",
		}},
		"/lanirpc.Lani/AdminTest": {{
			Entity: "laniakea",
			Action: "read",
		}},
	}
	duplicatePermissionMethod = "/lanirpc.Lani/StopDaemon"
	duplicatePermission       = []bakery.Op{
		{
			Entity: "uri",
			Action: "test",
		},
	}
	newPermissionMethod = "/lanirpc.Lani/TestCommand"
	newPermission       = []bakery.Op{
		{
			Entity: "laniakea",
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
		if err != errors.ErrDuplicateMacConstraints {
			t.Errorf("Unexpected error when adding permissions: %v", err)
		}
	})
	// invalid add permission
	t.Run("Invalid Add Permission", func(t *testing.T) {
		err := grpcInterceptor.AddPermission(duplicatePermissionMethod, duplicatePermission)
		if err != errors.ErrDuplicateMacConstraints {
			t.Errorf("Unexpected error when adding permissions: %v", err)
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
