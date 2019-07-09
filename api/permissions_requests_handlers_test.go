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
)

func beforeEachPermissionsRequestsHandlers(t *testing.T) {
	t.Helper()
	storage := helpers.GetStorage(t)
	rels := []string{"permissions", "role_bindings", "service_accounts", "roles"}
	if _, err := storage.PG.DB.Exec(
		fmt.Sprintf("TRUNCATE %s", strings.Join(rels, ", ")),
	); err != nil {
		panic(err)
	}
}

func TestPermissionsRequestsListHandler(t *testing.T) {
	beforeEachPermissionsRequestsHandlers(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	sa, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
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

// func TestPermissionsCreateRequestHandler(t *testing.T) {
// 	beforeEachRolesHandlers(t)
// 	saUC := helpers.GetServiceAccountsUseCase(t)
// 	sa, err := saUC.CreateKeyPairType("some sa")
// 	if err != nil {
// 		t.Errorf("Unexpected error %s", err.Error())
// 		return
// 	}
// 	app := helpers.GetApp(t)
// 	req, _ := http.NewRequest("PUT", "/permissions/requests", strings.NewReader(`{
// "service": "SomeService",
// "action": "SomeAction",
// "resourceHierarchy": "*",
// "message": "hey, can I have this permission?"
// 	}`))
// 	req.Header.Set("Authorization", fmt.Sprintf(
// 		"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
// 	))
// 	rec := helpers.DoRequest(t, req, app.GetRouter())
// 	if rec.Code != http.StatusAccepted {
// 		t.Errorf("Expected status 202. Got %d", rec.Code)
// 	}
// 	req, _ = http.NewRequest("GET", "/permissions/requests", nil)
// 	req.Header.Set("Authorization", fmt.Sprintf(
// 		"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
// 	))
// 	rec = helpers.DoRequest(t, req, app.GetRouter())
// 	if rec.Code != http.StatusOK {
// 		t.Errorf("Expected status 200. Got %d", rec.Code)
// 	}
// 	prs := []models.PermissionRequest{}
// 	err = json.Unmarshal([]byte(rec.Body.String()), &prs)
// 	if err != nil {
// 		t.Errorf("Unexpected error %s", err.Error())
// 		return
// 	}
// 	if len(prs) != 1 {
// 		t.Errorf("Expected to have 1 permission request. Got %d", len(prs))
// 		return
// 	}
// 	if prs[0].Service != "SomeService" {
// 		t.Errorf("Expected service to be SomeService. Got %s", prs[0].Service)
// 		return
// 	}
// 	if prs[0].Action.String() != "SomeAction" {
// 		t.Errorf("Expected action to be SomeAction. Got %s", prs[0].Action.String())
// 		return
// 	}
// 	msg := "hey, can I have this permission?"
// 	if prs[0].Message != msg {
// 		t.Errorf("Expected message to be '%s'. Got '%s'", msg, prs[0].Message)
// 		return
// 	}
// 	if prs[0].State != models.PermissionRequestStates.Created {
// 		t.Errorf("Expected state to be Created. Got %s", prs[0].State.String())
// 		return
// 	}
// }
