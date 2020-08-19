package oauth2

import (
	"context"

	"github.com/spf13/viper"
	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// Provider is the contract any OAuth2 implementation must follow
type Provider interface {
	BuildAuthURL(string) string
	ExchangeCode(string) (*models.AuthResult, error)
	Authenticate(string) (*models.AuthResult, error)
	WithContext(context.Context) Provider
}

// Given a provider selection on config, returns its instance
func GetOAuthProvider(config *viper.Viper, repo *repositories.All) Provider {
	providerType := config.GetString("oauth2.provider")

	if providerType == "dev" {
		return NewDevOAuth2Provider(DevOAuth2ProviderConfig{
			RedirectURL:      config.GetString("oauth2.dev.redirectUrl"),
			AuthorizationURL: config.GetString("oauth2.dev.authorizationUrl"),
		})
	}

	return NewGoogle(GoogleConfig{
		ClientID:          config.GetString("oauth2.google.clientId"),
		ClientSecret:      config.GetString("oauth2.google.clientSecret"),
		RedirectURL:       config.GetString("oauth2.google.redirectUrl"),
		CheckHostedDomain: config.GetBool("oauth2.google.checkHostedDomain"),
		HostedDomains:     config.GetStringSlice("oauth2.google.hostedDomains"),
	}, repo)
}