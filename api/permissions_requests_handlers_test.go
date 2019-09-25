// +build integration

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	helpers "github.com/topfreegames/Will.IAM/testing"
)

func TestPermissionsRequestsListHandler(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	sa, err := saUC.CreateKeyPairType("some sa")
	if err != nil {
		t.Fatalf("Unexpected error %v", err.Error())
	}
	app := helpers.GetApp(t)
	req, _ := http.NewRequest("POST", "/permissions/requests", strings.NewReader(`{
"service": "SomeService",
"ownershipLevel": "RL",
"action": "SomeAction",
"resourceHierarchy": "*",
"message": "hey, can I have this permission?"
	}`))
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", sa.KeyID, sa.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201. Got %d", rec.Code)
	}
	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
	req, _ = http.NewRequest("GET", "/permissions/requests/open", nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec = helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200. Got %d", rec.Code)
	}
	body := map[string]interface{}{}
	err = json.Unmarshal([]byte(rec.Body.String()), &body)
	if err != nil {
		t.Fatalf("Unexpected error %v", err.Error())
	}
	if body["count"].(float64) != 1 {
		t.Fatalf("Expected to have 1 permission request. Got %f", body["count"])
	}
	prs := body["results"].([]interface{})
	pr := prs[0].(map[string]interface{})
	if pr["service"].(string) != "SomeService" {
		t.Errorf("Expected service to be SomeService. Got %s", pr["service"])
	}
	if pr["action"].(string) != "SomeAction" {
		t.Errorf("Expected action to be SomeAction. Got %s", pr["action"])
	}
	msg := "hey, can I have this permission?"
	if pr["message"].(string) != msg {
		t.Errorf("Expected message to be '%s'. Got '%s'", msg, pr["message"])
	}
	if pr["state"].(string) != "open" {
		t.Errorf("Expected state to be Open. Got %s", pr["state"])
	}
}
