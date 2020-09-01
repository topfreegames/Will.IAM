package oauth2_test

import (
	"testing"

	"github.com/topfreegames/Will.IAM/oauth2"
	helpers "github.com/topfreegames/Will.IAM/testing"
)

func TestGetOAuthProvider(t *testing.T) {
	config := helpers.GetConfig(t)
	repo := helpers.GetRepo(t)

	t.Run("GoogleProvider", func (t *testing.T) {
		config.Set("oauth2.provider", "google")
		provider := oauth2.GetOAuthProvider(config, repo)

		if _, ok := provider.(*oauth2.Google); !ok {
			t.Errorf("Expected provider *oauth2.Google, received %T", provider)
		} 
	})

	t.Run("DevOAuth2Provider", func (t *testing.T) {
		config.Set("oauth2.provider", "dev")
		provider := oauth2.GetOAuthProvider(config, repo)

		if _, ok := provider.(*oauth2.DevOAuth2Provider); !ok {
			t.Errorf("Expected provider *oauth2.DevOAuth2Provider, received %T", provider)
		} 
	})
}