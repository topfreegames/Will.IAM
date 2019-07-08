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

	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
	extensionsHttp "github.com/topfreegames/extensions/http"
)

const tokenEndpoint = "https://www.googleapis.com/oauth2/v4/token"
const userEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo"

// GoogleConfig are the basic required informations to use Google
// as oauth2 provider
type GoogleConfig struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	HostedDomains []string
}

var googleConfig GoogleConfig

func buildURL(endpoint, queryStrings string) string {
	return fmt.Sprintf("%s?%s", endpoint, queryStrings)
}

func mapToQueryStrings(m map[string]string) string {
	s := []string{}
	for k, v := range m {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(s, "&")
}

// Google implements Provider
type Google struct {
	config GoogleConfig
	repo   *repositories.All
	client *http.Client
}

// BuildAuthURL returns an URL authenticate with Google
func (g *Google) BuildAuthURL(state string) string {
	qs := mapToQueryStrings(map[string]string{
		"state":        state,
		"redirect_uri": g.config.RedirectURL,
		"client_id":    g.config.ClientID,
		"scope": strings.Join([]string{
			url.QueryEscape("https://www.googleapis.com/auth/userinfo.profile"),
			url.QueryEscape("https://www.googleapis.com/auth/userinfo.email"),
		}, "+"),
		"access_type":            "offline",
		"include_granted_scopes": "true",
		"response_type":          "code",
		"prompt":                 "consent",
	})
	return buildURL("https://accounts.google.com/o/oauth2/v2/auth", qs)
}

func (g *Google) buildExchangeCodeForm(code string) string {
	v := url.Values{}
	v.Add("code", code)
	v.Add("client_id", g.config.ClientID)
	v.Add("client_secret", g.config.ClientSecret)
	v.Add("redirect_uri", g.config.RedirectURL)
	v.Add("grant_type", "authorization_code")
	return v.Encode()
}

// ExchangeCode will trade code for full token with Google
func (g *Google) ExchangeCode(code string) (*models.AuthResult, error) {
	t, err := g.tokenFromCode(code)
	if err != nil {
		return nil, err
	}
	userInfo, err := g.getUserInfo(t.AccessToken)
	if err != nil {
		return nil, err
	}
	allowed := g.checkHostedDomain(userInfo.HostedDomain)
	if !allowed {
		return nil, errors.NewNonAllowedEmailDomainError(userInfo.HostedDomain)
	}
	t.Email = userInfo.Email
	// TODO: don't return sso_access_token to user, return 2 tokens to sso
	t.Expiry = time.Now().UTC().Add(14 * 24 * 3600 * time.Second)
	if err := g.repo.Tokens.Save(t); err != nil {
		return nil, err
	}
	return &models.AuthResult{
		AccessToken: t.AccessToken,
		Email:       t.Email,
		Picture:     userInfo.Picture,
	}, nil
}

// GoogleToken is the expected response for token endpoints
type GoogleToken struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    float64 `json:"expires_in"`
}

// Validate GoogleToken
func (gt GoogleToken) Validate() *models.Validation {
	validation := &models.Validation{}
	if gt.AccessToken == "" {
		validation.AddError("access_token", "required")
	}
	if gt.TokenType == "" {
		validation.AddError("token_type", "required")
	}
	if gt.ExpiresIn <= 0 {
		validation.AddError("expires_in", "should be greater than 0")
	}
	return validation
}

func (g *Google) postToTokenEndpoint(
	urlencoded string,
) (*GoogleToken, error) {
	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(urlencoded))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	gt := &GoogleToken{}
	err = json.Unmarshal(body, gt)
	if err != nil {
		return nil, err
	}
	v := gt.Validate()
	if !v.Valid() {
		return nil, v.Error()
	}
	return gt, nil
}

func (g *Google) tokenFromCode(code string) (*models.Token, error) {
	ecf := g.buildExchangeCodeForm(code)
	gt, err := g.postToTokenEndpoint(ecf)
	if err != nil {
		return nil, err
	}
	return &models.Token{
		AccessToken:  gt.AccessToken,
		RefreshToken: gt.RefreshToken,
		TokenType:    gt.TokenType,
		Expiry: time.Now().UTC().Add(
			time.Second * time.Duration(gt.ExpiresIn),
		),
	}, nil
}

type userInfo struct {
	Email        string `json:"email"`
	HostedDomain string `json:"hd"`
	Picture      string `json:"picture"`
}

func (g *Google) getUserInfo(accessToken string) (*userInfo, error) {
	req, err := http.NewRequest("GET", userEndpoint, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	if err != nil {
		return nil, err
	}
	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	ui := &userInfo{}
	err = json.Unmarshal(body, ui)
	if err != nil {
		return nil, err
	}
	return ui, nil
}

func (g *Google) checkHostedDomain(hd string) bool {
	if g.config.HostedDomains == nil || len(g.config.HostedDomains) == 0 {
		return true
	}
	for _, allowed := range g.config.HostedDomains {
		if hd == allowed {
			return true
		}
	}
	return false
}

func (g *Google) buildRefreshTokenForm(refreshToken string) string {
	v := url.Values{}
	v.Add("refresh_token", refreshToken)
	v.Add("client_id", g.config.ClientID)
	v.Add("client_secret", g.config.ClientSecret)
	v.Add("grant_type", "refresh_token")
	return v.Encode()
}

func (g *Google) maybeRefresh(t *models.Token) (*userInfo, error) {
	if t.Expiry.After(time.Now().UTC()) || !t.ExpiredAt.IsZero() {
		return nil, nil
	}
	rtf := g.buildRefreshTokenForm(t.RefreshToken)
	gt, err := g.postToTokenEndpoint(rtf)
	if err != nil {
		return nil, err
	}
	oldT := t.Clone()
	oldT.ExpiredAt.Time = time.Now().UTC()
	t.ID = ""
	t.AccessToken = gt.AccessToken
	t.Expiry = time.Now().UTC().Add(
		time.Second * time.Duration(gt.ExpiresIn),
	)
	userInfo, err := g.getUserInfo(t.AccessToken)
	if err != nil {
		return nil, err
	}
	if err := g.repo.WithPGTx(
		context.Background(), func(repo *repositories.All,
		) error {
			if err := g.repo.Tokens.Save(t); err != nil {
				return err
			}
			if err := g.repo.Tokens.Save(oldT); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return userInfo, nil
}

// Authenticate verifies if an accessToken is valid and maybe refresh it
func (g *Google) Authenticate(accessToken string) (*models.AuthResult, error) {
	t, err := g.repo.Tokens.Get(accessToken)
	if err != nil {
		return nil, err
	}
	var userInfo *userInfo
	if userInfo, err = g.maybeRefresh(t); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	authResult := &models.AuthResult{
		AccessToken: t.AccessToken,
		Email:       t.Email,
	}
	if userInfo != nil {
		authResult.Picture = userInfo.Picture
	}
	return authResult, nil
}

// WithContext returns a new instance of *Google using ctx
func (g Google) WithContext(ctx context.Context) Provider {
	return NewGoogle(g.config, g.repo.WithContext(ctx))
}

// NewGoogle ctor
func NewGoogle(
	config GoogleConfig, repo *repositories.All,
) *Google {
	return &Google{
		config: config,
		repo:   repo,
		client: extensionsHttp.New(),
	}
}
