// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ghostec/Will.IAM/models"
	helpers "github.com/ghostec/Will.IAM/testing"
)

func beforeEachRolesHandlers(t *testing.T) {
	t.Helper()
	storage := helpers.GetStorage(t)
	rels := []string{"permissions", "role_bindings", "service_accounts", "roles"}
	for _, rel := range rels {
		if _, err := storage.PG.DB.Exec(
			fmt.Sprintf("DELETE FROM %s", rel),
		); err != nil {
			panic(err)
		}
	}
}

func TestRolesCreatePermissionHandler(t *testing.T) {
	beforeEachRolesHandlers(t)
	rootSA := helpers.CreateRootServiceAccount(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	sa, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	app := helpers.GetApp(t)
	p := "SomeService::RO::SomeAction::*"
	req, _ := http.NewRequest("POST", fmt.Sprintf(
		"/roles/%s/permissions?permission=%s", sa.BaseRoleID, p,
	), nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201. Got %d", rec.Code)
	}
	permissions, err := saUC.GetPermissions(sa.ID)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if len(permissions) != 1 {
		t.Errorf("Expected only 1 permission. Got %d", len(permissions))
		return
	}
	if permissions[0].String() != p {
		t.Errorf(
			"Expected permission to be %s. Got %s", p, permissions[0].String(),
		)
		return
	}
}

func TestRolesCreatePermissionHandlerNonRootSA(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	creatorSA, err := saUC.CreateKeyPairType("creator sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	sa, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	app := helpers.GetApp(t)
	p := "SomeService::RO::SomeAction::*"
	req, _ := http.NewRequest("POST", fmt.Sprintf(
		"/roles/%s/permissions?permission=%s", sa.BaseRoleID, p,
	), nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", creatorSA.KeyID, creatorSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403. Got %d", rec.Code)
	}
}

func TestRolesUpdateHandlerRootSA(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	rootSA := helpers.CreateRootServiceAccount(t)
	targetSA, err := saUC.CreateKeyPairType("creator sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	app := helpers.GetApp(t)

	body := map[string]interface{}{
		"name": "new role name",
		"permissions": []string{
			"SomeService::RO::SomeAction::*",
		},
	}
	bts, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	req, _ := http.NewRequest(
		"PUT",
		fmt.Sprintf("/roles/%s", targetSA.BaseRoleID),
		bytes.NewBuffer(bts),
	)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200. Got %d", rec.Code)
	}

	rsUC := helpers.GetRolesUseCase(t)
	r, err := rsUC.Get(targetSA.BaseRoleID)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if r["name"] != "new role name" {
		t.Errorf("Expected role name to be 'new role name'. Got %s", r["name"])
	}

	pSl, err := rsUC.GetPermissions(targetSA.BaseRoleID)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if len(pSl) != 1 {
		t.Errorf("Expected to have 1 permission. Got %d", len(pSl))
		return
	}
	if pSl[0].Service != "SomeService" {
		t.Errorf("Expected service to be SomeService. Got %s", pSl[0].Service)
		return
	}
	if pSl[0].OwnershipLevel != models.OwnershipLevels.Owner {
		t.Errorf(
			"Expected ownership level to be Owner. Got %s",
			pSl[0].OwnershipLevel.String(),
		)
		return
	}
	if pSl[0].Action != "SomeAction" {
		t.Errorf("Expected action to be SomeAction. Got %s", pSl[0].Action)
		return
	}
	if pSl[0].ResourceHierarchy.String() != "*" {
		t.Errorf("Expected resource hierarchy to be *. Got %s", pSl[0].ResourceHierarchy)
		return
	}
}

func TestRolesUpdateHandlerNonRootSA(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	creatorSA, err := saUC.CreateKeyPairType("creator sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	targetSA, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	app := helpers.GetApp(t)

	body := map[string]interface{}{
		"name": "new role name",
		"permissions": []string{
			"SomeService::RO::SomeAction::*",
		},
	}
	bts, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	req, _ := http.NewRequest(
		"PUT",
		fmt.Sprintf("/roles/%s", targetSA.BaseRoleID),
		bytes.NewBuffer(bts),
	)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", creatorSA.KeyID, creatorSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403. Got %d", rec.Code)
	}

	rsUC := helpers.GetRolesUseCase(t)
	r, err := rsUC.Get(creatorSA.BaseRoleID)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if r["name"] == "new role name" {
		t.Error("Expected role name to be != 'new role name'")
	}

	pSl, err := rsUC.GetPermissions(creatorSA.BaseRoleID)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if len(pSl) > 0 {
		t.Errorf("Expected to have 0 permissions. Got %d", len(pSl))
		return
	}
}

func TestRolesGetHandlerNonRootSA(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	nonRootSA, err := saUC.CreateKeyPairType("creator sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	targetSA, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	saUC.CreatePermission(nonRootSA.ID, &models.Permission{
		Service:           "Will.IAM",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "EditRole",
		ResourceHierarchy: models.ResourceHierarchy(targetSA.BaseRoleID),
	})
	app := helpers.GetApp(t)

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("/roles/%s", targetSA.BaseRoleID),
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", nonRootSA.KeyID, nonRootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200. Got %d", rec.Code)
	}
}
