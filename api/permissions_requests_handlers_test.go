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
		t.Errorf("Unexpected error %s", err.Error())
		return
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
		t.Errorf("Expected status 201. Got %d", rec.Code)
	}
	rootSA := helpers.CreateRootServiceAccount(t)
	req, _ = http.NewRequest("GET", "/permissions/requests/open", nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec = helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200. Got %d", rec.Code)
	}
	body := map[string]interface{}{}
	err = json.Unmarshal([]byte(rec.Body.String()), &body)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
	if body["count"].(float64) != 1 {
		t.Errorf("Expected to have 1 permission request. Got %f", body["count"])
		return
	}
	prs := body["results"].([]interface{})
	pr := prs[0].(map[string]interface{})
	if pr["service"].(string) != "SomeService" {
		t.Errorf("Expected service to be SomeService. Got %s", pr["service"])
		return
	}
	if pr["action"].(string) != "SomeAction" {
		t.Errorf("Expected action to be SomeAction. Got %s", pr["action"])
		return
	}
	msg := "hey, can I have this permission?"
	if pr["message"].(string) != msg {
		t.Errorf("Expected message to be '%s'. Got '%s'", msg, pr["message"])
		return
	}
	if pr["state"].(string) != "open" {
		t.Errorf("Expected state to be Open. Got %s", pr["state"])
		return
	}
}
