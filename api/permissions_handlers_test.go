// +build integration

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/topfreegames/Will.IAM/models"
	helpers "github.com/topfreegames/Will.IAM/testing"
	"github.com/gofrs/uuid"
)

func beforeEachPermissionsHandlers(t *testing.T) {
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

func TestPermissionsDeleteHandlerNonExistentID(t *testing.T) {
	beforeEachRolesHandlers(t)
	rootSA := helpers.CreateRootServiceAccount(t)
	app := helpers.GetApp(t)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf(
		"/permissions/%s", uuid.Must(uuid.NewV4()).String(),
	), nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status 204. Got %d", rec.Code)
	}
}

func TestPermissionsDeleteHandlerNonRootSA(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	rootSA := helpers.CreateRootServiceAccount(t)
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
	deleterSA, err := saUC.CreateKeyPairType("deleter sa")
	req, _ = http.NewRequest("DELETE", fmt.Sprintf(
		"/permissions/%s", permissions[0].ID,
	), nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", deleterSA.KeyID, deleterSA.KeySecret,
	))
	rec = helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403. Got %d", rec.Code)
	}
}

func TestPermissionsDeleteHandler(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	rootSA := helpers.CreateRootServiceAccount(t)
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
	req, _ = http.NewRequest("DELETE", fmt.Sprintf(
		"/permissions/%s", permissions[0].ID,
	), nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec = helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200. Got %d", rec.Code)
	}
}

func TestPermissionsCreateRequestHandler(t *testing.T) {
	beforeEachRolesHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	sa, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	app := helpers.GetApp(t)
	req, _ := http.NewRequest("PUT", "/permissions/requests", strings.NewReader(`{
"service": "SomeService",
"action": "SomeAction",
"resourceHierarchy": "*",
"message": "hey, can I have this permission?"
	}`))
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusAccepted {
		t.Errorf("Expected status 202. Got %d", rec.Code)
	}
	req, _ = http.NewRequest("GET", "/permissions/requests", nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
	))
	rec = helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200. Got %d", rec.Code)
	}
	prs := []models.PermissionRequest{}
	err = json.Unmarshal([]byte(rec.Body.String()), &prs)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if len(prs) != 1 {
		t.Errorf("Expected to have 1 permission request. Got %d", len(prs))
		return
	}
	if prs[0].Service != "SomeService" {
		t.Errorf("Expected service to be SomeService. Got %s", prs[0].Service)
		return
	}
	if prs[0].Action.String() != "SomeAction" {
		t.Errorf("Expected action to be SomeAction. Got %s", prs[0].Action.String())
		return
	}
	msg := "hey, can I have this permission?"
	if prs[0].Message != msg {
		t.Errorf("Expected message to be '%s'. Got '%s'", msg, prs[0].Message)
		return
	}
	if prs[0].State != models.PermissionRequestStates.Created {
		t.Errorf("Expected state to be Created. Got %s", prs[0].State.String())
		return
	}
}

func TestPermissionsHasHandler(t *testing.T) {
	beforeEachPermissionsHandlers(t)
	type hasPermissionTest struct {
		request        string
		expectedStatus int
	}
	tt := []hasPermissionTest{
		hasPermissionTest{
			request:        "/permissions/has",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		hasPermissionTest{
			request:        "/permissions/has?permission=Service::RL::TestAction::*",
			expectedStatus: http.StatusForbidden,
		},
	}

	saUC := helpers.GetServiceAccountsUseCase(t)
	sa, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	app := helpers.GetApp(t)
	for _, tt := range tt {
		req, _ := http.NewRequest("GET", tt.request, nil)
		req.Header.Set("Authorization", fmt.Sprintf(
			"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
		))
		rec := helpers.DoRequest(t, req, app.GetRouter())
		if rec.Code != tt.expectedStatus {
			t.Errorf("Expected %d. Got %d", tt.expectedStatus, rec.Code)
		}
	}
}
