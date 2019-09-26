// +build integration

package usecases_test

import (
	"fmt"
	"testing"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
	helpers "github.com/topfreegames/Will.IAM/testing"
	"github.com/topfreegames/Will.IAM/usecases"
)

func TestServiceAccountsCreate(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if saM.ID == "" {
		t.Errorf("Expected saM.ID to be non-empty")
	}
}

func TestServiceAccountsCreateShouldCreateRoleAndRoleBinding(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	rs, err := saUC.GetRoles(saM.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("Should have only 1 role binding. Found %d", len(rs))
	}
	rName := fmt.Sprintf("service-account:%s", saM.ID)
	if rs[0].Name != rName {
		t.Fatalf("Expected role name to be %s. Got %s", rName, rs[0].Name)
	}
}

type saHasPermissionTestCase struct {
	name                      string
	serviceAccountPermissions []string
	permission                string
	expected                  bool
}

var saHasPermissionTestCases = []saHasPermissionTestCase{
	// No permissions
	saHasPermissionTestCase{
		name:                      "No Permissions",
		serviceAccountPermissions: nil,
		permission:                "Service1::RL::Do1::x::*",
		expected:                  false,
	},
	// != Service
	saHasPermissionTestCase{
		name:                      "Different Service",
		serviceAccountPermissions: []string{"Service1::RL::Do1::x::*"},
		permission:                "Service2::RL::Do1::x::*",
		expected:                  false,
	},
	// Toying around with actions, * and multiple layers in RH
	saHasPermissionTestCase{
		name:                      "Same permission",
		serviceAccountPermissions: []string{"Service1::RL::Do1::x::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Different action same hierarchy multiple levels",
		serviceAccountPermissions: []string{"Service1::RL::Do2::x::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  false,
	},
	saHasPermissionTestCase{
		name:                      "Multiple permissions one equal match",
		serviceAccountPermissions: []string{"Service1::RL::Do2::x::*", "Service1::RL::Do1::x::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Multiple permissions one match with * in hierarchy",
		serviceAccountPermissions: []string{"Service1::RL::Do2::x::*", "Service1::RL::Do1::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Match action with * same resource hierarchy",
		serviceAccountPermissions: []string{"Service1::RL::*::x::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Match action with * match different hierarchy with *",
		serviceAccountPermissions: []string{"Service1::RL::*::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Match action with * match different hierarchy with *",
		serviceAccountPermissions: []string{"Service1::RL::*::*"},
		permission:                "Service1::RL::Do2::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Different action same hierarchy single level",
		serviceAccountPermissions: []string{"Service1::RL::Do1::*"},
		permission:                "Service1::RL::Do2::*",
		expected:                  false,
	},
	// Ownership levels
	saHasPermissionTestCase{
		name:                      "Different ownership level same hierarchy",
		serviceAccountPermissions: []string{"Service1::RL::Do1::x::*"},
		permission:                "Service1::RO::Do1::x::*",
		expected:                  false,
	},
	saHasPermissionTestCase{
		name:                      "Different ownership level different hierarchy matching with *",
		serviceAccountPermissions: []string{"Service1::RO::Do1::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Different ownership level same hierarchy",
		serviceAccountPermissions: []string{"Service1::RO::Do1::x::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  true,
	},
	saHasPermissionTestCase{
		name:                      "Different ownership level different hierarchy not match",
		serviceAccountPermissions: []string{"Service1::RO::Do1::y::*"},
		permission:                "Service1::RL::Do1::x::*",
		expected:                  false,
	},
}

func TestServiceAccountsHasPermissionWhenPermissionsOnBaseRole(t *testing.T) {
	for _, testCase := range saHasPermissionTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			helpers.CleanupPG(t)

			saUC := helpers.GetServiceAccountsUseCase(t)
			sa1Ps, err := models.BuildPermissions(testCase.serviceAccountPermissions)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			sa1 := &usecases.ServiceAccountWithNested{
				Name:               "sa1",
				Email:              "sa1@domain.com",
				Permissions:        sa1Ps,
				AuthenticationType: models.AuthenticationTypes.OAuth2,
			}
			if err := saUC.CreateWithNested(sa1); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			has, err := saUC.HasPermissionString(sa1.ID, testCase.permission)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if has != testCase.expected {
				t.Fatalf("Expected has to be %v. Got %v", testCase.expected, has)
			}
		})
	}
}

func TestServiceAccountsHasPermissionWhenPermissionsOnNonBaseRole(t *testing.T) {
	for _, testCase := range saHasPermissionTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			helpers.CleanupPG(t)

			saUC := helpers.GetServiceAccountsUseCase(t)
			sa1 := &models.ServiceAccount{
				Name:  "sa1",
				Email: "test@domain.com",
			}
			if err := saUC.Create(sa1); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			ps, err := models.BuildPermissions(testCase.serviceAccountPermissions)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			rl1 := &usecases.RoleWithNested{
				Name:               "role1",
				Permissions:        ps,
				ServiceAccountsIDs: []string{sa1.ID},
			}
			rsUC := helpers.GetRolesUseCase(t)
			if err := rsUC.Create(rl1); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			has, err := saUC.HasPermissionString(sa1.ID, testCase.permission)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if has != testCase.expected {
				t.Fatalf("Expected has to be %v. Got %v", testCase.expected, has)
			}
		})
	}
}

type saListWithPermissionTestCase struct {
	name       string
	sasPs      [][]string
	permission string
	expected   []string
}

var saListWithPermissionTestCases = []saListWithPermissionTestCase{
	saListWithPermissionTestCase{
		name: "Service Accounts equal match permission",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RL::Do1::x::z"},
		},
		permission: "Service1::RL::Do1::x::z",
		expected:   []string{"rootSAKeyPair", "sa0", "sa2"},
	},
	saListWithPermissionTestCase{
		name: "Service Accounts with equal match and hierarchy with *",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::Do1::x::z"},
		},
		permission: "Service1::RO::Do1::x::z",
		expected:   []string{"rootSAKeyPair", "sa2"},
	},
	saListWithPermissionTestCase{
		name: "Service Accounts with different service permission",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::Do1::x::z"},
		},
		permission: "Service2::RO::Do1::x::z",
		// Only rootSA matches because it has access to everything
		expected: []string{"rootSAKeyPair"},
	},
	saListWithPermissionTestCase{
		name: "Service Accounts with match in action with *",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::*::x::z"},
		},
		permission: "Service1::RO::Do1::x::z",
		expected:   []string{"rootSAKeyPair", "sa2"},
	},
}

func TestServiceAccountsListWithPermissionWhenPermissionOnBaseRole(t *testing.T) {
	for _, testCase := range saListWithPermissionTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			helpers.CleanupPG(t)

			saUC := helpers.GetServiceAccountsUseCase(t)
			root := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
			sas := []*usecases.ServiceAccountWithNested{}

			for j, psStr := range testCase.sasPs {
				ps, err := models.BuildPermissions(psStr)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				sa := &usecases.ServiceAccountWithNested{
					Name:               fmt.Sprintf("sa%d", j),
					Email:              fmt.Sprintf("sa%d@domain.com", j),
					Permissions:        ps,
					AuthenticationType: models.AuthenticationTypes.OAuth2,
				}
				if err := saUC.CreateWithNested(sa); err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				sas = append(sas, sa)
			}
			ps, err := models.BuildPermission(testCase.permission)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			list, count, err := saUC.ListWithPermission(root.ID, &repositories.ListOptions{}, ps)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if count != int64(len(testCase.expected)) {
				t.Fatalf("Expected to have %d service accounts. Got %d", len(testCase.expected), count)
			}
			if len(list) != len(testCase.expected) {
				t.Fatalf("Expected len(list) to be %d. Got %d", len(testCase.expected), len(list))
				t.Fatalf("List: %#v. Expected: %#v", list, testCase.expected)
			}
			for j := range list {
				if list[j].Name != testCase.expected[j] {
					t.Errorf("Expected list[%d] to be %s. Got %s", j, list[j], testCase.expected[j])
				}
			}
		})
	}
}

func TestServiceAccountsListWithPermissionWhenPermissionOnNonBaseRole(t *testing.T) {
	for _, testCase := range saListWithPermissionTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			helpers.CleanupPG(t)

			saUC := helpers.GetServiceAccountsUseCase(t)
			root := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
			sas := []*models.ServiceAccount{}

			for j, psStr := range testCase.sasPs {
				sa := &models.ServiceAccount{
					Name:  fmt.Sprintf("sa%d", j),
					Email: fmt.Sprintf("sa%d@domain.com", j),
				}
				if err := saUC.Create(sa); err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				ps, err := models.BuildPermissions(psStr)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				rl := &usecases.RoleWithNested{
					Name:               fmt.Sprintf("roleSa%d", j),
					Permissions:        ps,
					ServiceAccountsIDs: []string{sa.ID},
				}
				rsUC := helpers.GetRolesUseCase(t)
				if err := rsUC.Create(rl); err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				sas = append(sas, sa)
			}
			ps, err := models.BuildPermission(testCase.permission)
			if err != nil {
				t.Fatalf("Unexpected error: %v.", err)
			}
			list, count, err := saUC.ListWithPermission(root.ID, &repositories.ListOptions{}, ps)
			if err != nil {
				t.Fatalf("Unexpected error: %v.", err)
			}
			if count != int64(len(testCase.expected)) {
				t.Fatalf("Expected to have %d service accounts. Got %d.", len(testCase.expected), count)
			}
			if len(list) != len(testCase.expected) {
				t.Fatalf("Expected len(list) to be %d. Got %d.", len(testCase.expected), len(list))
				t.Fatalf("List: %#v. Expected: %#v.", list, testCase.expected)
			}
			for j := range list {
				if list[j].Name != testCase.expected[j] {
					t.Errorf("Expected list[%d] to be %s. Got %s.", j, list[j], testCase.expected[j])
				}
			}
		})
	}
}
