package repositories

import (
	"fmt"

	"github.com/ghostec/Will.IAM/errors"
	"github.com/ghostec/Will.IAM/models"
)

// Roles repository
type Roles interface {
	GetServiceAccounts(string) ([]models.ServiceAccount, error)
	ForServiceAccountID(string) ([]models.Role, error)
	Create(*models.Role) error
	Update(*models.Role) error
	Bind(*models.RoleBinding) error
	WithNamePrefix(string, int) ([]models.Role, error)
	List(*ListOptions) ([]models.Role, error)
	ListCount() (int64, error)
	Search(string, *ListOptions) ([]models.Role, error)
	SearchCount(string) (int64, error)
	Get(string) (*models.Role, error)
	DropPermissions(string) error
	DropBindings(string) error
	Clone() Roles
	setStorage(*Storage)
}

type roles struct {
	*withStorage
}

func (rs *roles) Clone() Roles {
	return NewRoles(rs.storage.Clone())
}

func (rs roles) GetServiceAccounts(
	roleID string,
) ([]models.ServiceAccount, error) {
	sas := []models.ServiceAccount{}
	_, err := rs.storage.PG.DB.Query(
		&sas,
		`SELECT sa.id, sa.name, sa.picture, sa.email FROM service_accounts sa
		JOIN role_bindings rb ON rb.service_account_id = sa.id
		WHERE rb.role_id = ?
		ORDER BY sa.created_at DESC`,
		roleID,
	)
	if err != nil {
		return nil, err
	}
	return sas, nil
}

func (rs roles) ForServiceAccountID(serviceAccountID string) ([]models.Role, error) {
	var roles []models.Role
	_, err := rs.storage.PG.DB.Query(
		&roles,
		`SELECT r.id, r.name, r.is_base_role FROM roles r
		JOIN role_bindings rb ON rb.role_id = r.id
		WHERE rb.service_account_id = ?`,
		serviceAccountID,
	)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (rs roles) Create(r *models.Role) error {
	_, err := rs.storage.PG.DB.Query(
		r, `INSERT INTO roles (name, is_base_role) VALUES (?name, ?is_base_role)
		RETURNING id`, r,
	)
	return err
}

func (rs roles) Update(r *models.Role) error {
	// tx, err := rs.storage.PG.Begin(rs.storage.PG.DB)
	_, err := rs.storage.PG.DB.Query(
		r, `UPDATE roles SET name = ?name WHERE id = ?id`, r,
	)
	return err
}

func (rs roles) Bind(rb *models.RoleBinding) error {
	_, err := rs.storage.PG.DB.Exec(
		`INSERT INTO role_bindings (role_id, service_account_id)
		VALUES (?role_id, ?service_account_id)`, rb,
	)
	return err
}

func (rs roles) WithNamePrefix(
	prefix string, maxResults int,
) ([]models.Role, error) {
	var rsSl []models.Role
	if _, err := rs.storage.PG.DB.Query(
		&rsSl, "SELECT id, name FROM roles WHERE name ILIKE ?",
		fmt.Sprintf("%s%%", prefix),
	); err != nil {
		return nil, err
	}
	return rsSl, nil
}

func (rs roles) List(lo *ListOptions) ([]models.Role, error) {
	var rsSl []models.Role
	if _, err := rs.storage.PG.DB.Query(
		&rsSl, `SELECT id, name FROM roles WHERE is_base_role = false
		ORDER BY name ASC LIMIT ? OFFSET ?`, lo.Limit(), lo.Offset(),
	); err != nil {
		return nil, err
	}
	return rsSl, nil
}

func (rs roles) ListCount() (int64, error) {
	var count int64
	if _, err := rs.storage.PG.DB.Query(
		&count, `SELECT count(*) FROM roles WHERE is_base_role = false`,
	); err != nil {
		return 0, err
	}
	return count, nil
}

func (rs roles) Search(term string, lo *ListOptions) ([]models.Role, error) {
	var rsSl []models.Role
	if _, err := rs.storage.PG.DB.Query(
		&rsSl, `SELECT id, name FROM roles WHERE name ILIKE ?
		AND is_base_role = false ORDER BY name ASC LIMIT ? OFFSET ?`,
		fmt.Sprintf("%%%s%%", term),
		lo.Limit(), lo.Offset(),
	); err != nil {
		return nil, err
	}
	return rsSl, nil
}

func (rs roles) SearchCount(term string) (int64, error) {
	var count int64
	if _, err := rs.storage.PG.DB.Query(
		&count, `SELECT count(*) FROM roles WHERE name ILIKE ?
		AND is_base_role = false`,
		fmt.Sprintf("%%%s%%", term),
	); err != nil {
		return 0, err
	}
	return count, nil
}

func (rs roles) Get(id string) (*models.Role, error) {
	r := new(models.Role)
	if _, err := rs.storage.PG.DB.Query(
		r, `SELECT id, name, is_base_role, created_at, updated_at
		FROM roles WHERE id = ?`, id,
	); err != nil {
		return nil, err
	}
	if r.ID == "" {
		return nil, errors.NewEntityNotFoundError(models.Role{}, id)
	}
	return r, nil
}

func (rs roles) DropPermissions(roleID string) error {
	_, err := rs.storage.PG.DB.Exec(
		`DELETE FROM permissions WHERE role_id = ?`, roleID,
	)
	return err
}

func (rs roles) DropBindings(roleID string) error {
	_, err := rs.storage.PG.DB.Exec(
		`DELETE FROM role_bindings WHERE role_id = ?`, roleID,
	)
	return err
}

// NewRoles roles ctor
func NewRoles(s *Storage) Roles {
	return &roles{&withStorage{storage: s}}
}
