package repositories

import (
	"context"

	"github.com/topfreegames/extensions/pg"
)

// All holds a reference to each possible repository interface
type All struct {
	Permissions
	PermissionsRequests
	Roles
	ServiceAccounts
	Services
	Tokens
	Healthcheck
	storage *Storage
}

// New All ctor
func New(s *Storage) *All {
	return &All{
		Permissions:         NewPermissions(s),
		PermissionsRequests: NewPermissionsRequests(s),
		Roles:               NewRoles(s),
		ServiceAccounts:     NewServiceAccounts(s),
		Services:            NewServices(s),
		Tokens:              NewTokens(s),
		Healthcheck:         NewHealthcheck(s),
		storage:             s,
	}
}

// WithContext clones All and all its contents and injects a context
// in all storages
func (a *All) WithContext(ctx context.Context) *All {
	s := a.storage.Clone()
	s.PG.DB = pg.WithContext(ctx, s.PG.DB)
	return a.cloneWithStorage(s)
}

// WithPGTx clones All and all its contents and injects a PG tx
// in it's storage.PG.DB and in all inner repo storages
func (a *All) WithPGTx(ctx context.Context, fn func(repo *All) error) error {
	s := a.storage.Clone()
	tx, err := a.storage.PG.Begin(a.storage.PG.DB.WithContext(ctx))
	if err != nil {
		return err
	}
	s.PG.DB = tx
	c := a.cloneWithStorage(s)

	defer pg.Rollback(c.storage.PG.DB)
	err = fn(c)
	if err == nil {
		return pg.Commit(c.storage.PG.DB)
	}
	return err
}

func (a *All) cloneWithStorage(s *Storage) *All {
	c := &All{
		Permissions:         a.Permissions.Clone(),
		PermissionsRequests: a.PermissionsRequests.Clone(),
		Roles:               a.Roles.Clone(),
		ServiceAccounts:     a.ServiceAccounts.Clone(),
		Services:            a.Services.Clone(),
		Tokens:              a.Tokens.Clone(),
		storage:             s,
	}
	c.Permissions.setStorage(s)
	c.PermissionsRequests.setStorage(s)
	c.Roles.setStorage(s)
	c.ServiceAccounts.setStorage(s)
	c.Services.setStorage(s)
	c.Tokens.setStorage(s)
	return c
}
