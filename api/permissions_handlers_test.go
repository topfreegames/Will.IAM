// +build integration

package api_test

import (
	"fmt"
	"github.com/topfreegames/Will.IAM/models"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	helpers "github.com/topfreegames/Will.IAM/testing"
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
	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
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
	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
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
	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
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

func TestPermissionsHasHandler(t *testing.T) {
	beforeEachPermissionsHandlers(t)
	type hasPermissionTest struct {
		name        string
		request     string
		wantStatus  int
		wantMessage string
	}
	testCases := []hasPermissionTest{
		hasPermissionTest{
			name:        "MissingQueryString",
			request:     "/permissions/has",
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: `{"error": "querystrings.permission is required"}`,
		},
		hasPermissionTest{
			name:        "MissingPermissionQueryStringContent",
			request:     "/permissions/has?permission=",
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: `{"error": "Incomplete permission. Expected format: Service::OwnershipLevel::Action::{ResourceHierarchy}"}`,
		},
		hasPermissionTest{
			name:        "WrongFormatPermissionQueryString",
			request:     "/permissions/has?permission=X",
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: `{"error": "Incomplete permission. Expected format: Service::OwnershipLevel::Action::{ResourceHierarchy}"}`,
		},
		hasPermissionTest{
			name:        "NotAuthorizedPermission",
			request:     "/permissions/has?permission=Service::RL::TestAction::*",
			wantStatus:  http.StatusForbidden,
			wantMessage: "",
		},
		hasPermissionTest{
			name:        "AuthorizedPermission",
			request:     "/permissions/has?permission=Service::RL::TestAction2::*",
			wantStatus:  http.StatusOK,
			wantMessage: "",
		},
	}

	sa := helpers.CreateServiceAccountWithPermissions(t, "sa", "sa@test.com", models.AuthenticationTypes.KeyPair, "Service::RL::TestAction2::*")
	app := helpers.GetApp(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", testCase.request, nil)
			req.Header.Set("Authorization", fmt.Sprintf(
				"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
			))

			rec := helpers.DoRequest(t, req, app.GetRouter())

			if rec.Code != testCase.wantStatus {
				t.Errorf("Expected HTTP status %d. Got %d", testCase.wantStatus, rec.Code)
			}

			if rec.Body.String() != testCase.wantMessage {
				t.Errorf("Expected response body %s. Got %s", testCase.wantMessage, rec.Body)
			}
		})
	}
}
