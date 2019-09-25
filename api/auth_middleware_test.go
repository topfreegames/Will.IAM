package api_test

import (
	b64 "encoding/base64"
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
	emptyMap := make(map[string]string)

	testCases := []struct {
		testName                string
		serviceAccount          *models.ServiceAccount
		requestHeaders          map[string]string
		expectedResponseHeaders map[string]string
		expectedResponseCode    int
	}{
		{
			testName:                "KeyPairAuthorization",
			serviceAccount:          keyPairSA,
			requestHeaders:          map[string]string{"authorization": fmt.Sprintf("keypair %s:%s", keyPairSA.KeyID, keyPairSA.KeySecret)},
			expectedResponseHeaders: map[string]string{"x-service-account-name": "keyPairUser"},
			expectedResponseCode:    http.StatusOK,
		},
		{
			testName:                "OAuthAuthorization",
			serviceAccount:          oauthSA,
			requestHeaders:          map[string]string{"authorization": fmt.Sprintf("bearer %s", token.AccessToken)},
			expectedResponseHeaders: map[string]string{"x-email": oauthSA.Email},
			expectedResponseCode:    http.StatusOK,
		},
		{
			testName:                "WrongOAuthTokenAuthorization",
			serviceAccount:          oauthSA,
			requestHeaders:          map[string]string{"authorization": fmt.Sprintf("bearer %s", "wrong token")},
			expectedResponseHeaders: emptyMap,
			expectedResponseCode:    http.StatusUnauthorized,
		},
		{
			testName:                "WrongKeyPairAuthorization",
			serviceAccount:          keyPairSA,
			requestHeaders:          map[string]string{"authorization": fmt.Sprintf("keypair %s:%s", "wrong key_id", "wrong_key_secret")},
			expectedResponseHeaders: emptyMap,
			expectedResponseCode:    http.StatusUnauthorized,
		},
		{
			testName:                "UndefinedAuthorization",
			serviceAccount:          nil,
			requestHeaders:          emptyMap,
			expectedResponseHeaders: emptyMap,
			expectedResponseCode:    http.StatusUnauthorized,
		},
		{
			testName:                "UnsupportedAuthorization",
			serviceAccount:          oauthSA,
			requestHeaders:          map[string]string{"authorization": fmt.Sprintf("basic %s", b64.StdEncoding.EncodeToString([]byte("user:password")))},
			expectedResponseHeaders: emptyMap,
			expectedResponseCode:    http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			app := helpers.GetApp(t)
			req, err := http.NewRequest(http.MethodGet, "/service_accounts", nil)

			if err != nil {
				t.Errorf("Could not create HTTP request")
			}

			for header, content := range testCase.requestHeaders {
				req.Header.Set(header, content)
			}

			response := helpers.DoRequest(t, req, app.GetRouter())

			if response.Code != testCase.expectedResponseCode {
				t.Errorf("Expected status %d. Got %d", testCase.expectedResponseCode, response.Code)
			}

			for header, expectedHeaderVal := range testCase.expectedResponseHeaders {
				headerVal := response.Header().Get(header)
				if headerVal != expectedHeaderVal {
					t.Errorf("Expected Header %s to be %s, got %s", header, expectedHeaderVal, headerVal)
				}
			}
		})
	}
}
