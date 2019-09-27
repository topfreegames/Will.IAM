package api_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/topfreegames/Will.IAM/models"
	helpers "github.com/topfreegames/Will.IAM/testing"
)

func TestAuthMiddleware(t *testing.T) {
	helpers.CleanupPG(t)

	oauthSA := helpers.CreateRootServiceAccountWithOAuth(t, "oauthUser", "oauth.user@test.com")
	keyPairSA := helpers.CreateRootServiceAccountWithKeyPair(t, "keyPairUser", "keypair.user@test.com")
	tokensRepo := helpers.GetRepo(t).Tokens
	tokens, _ := tokensRepo.FindByEmail(oauthSA.Email)
	token := tokens[0]

	testCases := []struct {
		name                string
		serviceAccount      *models.ServiceAccount
		requestHeaders      map[string]string
		wantResponseHeaders map[string]string
		wantResponseCode    int
	}{
		{
			name:                "KeyPairAuthorization",
			serviceAccount:      keyPairSA,
			requestHeaders:      map[string]string{"authorization": fmt.Sprintf("keypair %v:%v", keyPairSA.KeyID, keyPairSA.KeySecret)},
			wantResponseHeaders: map[string]string{"x-service-account-name": "keyPairUser"},
			wantResponseCode:    http.StatusOK,
		},
		{
			name:                "OAuthAuthorization",
			serviceAccount:      oauthSA,
			requestHeaders:      map[string]string{"authorization": fmt.Sprintf("bearer %v", token.AccessToken)},
			wantResponseHeaders: map[string]string{"x-email": oauthSA.Email},
			wantResponseCode:    http.StatusOK,
		},
		{
			name:             "WrongOAuthTokenAuthorization",
			serviceAccount:   oauthSA,
			requestHeaders:   map[string]string{"authorization": "bearer wrong_token"},
			wantResponseCode: http.StatusUnauthorized,
		},
		{
			name:             "WrongKeyPairAuthorization",
			serviceAccount:   keyPairSA,
			requestHeaders:   map[string]string{"authorization": "keypair wrong_key_id:wrong_key_secret"},
			wantResponseCode: http.StatusUnauthorized,
		},
		{
			name:             "IncompleteOAuthTokenAuthorization",
			serviceAccount:   keyPairSA,
			requestHeaders:   map[string]string{"authorization": "bearer"},
			wantResponseCode: http.StatusUnauthorized,
		},
		{
			name:             "IncompleteKeyPairAuthorization",
			serviceAccount:   keyPairSA,
			requestHeaders:   map[string]string{"authorization": "keypair wrong_key_id"},
			wantResponseCode: http.StatusUnauthorized,
		},
		{
			name:             "UndefinedAuthorization",
			serviceAccount:   nil,
			wantResponseCode: http.StatusUnauthorized,
		},
		{
			name:             "UnsupportedAuthorization",
			serviceAccount:   oauthSA,
			requestHeaders:   map[string]string{"authorization": fmt.Sprintf("basic %v", base64.StdEncoding.EncodeToString([]byte("user:password")))},
			wantResponseCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app := helpers.GetApp(t)
			req, err := http.NewRequest(http.MethodGet, "/service_accounts", nil)
			if err != nil {
				t.Fatalf("Could not create HTTP request")
			}

			for header, content := range testCase.requestHeaders {
				req.Header.Set(header, content)
			}

			response := helpers.DoRequest(t, req, app.GetRouter())

			if response.Code != testCase.wantResponseCode {
				t.Errorf("Status = %v, want %v", response.Code, testCase.wantResponseCode)
			}

			for header, wantHeader := range testCase.wantResponseHeaders {
				if gotHeader := response.Header().Get(header); gotHeader != wantHeader {
					t.Errorf("Header %v = %v, want %v", header, gotHeader, wantHeader)
				}
			}
		})
	}
}
