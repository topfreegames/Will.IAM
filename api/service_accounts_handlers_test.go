// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	helpers "github.com/topfreegames/Will.IAM/testing"
)

func beforeEachServiceAccountsHandlers(t *testing.T) {
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

func TestServiceAccountCreateHandler(t *testing.T) {
	type createTest struct {
		body           map[string]interface{}
		expectedStatus int
	}
	tt := []createTest{
		createTest{
			body: map[string]interface{}{
				"name": "some name",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		createTest{
			body: map[string]interface{}{
				"name":               "some name",
				"authenticationType": "keypair",
			},
			expectedStatus: http.StatusCreated,
		},
		createTest{
			body: map[string]interface{}{
				"name":               "some name",
				"authenticationType": "oauth2",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		createTest{
			body: map[string]interface{}{
				"name":               "some name",
				"email":              "email@email.com",
				"authenticationType": "oauth2",
			},
			expectedStatus: http.StatusCreated,
		},
		createTest{
			body:           map[string]interface{}{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	app := helpers.GetApp(t)
	for _, tt := range tt {
		beforeEachServiceAccountsHandlers(t)
		rootSA := helpers.CreateRootServiceAccount(t)
		bts, err := json.Marshal(tt.body)
		if err != nil {
			t.Errorf("Unexpected error %s", err.Error())
			return
		}
		req, _ := http.NewRequest("POST", "/service_accounts", bytes.NewBuffer(bts))
		req.Header.Set("Authorization", fmt.Sprintf(
			"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
		))
		rec := helpers.DoRequest(t, req, app.GetRouter())
		if rec.Code != tt.expectedStatus {
			t.Errorf("Expected status %d. Got %d", tt.expectedStatus, rec.Code)
		}
	}
}
