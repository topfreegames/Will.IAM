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

func beforeEachServiceAccounts(t *testing.T) {
	t.Helper()
	storage := helpers.GetStorage(t)
	rels := []string{"permissions", "role_bindings", "service_accounts", "roles"}
	for _, rel := range rels {
		if _, err := storage.PG.DB.Exec(
			fmt.Sprintf("DELETE FROM %s;", rel),
		); err != nil {
			panic(err)
		}
	}
}

func TestServiceAccountsCreate(t *testing.T) {
	beforeEachServiceAccounts(t)
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
	beforeEachServiceAccounts(t)
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
	regPs    []string
	test     string
	expected bool
}

var saHasPermissionTestCases = []saHasPermissionTestCase{
	// No permissions
	saHasPermissionTestCase{
		regPs:    nil,
		test:     "Service1::RL::Do1::x::*",
		expected: false,
	},
	// != Service
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do1::x::*"},
		test:     "Service2::RL::Do1::x::*",
		expected: false,
	},
	// Toying around with actions, * and multiple layers in RH
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do1::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do2::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: false,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do2::x::*", "Service1::RL::Do1::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do2::x::*", "Service1::RL::Do1::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::*::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::*::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::*::*"},
		test:     "Service1::RL::Do2::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do1::*"},
		test:     "Service1::RL::Do2::*",
		expected: false,
	},
	// Ownership levels
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RL::Do1::x::*"},
		test:     "Service1::RO::Do1::x::*",
		expected: false,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RO::Do1::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RO::Do1::x::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: true,
	},
	saHasPermissionTestCase{
		regPs:    []string{"Service1::RO::Do1::y::*"},
		test:     "Service1::RL::Do1::x::*",
		expected: false,
	},
}

func TestServiceAccountsHasPermissionWhenPermissionsOnBaseRole(t *testing.T) {
	for i, tt := range saHasPermissionTestCases {
		beforeEachServiceAccounts(t)
		saUC := helpers.GetServiceAccountsUseCase(t)
		sa1Ps, err := models.BuildPermissions(tt.regPs)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		sa1 := &usecases.ServiceAccountWithNested{
			Name:               "sa1",
			Email:              "sa1@domain.com",
			Permissions:        sa1Ps,
			AuthenticationType: models.AuthenticationTypes.OAuth2,
		}
		if err := saUC.CreateWithNested(sa1); err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		has, err := saUC.HasPermissionString(sa1.ID, tt.test)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		if has != tt.expected {
			t.Errorf("Expected has to be %v. Got %v. Case: %d", tt.expected, has, i)
			return
		}
	}
}

func TestServiceAccountsHasPermissionWhenPermissionsOnNonBaseRole(t *testing.T) {
	for i, tt := range saHasPermissionTestCases {
		beforeEachServiceAccounts(t)
		saUC := helpers.GetServiceAccountsUseCase(t)
		sa1 := &models.ServiceAccount{
			Name:  "sa1",
			Email: "test@domain.com",
		}
		if err := saUC.Create(sa1); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		ps, err := models.BuildPermissions(tt.regPs)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
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
		has, err := saUC.HasPermissionString(sa1.ID, tt.test)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		if has != tt.expected {
			t.Errorf("Expected has to be %v. Got %v. Case: %d", tt.expected, has, i)
			return
		}
	}
}

type saListWithPermissionTestCase struct {
	sasPs    [][]string
	test     string
	expected []string
}

var saListWithPermissionTestCases = []saListWithPermissionTestCase{
	saListWithPermissionTestCase{
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RL::Do1::x::z"},
		},
		test:     "Service1::RL::Do1::x::z",
		expected: []string{"sa0", "sa2"},
	},
	saListWithPermissionTestCase{
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::Do1::x::z"},
		},
		test:     "Service1::RO::Do1::x::z",
		expected: []string{"sa2"},
	},
	saListWithPermissionTestCase{
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::Do1::x::z"},
		},
		test:     "Service2::RO::Do1::x::z",
		expected: []string{},
	},
	saListWithPermissionTestCase{
		sasPs: [][]string{
			[]string{"Service1::RL::Do1::x::*"},
			[]string{"Service1::RL::Do1::x::y"},
			[]string{"Service1::RO::*::x::z"},
		},
		test:     "Service1::RO::Do1::x::z",
		expected: []string{"sa2"},
	},
}

func TestServiceAccountsListWithPermissionWhenPermissionOnBaseRole(t *testing.T) {
	for i, tt := range saListWithPermissionTestCases {
		beforeEachServiceAccounts(t)
		saUC := helpers.GetServiceAccountsUseCase(t)
		sas := []*usecases.ServiceAccountWithNested{}
		for j, psStr := range tt.sasPs {
			ps, err := models.BuildPermissions(psStr)
			if err != nil {
				t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
				return
			}
			sa := &usecases.ServiceAccountWithNested{
				Name:               fmt.Sprintf("sa%d", j),
				Email:              fmt.Sprintf("sa%d@domain.com", j),
				Permissions:        ps,
				AuthenticationType: models.AuthenticationTypes.OAuth2,
			}
			if err := saUC.CreateWithNested(sa); err != nil {
				t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
				return
			}
			sas = append(sas, sa)
		}
		ps, err := models.BuildPermission(tt.test)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		list, count, err := saUC.ListWithPermission(&repositories.ListOptions{}, ps)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		if count != int64(len(tt.expected)) {
			t.Errorf("Expected to have %d service accounts. Got %d. Case: %d", len(tt.sasPs), count, i)
			return
		}
		if len(list) != len(tt.expected) {
			t.Errorf("Expected len(list) to be %d. Got %d. Case: %d", len(tt.expected), len(list), i)
			t.Errorf("List: %#v. Expected: %#v. Case: %d", list, tt.expected, i)
			return
		}
		for j := range list {
			if list[j].Name != tt.expected[j] {
				t.Errorf("Expected list[%d] to be %s. Got %s. Case: %d", j, list[j], tt.expected[j], i)
			}
		}
	}
}

func TestServiceAccountsListWithPermissionWhenPermissionOnNonBaseRole(t *testing.T) {
	for i, tt := range saListWithPermissionTestCases {
		beforeEachServiceAccounts(t)
		saUC := helpers.GetServiceAccountsUseCase(t)
		sas := []*models.ServiceAccount{}
		for j, psStr := range tt.sasPs {
			sa := &models.ServiceAccount{
				Name:  fmt.Sprintf("sa%d", j),
				Email: fmt.Sprintf("sa%d@domain.com", j),
			}
			if err := saUC.Create(sa); err != nil {
				t.Errorf("Unexpected error: %s. Case %d", err.Error(), i)
			}
			ps, err := models.BuildPermissions(psStr)
			if err != nil {
				t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
				return
			}
			rl := &usecases.RoleWithNested{
				Name:               fmt.Sprintf("roleSa%d", j),
				Permissions:        ps,
				ServiceAccountsIDs: []string{sa.ID},
			}
			rsUC := helpers.GetRolesUseCase(t)
			if err := rsUC.Create(rl); err != nil {
				t.Errorf("Unexpected error: %s. Case %d", err.Error(), i)
			}
			sas = append(sas, sa)
		}
		ps, err := models.BuildPermission(tt.test)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		list, count, err := saUC.ListWithPermission(&repositories.ListOptions{}, ps)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		if count != int64(len(tt.expected)) {
			t.Errorf("Expected to have %d service accounts. Got %d. Case: %d", len(tt.sasPs), count, i)
			return
		}
		if len(list) != len(tt.expected) {
			t.Errorf("Expected len(list) to be %d. Got %d. Case: %d", len(tt.expected), len(list), i)
			t.Errorf("List: %#v. Expected: %#v. Case: %d", list, tt.expected, i)
			return
		}
		for j := range list {
			if list[j].Name != tt.expected[j] {
				t.Errorf("Expected list[%d] to be %s. Got %s. Case: %d", j, list[j], tt.expected[j], i)
			}
		}
	}
}
