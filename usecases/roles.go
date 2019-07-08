package usecases

import (
	"context"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// Roles define entrypoints for ServiceAccount actions
type Roles interface {
	Create(*RoleWithNested) error
	CreatePermission(string, *models.Permission) error
	Update(*RoleWithNested) error
	Get(string) (map[string]interface{}, error)
	GetPermissions(string) ([]models.Permission, error)
	GetServiceAccounts(string) ([]models.ServiceAccount, error)
	WithNamePrefix(string, int) ([]models.Role, error)
	List(*repositories.ListOptions) ([]models.Role, int64, error)
	Search(string, *repositories.ListOptions) ([]models.Role, int64, error)
	WithContext(context.Context) Roles
}

type roles struct {
	repo *repositories.All
	ctx  context.Context
}

func (rs roles) WithContext(ctx context.Context) Roles {
	return &roles{rs.repo.WithContext(ctx), ctx}
}

func (rs roles) Create(rwn *RoleWithNested) error {
	return rs.repo.WithPGTx(rs.ctx, func(repo *repositories.All) error {
		role := &models.Role{Name: rwn.Name}
		if err := repo.Roles.Create(role); err != nil {
			return err
		}
		rwn.ID = role.ID
		for i := range rwn.Permissions {
			rwn.Permissions[i].RoleID = role.ID
			if err := createPermission(repo, &rwn.Permissions[i]); err != nil {
				return err
			}
		}
		for i := range rwn.ServiceAccountsIDs {
			if err := repo.Roles.Bind(&models.RoleBinding{
				RoleID:           role.ID,
				ServiceAccountID: rwn.ServiceAccountsIDs[i],
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (rs roles) CreatePermission(roleID string, p *models.Permission) error {
	p.RoleID = roleID
	return createPermission(rs.repo, p)
}

func createPermission(repo *repositories.All, p *models.Permission) error {
	return repo.Permissions.Create(p)
}

// RoleWithNested is the required data to update a role
type RoleWithNested struct {
	ID                 string              `json:"-"`
	Name               string              `json:"name"`
	PermissionsStrings []string            `json:"permissions"`
	PermissionsAliases map[string]string   `json:"permissionsAliases"`
	Permissions        []models.Permission `json:"-"`
	ServiceAccountsIDs []string            `json:"serviceAccountsIds"`
}

// Validate RoleWithNested fields
func (rwn RoleWithNested) Validate() models.Validation {
	v := &models.Validation{}
	if rwn.Name == "" {
		v.AddError("name", "required")
	}
	return *v
}

func (rs roles) Update(rwn *RoleWithNested) error {
	return rs.repo.WithPGTx(rs.ctx, func(repo *repositories.All) error {
		if err := repo.Roles.DropPermissions(rwn.ID); err != nil {
			return err
		}
		for i := range rwn.Permissions {
			rwn.Permissions[i].RoleID = rwn.ID
			if err := createPermission(repo, &rwn.Permissions[i]); err != nil {
				return err
			}
		}
		if err := repo.Roles.DropBindings(rwn.ID); err != nil {
			return err
		}
		for i := range rwn.ServiceAccountsIDs {
			if err := repo.Roles.Bind(&models.RoleBinding{
				RoleID:           rwn.ID,
				ServiceAccountID: rwn.ServiceAccountsIDs[i],
			}); err != nil {
				return err
			}
		}
		role := &models.Role{ID: rwn.ID, Name: rwn.Name}
		return repo.Roles.Update(role)
	})
}

func (rs roles) GetPermissions(roleID string) ([]models.Permission, error) {
	return rs.repo.Permissions.ForRole(roleID)
}

func (rs roles) GetServiceAccounts(
	roleID string,
) ([]models.ServiceAccount, error) {
	return rs.repo.Roles.GetServiceAccounts(roleID)
}

func (rs roles) WithNamePrefix(
	prefix string, maxResults int,
) ([]models.Role, error) {
	return rs.repo.Roles.WithNamePrefix(prefix, maxResults)
}

func (rs roles) List(
	lo *repositories.ListOptions,
) ([]models.Role, int64, error) {
	rsSl, err := rs.repo.Roles.List(lo)
	if err != nil {
		return nil, 0, err
	}
	count, err := rs.repo.Roles.ListCount()
	if err != nil {
		return nil, 0, err
	}
	return rsSl, count, nil
}

// Search over Roles names
func (rs roles) Search(
	term string, lo *repositories.ListOptions,
) ([]models.Role, int64, error) {
	rsSl, err := rs.repo.Roles.Search(term, lo)
	if err != nil {
		return nil, 0, err
	}
	count, err := rs.repo.Roles.SearchCount(term)
	if err != nil {
		return nil, 0, err
	}
	return rsSl, count, nil
}

func (rs roles) Get(id string) (map[string]interface{}, error) {
	r, err := rs.repo.Roles.Get(id)
	if err != nil {
		return nil, err
	}
	pSl, err := rs.GetPermissions(id)
	if err != nil {
		return nil, err
	}
	permissionsAliases := map[string]string{}
	permissions := make([]string, len(pSl))
	for i := range pSl {
		str := pSl[i].String()
		permissions[i] = str
		if pSl[i].Alias != "" {
			permissionsAliases[str] = pSl[i].Alias
		}
	}
	sas, err := rs.GetServiceAccounts(id)
	if err != nil {
		return nil, err
	}
	sasFiltered := make([]map[string]interface{}, len(sas))
	for i, sa := range sas {
		sasFiltered[i] = map[string]interface{}{
			"id":      sa.ID,
			"name":    sa.Name,
			"picture": sa.Picture,
			"email":   sa.Email,
		}
	}
	return map[string]interface{}{
		"id":                 r.ID,
		"name":               r.Name,
		"permissions":        permissions,
		"permissionsAliases": permissionsAliases,
		"serviceAccounts":    sasFiltered,
	}, nil
}

// NewRoles ctor
func NewRoles(repo *repositories.All) Roles {
	return &roles{repo: repo}
}
