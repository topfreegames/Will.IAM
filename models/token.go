package models

import (
	"time"

	"github.com/go-pg/pg"
)

// Token type
type Token struct {
	ID           string      `json:"-" pg:"id"`
	AccessToken  string      `json:"access_token" pg:"access_token"`
	RefreshToken string      `json:"refresh_token" pg:"refresh_token"`
	TokenType    string      `json:"token_type" pg:"token_type"`
	Expiry       time.Time   `json:"expiry" pg:"expiry"`
	ExpiredAt    pg.NullTime `json:"expiredAt" pg:"expired_at"`
	Email        string      `json:"email" pg:"email"`
	CreatedUpdatedAt
}

// Clone a token
func (t Token) Clone() *Token {
	tt := &Token{}
	*tt = t
	return tt
}

// AccessTokenAuth stores a ServiceAccountID and the (maybe refreshed)
// AccessToken
type AccessTokenAuth struct {
	ServiceAccountID string
	AccessToken      string
	Email            string
}

// AccessKeyPairAuth store the ServicesAccountID and the name of the service
type AccessKeyPairAuth struct {
	ServiceAccountID string
	Name             string
}

// AuthResult is the result of a successful authentication
type AuthResult struct {
	AccessToken string `json:"accessToken"`
	Email       string `json:"email"`
	Picture     string `json:"picture"`
}
