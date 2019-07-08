// +build unit

package models_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/topfreegames/Will.IAM/models"
)

func TestValidatePermission(t *testing.T) {
	type testCase struct {
		str   string
		valid bool
		err   error
	}
	tt := []testCase{
		testCase{
			str:   "Maestro::RX::ListSchedulers::*",
			valid: false,
			err:   fmt.Errorf("OwnershipLevel needs to be RO or RL"),
		},
		testCase{
			str:   "Maestro::RO::ListSchedulers::*",
			valid: true,
			err:   nil,
		},
		testCase{
			str:   "Maestro::RL::ListSchedulers::*",
			valid: true,
			err:   nil,
		},
		testCase{
			str:   "Maestro::RO::ListSchedulers::",
			valid: false,
			err:   fmt.Errorf("No parts can be empty"),
		},
		testCase{
			str:   "Maestro::RL::ListSchedulers",
			valid: false,
			err: fmt.Errorf(
				"Incomplete permission. Expected format: " +
					"Service::OwnershipLevel::Action::{ResourceHierarchy}",
			),
		},
	}

	for _, tt := range tt {
		valid, err := models.ValidatePermission(tt.str)
		if valid != tt.valid {
			t.Errorf("Expected valid to be %t. Got %t", valid, tt.valid)
			if err != nil {
				t.Errorf(err.Error())
			}
		}
		if err != nil && err.Error() != tt.err.Error() {
			t.Errorf("Expected error to be %s. Got %s", err.Error(), tt.err.Error())
		}
	}
}

func TestResourceHierarchyPermissionMatches(t *testing.T) {
	type testCase struct {
		str      string
		expected []string
	}
	tt := []testCase{
		testCase{
			str:      "*",
			expected: []string{"*"},
		},
		testCase{
			str:      "x::*",
			expected: []string{"*", "x::*"},
		},
		testCase{
			str:      "x::y",
			expected: []string{"*", "x::*", "x::y"},
		},
		testCase{
			str:      "x::y::*",
			expected: []string{"*", "x::*", "x::y::*"},
		},
		testCase{
			str:      "x::y::z",
			expected: []string{"*", "x::*", "x::y::*", "x::y::z"},
		},
	}
	for _, tt := range tt {
		rh := models.BuildResourceHierarchy(tt.str)
		matches := rh.PermissionMatches()
		if len(matches) != len(tt.expected) {
			t.Errorf("Expected matches to have len %d. Got %d", len(tt.expected), len(matches))
			t.Errorf("Expected matches: %#v. Got %#v", tt.expected, matches)
			return
		}
		for i := range matches {
			if matches[i] != tt.expected[i] {
				t.Errorf("Expected matches[%d] to be %s. Got %s", i, tt.expected[i], matches[i])
				return
			}
		}
	}
}

func TestBuildPermission(t *testing.T) {
	type testCase struct {
		str        string
		permission models.Permission
		err        error
	}
	tt := []testCase{
		testCase{
			str: "Maestro::RO::ListSchedulers::*",
			permission: models.Permission{
				OwnershipLevel:    models.OwnershipLevels.Owner,
				Action:            models.BuildAction("ListSchedulers"),
				Service:           "Maestro",
				ResourceHierarchy: models.ResourceHierarchy("*"),
			},
			err: nil,
		},
	}

	for _, tt := range tt {
		permission, _ := models.BuildPermission(tt.str)
		if equal := reflect.DeepEqual(permission, tt.permission); !equal {
			t.Errorf(
				"Expected permission to be %#v. Got %#v", tt.permission, permission,
			)
		}
	}
}

func buildPermissions(strSl []string) []models.Permission {
	permissionsSl := make([]models.Permission, len(strSl))
	for i, str := range strSl {
		permissionsSl[i], _ = models.BuildPermission(str)
	}
	return permissionsSl
}

func TestHasPermission(t *testing.T) {
	type testCase struct {
		permission  string
		permissions []models.Permission
		isPresent   bool
	}

	sniperPermissions := []string{
		"Maestro::RO::ListSchedulers::Sniper3D::*",
		"Maestro::RO::ListSchedulers::Sniper3D::sniper3d-game",
	}

	tt := []testCase{
		testCase{
			permission:  "Maestro::RO::ListSchedulers::Sniper3D::sniper3d-game",
			permissions: buildPermissions(sniperPermissions),
			isPresent:   true,
		},
		testCase{
			permission:  "Maestro::RO::ListSchedulers::WarMachines::*",
			permissions: buildPermissions(sniperPermissions),
			isPresent:   false,
		},
	}

	for i, tt := range tt {
		permission, err := models.BuildPermission(tt.permission)
		if err != nil {
			t.Errorf(
				"Expected error not to have happened. Str: %s. Error: %s",
				tt.permission, err.Error(),
			)
		}
		isPresent := permission.IsPresent(tt.permissions)
		if isPresent != tt.isPresent {
			t.Errorf(
				"Expected IsPresent to be %t. Got: %t. Case #%d",
				tt.isPresent, isPresent, i,
			)
		}
	}
}
