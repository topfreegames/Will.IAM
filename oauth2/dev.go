package oauth2

import (
	"context"
	"fmt"

	"github.com/topfreegames/Will.IAM/models"
)

// DevOAuth2Provider is a Provider used in development environment 
type DevOAuth2Provider struct {
	config DevOAuth2ProviderConfig
}

// DevOAuth2ProviderConfig are the basic required informations to use 
// our OAuth2 dev server as oauth2 provider
type DevOAuth2ProviderConfig struct {
	RedirectURL      string
	AuthorizationURL string
}

// NewDevOAuth2Provider ctor
func NewDevOAuth2Provider(config DevOAuth2ProviderConfig) *DevOAuth2Provider {
	return &DevOAuth2Provider{ config: config }
}

// BuildAuthURL creates the url used to authorize an user against OAuth2 dev server
func (p *DevOAuth2Provider) BuildAuthURL(any string) string {
	return fmt.Sprintf("%s?response_type=code&redirect_uri=%s", p.config.AuthorizationURL, p.config.RedirectURL)
}

// ExchangeCode dummy
func (p *DevOAuth2Provider) ExchangeCode(any string) (*models.AuthResult, error) {
	return &models.AuthResult{
		AccessToken: "any",
		Email:       "any",
	}, nil
}

// Authenticate dummy
func (p *DevOAuth2Provider) Authenticate(accessToken string) (*models.AuthResult, error) {
	return &models.AuthResult{
		AccessToken: accessToken,
		Email:       "any",
	}, nil
}

// WithContext does nothing
func (p *DevOAuth2Provider) WithContext(ctx context.Context) Provider {
	return p
}
