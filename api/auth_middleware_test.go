package api_test

import (
	"fmt"
	"net/http"
	"testing"

	helpers "github.com/topfreegames/Will.IAM/testing"
)

func TestAuthMiddlewareKeyPairShouldAuthenticateUser(t *testing.T) {
	helpers.CleanupPG(t)

	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t)
	app := helpers.GetApp(t)
	req, _ := http.NewRequest(http.MethodGet, "/service_accounts", nil)

	req.Header.Set("Authorization", fmt.Sprintf(
		"keypair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))

	response := helpers.DoRequest(t, req, app.GetRouter())

	if response.Code != http.StatusOK {
		t.Errorf("Expected status %d. Got %d", http.StatusOK, response.Code)
	}

	serviceAccountName := response.Header().Get("x-service-account-name")
	if serviceAccountName != "root" {
		t.Errorf("Expected service account name %s. Got %s", "root", serviceAccountName)
	}
}

func TestAuthMiddlewareBearerShouldAuthenticateUser(t *testing.T) {
	helpers.CleanupPG(t)

	rootSA := helpers.CreateRootServiceAccountWithOAuth(t)
	tokensRepo := helpers.GetRepo(t).Tokens
	tokens, _ := tokensRepo.FindByEmail(rootSA.Email)
	token := tokens[0]
	app := helpers.GetApp(t)
	req, _ := http.NewRequest(http.MethodGet, "/service_accounts", nil)

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", token))
	response := helpers.DoRequest(t, req, app.GetRouter())

	if response.Code != http.StatusOK {
		t.Errorf("Expected status %d. Got %d", http.StatusOK, response.Code)
	}

	serviceAccountEmail := response.Header().Get("x-email")
	if serviceAccountEmail != rootSA.Email {
		t.Errorf("Expected service account email %s. Got %s", rootSA.Email, serviceAccountEmail)
	}
}
