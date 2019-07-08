// +build integration

package usecases_test

import (
	"fmt"
	"testing"

	"github.com/topfreegames/Will.IAM/models"
	helpers "github.com/topfreegames/Will.IAM/testing"
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
