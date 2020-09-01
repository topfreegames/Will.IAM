package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
	extensionsHttp "github.com/topfreegames/extensions/http"
)

// DevOAuth2Provider is a Provider used in development environment
type DevOAuth2Provider struct {
	config DevOAuth2ProviderConfig
	repo   *repositories.All
}

// DevOAuth2ProviderConfig are the basic required informations to use
// our OAuth2 dev server as oauth2 provider
type DevOAuth2ProviderConfig struct {
	RedirectURL      string
	AuthorizationURL string
	TokenURL         string
}

// NewDevOAuth2Provider ctor
func NewDevOAuth2Provider(config DevOAuth2ProviderConfig, repo *repositories.All) *DevOAuth2Provider {
	return &DevOAuth2Provider{
		config: config,
		repo:   repo,
	}
}

// BuildAuthURL creates the url used to authorize an user against OAuth2 dev server
func (p *DevOAuth2Provider) BuildAuthURL(state string) string {
	return fmt.Sprintf("%s?response_type=code&redirect_uri=%s&state=%s", p.config.AuthorizationURL, p.config.RedirectURL, state)
}

// ExchangeCode validates an auth code against a OAuth2 server
func (p *DevOAuth2Provider) ExchangeCode(code string) (*models.AuthResult, error) {
	token, err := p.getToken(code)

	if err != nil {
		return nil, err
	}

	// with the token in hands we could go to the service behind the OAuth2 server
	// and retrieve some data, like user email and photo

	token.Email = "any@example.org"
	token.Expiry = time.Now().UTC().Add(14 * 24 * 3600 * time.Second)

	if err := p.repo.Tokens.Save(token); err != nil {
		return nil, err
	}
	return &models.AuthResult{
		AccessToken: token.AccessToken,
		Email:       token.Email,
		Picture:     "http://lorempixel.com/400/200/cats",
	}, nil
}

// Authenticate verifies if an accessToken is valid and maybe refresh it
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

// OAuthToken represents a token received from an OAuth2 server
type OAuthToken struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    float64 `json:"expires_in"`
}

func (p *DevOAuth2Provider) getToken(code string) (*models.Token, error) {
	v := url.Values{}
	v.Add("code", code)
	v.Add("redirect_uri", p.config.RedirectURL)
	v.Add("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", p.config.TokenURL, strings.NewReader(v.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return nil, err
	}

	client := extensionsHttp.New()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	oauthToken := &OAuthToken{}
	err = json.Unmarshal(body, oauthToken)
	if err != nil {
		return nil, err
	}

	accessToken := oauthToken.AccessToken

	// TODO: Will.IAM is not able to support JWT tokens due a restriction of 300 characters
	// in accessToken field when saving a Token on database
	if len(accessToken) > 300 {
		accessToken = accessToken[0:300]
	}

	return &models.Token{
		AccessToken:  accessToken,
		RefreshToken: oauthToken.RefreshToken,
		TokenType:    oauthToken.TokenType,
		Expiry: time.Now().UTC().Add(
			time.Second * time.Duration(oauthToken.ExpiresIn),
		),
	}, nil
}
