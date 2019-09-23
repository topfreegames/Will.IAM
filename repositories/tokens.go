package repositories

import (
	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/models"
)

// Tokens contract
type Tokens interface {
	Get(string) (*models.Token, error)
	Save(*models.Token) error
	Clone() Tokens
	setStorage(*Storage)
	FindByEmail(string) ([]models.Token, error)
}

type tokens struct {
	*withStorage
}

func (ts *tokens) Clone() Tokens {
	return NewTokens(ts.storage.Clone())
}

func (ts tokens) Get(accessToken string) (*models.Token, error) {
	t := new(models.Token)
	if _, err := ts.storage.PG.DB.Query(
		t, `SELECT * FROM tokens WHERE access_token = ?0 AND
		(expired_at IS NULL OR expired_at > now() - INTERVAL '60 sec')`,
		accessToken,
	); err != nil {
		return nil, err
	}
	if t.AccessToken == "" {
		return nil, errors.NewEntityNotFoundError(models.Token{}, accessToken)
	}
	return t, nil
}

func (ts tokens) Save(token *models.Token) error {
	_, err := ts.storage.PG.DB.Exec(`INSERT INTO tokens (access_token,
	refresh_token, expired_at, token_type, expiry, email, updated_at)
	VALUES (?access_token, ?refresh_token, ?expired_at, ?token_type,
	?expiry, ?email, now()) ON CONFLICT (access_token) DO UPDATE SET
	expired_at = ?expired_at, updated_at = now()`, token)
	return err
}

func (ts tokens) FindByEmail(email string) ([]models.Token, error) {
	tokens := []models.Token{}
	if _, err := ts.storage.PG.DB.Query(
		&tokens, `SELECT * FROM tokens WHERE email = ?`, email,
	); err != nil {
		return nil, err
	}
	return tokens, nil
}

// NewTokens ctor
func NewTokens(storage *Storage) Tokens {
	return &tokens{&withStorage{storage: storage}}
}
