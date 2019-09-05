package api_test

import (
	"fmt"
	"net/http"
	"testing"

	helpers "github.com/topfreegames/Will.IAM/testing"
)

func TestAuthMiddlewareKeyPair(t *testing.T) {
	helpers.CleanupPG(t)

	rootSA := helpers.CreateRootServiceAccount(t)

	app := helpers.GetApp(t)

	req, _ := http.NewRequest(http.MethodGet, "/service_accounts", nil)
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d. Got %d", http.StatusOK, rec.Code)
	}

	serviceAccountName := rec.Header().Get("x-service-account-name")
	if serviceAccountName != "root" {
		t.Errorf("Expected service account name %s. Got %s", "root", serviceAccountName)
	}
}
