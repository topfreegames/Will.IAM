// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
		rootSA := helpers.CreateRootServiceAccountWithKeyPair(t)
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

func TestServiceAccountListHandler(t *testing.T) {
	beforeEachServiceAccountsHandlers(t)

	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t)

	app := helpers.GetApp(t)

	req, _ := http.NewRequest(http.MethodGet, "/service_accounts", nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d. Got %d", http.StatusOK, rec.Code)
	}

	jsRet := struct {
		Count  int64 `json:"count"`
		Result []struct {
			ID                 string `json:"id"`
			AuthenticationType string `json:"authenticationType"`
			Email              string `json:"email"`
			Name               string `json:"name"`
			Picture            string `json:"picture"`
		} `json:"results"`
	}{}

	json.Unmarshal(rec.Body.Bytes(), &jsRet)
	if jsRet.Count != 1 {
		t.Errorf("Expected count %d. Got %d", 1, jsRet.Count)
	}

	if len(jsRet.Result) != 1 {
		t.Fatalf("Expected result len %d. Got %d", 1, len(jsRet.Result))
	}

	if jsRet.Result[0].Name != "root" {
		t.Errorf("Expected name %s. Got %s", "root", jsRet.Result[0].Name)
	}

	if jsRet.Result[0].ID != rootSA.ID {
		t.Errorf("Expected id %s. Got %s", rootSA.ID, jsRet.Result[0].ID)
	}
}

func TestServiceAccountListWithPermissionHandler(t *testing.T) {
	app := helpers.GetApp(t)

	params := url.Values{}
	params.Set("permission", "SC::RL::Edit::*")

	saListWithPermissionTestCases := []struct {
		sasPs    []string
		test     string
		expected []string
	}{
		{
			sasPs: []string{
				"Service1::RL::Do1::x::*",
				"Service1::RL::Do1::x::y",
				"Service1::RL::Do1::x::z",
			},
			test:     "Service1::RL::Do1::x::z",
			expected: []string{"root", "sa0", "sa2"},
		},
		{
			sasPs: []string{
				"Service1::RL::Do1::x::*",
				"Service1::RL::Do1::x::y",
				"Service1::RO::Do1::x::z",
			},
			test:     "Service1::RO::Do1::x::z",
			expected: []string{"root", "sa2"},
		},
		{
			sasPs: []string{
				"Service1::RL::Do1::x::*",
				"Service1::RL::Do1::x::y",
				"Service1::RO::Do1::x::z",
			},
			test:     "Service2::RO::Do1::x::z",
			expected: []string{"root"},
		},
		{
			sasPs: []string{
				"Service1::RL::Do1::x::*",
				"Service1::RL::Do1::x::y",
				"Service1::RO::*::x::z",
			},
			test:     "Service1::RO::Do1::x::z",
			expected: []string{"root", "sa2"},
		},
	}

	for caseID, tt := range saListWithPermissionTestCases {
		beforeEachServiceAccountsHandlers(t)
		rootSA := helpers.CreateRootServiceAccountWithKeyPair(t)

		for i, p := range tt.sasPs {
			helpers.CreateServiceAccountWithPermissions(t, fmt.Sprintf("sa%d", i), fmt.Sprintf("sa%d@email.com", i), "OAuth", p)
		}

		params := url.Values{}
		params.Set("permission", tt.test)

		req, _ := http.NewRequest(http.MethodGet, "/service_accounts/with_permission?"+params.Encode(), nil)
		req.Header.Set("Authorization", fmt.Sprintf(
			"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
		))
		rec := helpers.DoRequest(t, req, app.GetRouter())
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d. Got %d", http.StatusOK, rec.Code)
		}

		jsRet := struct {
			Count  int64 `json:"count"`
			Result []struct {
				ID                 string `json:"id"`
				AuthenticationType string `json:"authenticationType"`
				Email              string `json:"email"`
				Name               string `json:"name"`
				Picture            string `json:"picture"`
			} `json:"results"`
		}{}

		json.Unmarshal(rec.Body.Bytes(), &jsRet)
		if int(jsRet.Count) != len(tt.expected) {
			t.Errorf("Expected count %d. Got %d", len(tt.expected), jsRet.Count)
		}

		if len(jsRet.Result) != len(tt.expected) {
			t.Fatalf("Expected result len %d. Got %d", len(tt.expected), len(jsRet.Result))
		}

		for i := range tt.expected {
			if tt.expected[i] != jsRet.Result[i].Name {
				t.Errorf("Expected list[%d] to be %s. Got %s. Case: %d", i, tt.expected[i], jsRet.Result[i].Name, caseID)
			}
		}
	}
}
