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
		t.Errorf("Unexpected error: %s", err.Error())
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
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rs, err := saUC.GetRoles(saM.ID)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if len(rs) != 1 {
		t.Errorf("Should have only 1 role binding. Found %d", len(rs))
		return
	}
	rName := fmt.Sprintf("service-account:%s", saM.ID)
	if rs[0].Name != rName {
		t.Errorf("Expected role name to be %s. Got %s", rName, rs[0].Name)
		return
	}
}

type saHasPermissionTestCase struct {
	name     string
	regPs    []string
	test     string
	expected bool
}

var saHasPermissionTestCases = []saHasPermissionTestCase{
	// No permissions
	saHasPermissionTestCase{
		name:     "No Permissions",
		regPs:    nil,
		test:     "Service1::RL::Do1::x::*",
		expected: false,
	},
	// != Service
	saHasPermissionTestCase{
		name:     "Different Service",
		regPs:    []string{"Service1::RL::Do1::x::*"},
		test:     "Service2::RL::Do1::x::*",
		expected: false,
	},
	// Toying around with actions, * and multiple layers in RH
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 1",
		regPs:    []string{"Service1::RL::Do1::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 2",
		regPs:    []string{"Service1::RL::Do2::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: false,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 3",
		regPs:    []string{"Service1::RL::Do2::x::*", "Service1::RL::Do1::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 4",
		regPs:    []string{"Service1::RL::Do2::x::*", "Service1::RL::Do1::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 5",
		regPs:    []string{"Service1::RL::*::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 6",
		regPs:    []string{"Service1::RL::*::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 7",
		regPs:    []string{"Service1::RL::*::*"},
		test:     "Service1::RL::Do2::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Wildcard operator Scenario 8",
		regPs:    []string{"Service1::RL::Do1::*"},
		test:     "Service1::RL::Do2::*",
		expected: false,
	},
	// Ownership levels
	saHasPermissionTestCase{
		name:     "Different Ownership Level Scenario 1",
		regPs:    []string{"Service1::RL::Do1::x::*"},
		test:     "Service1::RO::Do1::x::*",
		expected: false,
	},
	saHasPermissionTestCase{
		name:     "Different Ownership Level Scenario 2",
		regPs:    []string{"Service1::RO::Do1::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Different Ownership Level Scenario 3",
		regPs:    []string{"Service1::RO::Do1::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		name:     "Different Ownership Level Scenario 4",
		regPs:    []string{"Service1::RO::Do1::y::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: false,
	},
}

func TestServiceAccountsHasPermissionWhenPermissionsOnBaseRole(t *testing.T) {
	for _, testCase := range saHasPermissionTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			helpers.CleanupPG(t)

			saUC := helpers.GetServiceAccountsUseCase(t)
			sa1Ps, err := models.BuildPermissions(testCase.regPs)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			sa1 := &usecases.ServiceAccountWithNested{
				Name:               "sa1",
				Email:              "sa1@domain.com",
				Permissions:        sa1Ps,
				AuthenticationType: models.AuthenticationTypes.OAuth2,
			}
			if err := saUC.CreateWithNested(sa1); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			has, err := saUC.HasPermissionString(sa1.ID, testCase.test)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			if has != testCase.expected {
				t.Errorf("Expected has to be %v. Got %v", testCase.expected, has)
				return
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
				t.Errorf("Unexpected error: %s", err.Error())
			}
			ps, err := models.BuildPermissions(testCase.regPs)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			rl1 := &usecases.RoleWithNested{
				Name:               "role1",
				Permissions:        ps,
				ServiceAccountsIDs: []string{sa1.ID},
			}
			rsUC := helpers.GetRolesUseCase(t)
			if err := rsUC.Create(rl1); err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
			}
			has, err := saUC.HasPermissionString(sa1.ID, testCase.test)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			if has != testCase.expected {
				t.Errorf("Expected has to be %v. Got %v", testCase.expected, has)
				return
			}
		})
	}
}

type saListWithPermissionTestCase struct {
	name     string
	sasPs    [][]string
	test     string
	expected []string
}

var saListWithPermissionTestCases = []saListWithPermissionTestCase{
	saListWithPermissionTestCase{
		name: "Scenario 1",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RL::Do1::x::z"},
		},
		test:     "Service1::RL::Do1::x::z",
		expected: []string{"sa0", "sa2", "user"},
	},
	saListWithPermissionTestCase{
		name: "Scenario 2",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::Do1::x::z"},
		},
		test:     "Service1::RO::Do1::x::z",
		expected: []string{"sa2", "user"},
	},
	saListWithPermissionTestCase{
		name: "Scenario 3",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::Do1::x::z"},
		},
		test:     "Service2::RO::Do1::x::z",
		expected: []string{"user"},
	},
	saListWithPermissionTestCase{
		name: "Scenario 4",
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::*::x::z"},
		},
		test:     "Service1::RO::Do1::x::z",
		expected: []string{"sa2", "user"},
	},
}

func TestServiceAccountsListWithPermissionWhenPermissionOnBaseRole(t *testing.T) {
	for _, testCase := range saListWithPermissionTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			helpers.CleanupPG(t)

			saUC := helpers.GetServiceAccountsUseCase(t)
			root := helpers.CreateRootServiceAccountWithKeyPair(t, "user", "user@test.com")
			sas := []*usecases.ServiceAccountWithNested{}
			for j, psStr := range testCase.sasPs {
				ps, err := models.BuildPermissions(psStr)
				if err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				sa := &usecases.ServiceAccountWithNested{
					Name:               fmt.Sprintf("sa%d", j),
					Email:              fmt.Sprintf("sa%d@domain.com", j),
					Permissions:        ps,
					AuthenticationType: models.AuthenticationTypes.OAuth2,
				}
				if err := saUC.CreateWithNested(sa); err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				sas = append(sas, sa)
			}
			ps, err := models.BuildPermission(testCase.test)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			list, count, err := saUC.ListWithPermission(root.ID, &repositories.ListOptions{}, ps)
			if err != nil {
				t.Errorf("Unexpected error: %s", err.Error())
				return
			}
			if count != int64(len(testCase.expected)) {
				t.Errorf("Expected to have %d service accounts. Got %d", len(testCase.expected), count)
				return
			}
			if len(list) != len(testCase.expected) {
				t.Errorf("Expected len(list) to be %d. Got %d", len(testCase.expected), len(list))
				t.Errorf("List: %#v. Expected: %#v", list, testCase.expected)
				return
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
			root := helpers.CreateRootServiceAccountWithKeyPair(t, "user", "user@test.com")
			sas := []*models.ServiceAccount{}

			for j, psStr := range testCase.sasPs {
				sa := &models.ServiceAccount{
					Name:  fmt.Sprintf("sa%d", j),
					Email: fmt.Sprintf("sa%d@domain.com", j),
				}
				if err := saUC.Create(sa); err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
				}
				ps, err := models.BuildPermissions(psStr)
				if err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				rl := &usecases.RoleWithNested{
					Name:               fmt.Sprintf("roleSa%d", j),
					Permissions:        ps,
					ServiceAccountsIDs: []string{sa.ID},
				}
				rsUC := helpers.GetRolesUseCase(t)
				if err := rsUC.Create(rl); err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
				}
				sas = append(sas, sa)
			}
			ps, err := models.BuildPermission(testCase.test)
			if err != nil {
				t.Errorf("Unexpected error: %s.", err.Error())
				return
			}
			list, count, err := saUC.ListWithPermission(root.ID, &repositories.ListOptions{}, ps)
			if err != nil {
				t.Errorf("Unexpected error: %s.", err.Error())
				return
			}
			if count != int64(len(testCase.expected)) {
				t.Errorf("Expected to have %d service accounts. Got %d.", len(testCase.expected), count)
				return
			}
			if len(list) != len(testCase.expected) {
				t.Errorf("Expected len(list) to be %d. Got %d.", len(testCase.expected), len(list))
				t.Errorf("List: %#v. Expected: %#v.", list, testCase.expected)
				return
			}
			for j := range list {
				if list[j].Name != testCase.expected[j] {
					t.Errorf("Expected list[%d] to be %s. Got %s.", j, list[j], testCase.expected[j])
				}
			}
		})
	}
}
