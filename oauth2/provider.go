package oauth2

import (
	"context"

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

// ProviderBlankMock is a Provider mock will all dummy implementations
type ProviderBlankMock struct {
	Email string
	repo  *repositories.All
}

// NewProviderBlankMock ctor
func NewProviderBlankMock() *ProviderBlankMock {
	return &ProviderBlankMock{}
}

// BuildAuthURL dummy
func (p *ProviderBlankMock) BuildAuthURL(any string) string {
	return "any"
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
	tokensRepo := p.repo.Tokens
	token, _ := tokensRepo.Get(accessToken)

	return &models.AuthResult{
		AccessToken: token.AccessToken,
		Email:       token.Email,
	}, nil
}

// WithContext does nothing
func (p *ProviderBlankMock) WithContext(ctx context.Context) Provider {
	return p
}
