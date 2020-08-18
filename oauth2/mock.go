package oauth2

import (
	"context"
	"fmt"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// ProviderBlankMock is a Provider mock will all dummy implementations
type ProviderBlankMock struct {
	Email string
	repo  *repositories.All
}

// NewProviderBlankMock ctor
func NewProviderBlankMock(repo *repositories.All) *ProviderBlankMock {
	return &ProviderBlankMock{ repo: repo }
}

// BuildAuthURL dummy
func (p *ProviderBlankMock) BuildAuthURL(any string) string {
	endpoint := "http://localhost:9000/authorize"
	redirectUri := "http://localhost:4040/sso/auth/done"

	return fmt.Sprintf("%s?response_type=code&redirect_uri=%s", endpoint, redirectUri)
}

// ExchangeCode dummy
func (p *ProviderBlankMock) ExchangeCode(any string) (*models.AuthResult, error) {
	return &models.AuthResult{
		AccessToken: "any",
		Email:       "any",
	}, nil
}

// Authenticate dummy
func (p *ProviderBlankMock) Authenticate(accessToken string) (*models.AuthResult, error) {
	return &models.AuthResult{
		AccessToken: accessToken,
		Email:       "any",
	}, nil
}

// WithContext does nothing
func (p *ProviderBlankMock) WithContext(ctx context.Context) Provider {
	return p
}
