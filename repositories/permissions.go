package repositories

import (
	"github.com/ghostec/Will.IAM/errors"
	"github.com/ghostec/Will.IAM/models"
)

// Permissions repository
type Permissions interface {
	Get(string) (*models.Permission, error)
	ForServiceAccount(string) ([]models.Permission, error)
	ForRole(string) ([]models.Permission, error)
	Create(*models.Permission) error
	CreateRequest(string, *models.PermissionRequest) error
	GetPermissionRequests(string) ([]models.PermissionRequest, error)
	Delete(string) error
	Clone() Permissions
	setStorage(*Storage)
}

type permissions struct {
	*withStorage
}

func (ps *permissions) Clone() Permissions {
	return NewPermissions(ps.storage.Clone())
}

// Get retrieves a permission by id
func (ps *permissions) Get(id string) (*models.Permission, error) {
	p := new(models.Permission)
	if info, err := ps.storage.PG.DB.Query(
		p, `SELECT id, role_id, service, ownership_level,
action, resource_hierarchy, alias FROM permissions
	WHERE id = ?`, id,
	); err != nil {
		return nil, err
	} else if info.RowsReturned() == 0 {
		return nil, errors.NewEntityNotFoundError(models.Permission{}, id)
	}
	return p, nil
}

// GetPermissionRequets retrieve permission requests for a service account
func (ps *permissions) GetPermissionRequests(
	saID string,
) ([]models.PermissionRequest, error) {
	prs := []models.PermissionRequest{}
	_, err := ps.storage.PG.DB.Query(
		&prs, `SELECT id, service, action, resource_hierarchy, message, state,
		created_at, updated_at FROM permissions_requests
		WHERE service_account_id = ?`, saID,
	)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (ps *permissions) ForServiceAccount(
	saID string,
) ([]models.Permission, error) {
	permissions := []models.Permission{}
	if _, err := ps.storage.PG.DB.Query(
		&permissions, `SELECT p.id, p.role_id, p.service, p.ownership_level,
p.action, p.resource_hierarchy, p.alias FROM permissions p
	JOIN role_bindings rb ON rb.role_id = p.role_id
	WHERE rb.service_account_id = ?
	ORDER BY p.service, p.ownership_level, p.action, p.resource_hierarchy`, saID,
	); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (ps *permissions) ForRole(roleID string) ([]models.Permission, error) {
	permissions := []models.Permission{}
	if _, err := ps.storage.PG.DB.Query(
		&permissions, `SELECT id, role_id, service, ownership_level,
action, resource_hierarchy, alias FROM permissions
	WHERE role_id = ?
	ORDER BY service, ownership_level, action, resource_hierarchy`, roleID,
	); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (ps *permissions) Create(p *models.Permission) error {
	_, err := ps.storage.PG.DB.Exec(
		`INSERT INTO permissions (role_id, service, ownership_level, action,
		resource_hierarchy, alias) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING RETURNING id`, p.RoleID, p.Service, p.OwnershipLevel,
		p.Action, p.ResourceHierarchy, p.Alias,
	)
	return err
}
func (ps *permissions) CreateRequest(
	saID string, r *models.PermissionRequest,
) error {
	_, err := ps.storage.PG.DB.Query(r,
		`INSERT INTO permissions_requests (service, action, resource_hierarchy,
		message, state, service_account_id) VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
		r.Service, r.Action, r.ResourceHierarchy, r.Message, r.State, saID)
	return err
}

func (ps *permissions) Delete(id string) error {
	_, err := ps.storage.PG.DB.Exec(
		`DELETE FROM permissions WHERE id = ?`, id,
	)
	return err
}

// NewPermissions users ctor
func NewPermissions(s *Storage) Permissions {
	return &permissions{&withStorage{storage: s}}
}
