package usecases

import (
	"context"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// Permissions define entrypoints for Permissions actions
type Permissions interface {
	Get(string) (*models.Permission, error)
	Delete(string) error
	Create(*models.Permission) error
	CreateRequest(string, *models.PermissionRequest) error
	GetPermissionRequests(string) ([]models.PermissionRequest, error)
	Attribute(*PermissionsAttribute) error
	AttributeToEmails(*PermissionsAttributeToEmails) error
	WithContext(context.Context) Permissions
}

// PermissionsAttribute are used in PUT /permissions/attribute
type PermissionsAttribute struct {
	RolesIDs           []string            `json:"rolesIds"`
	PermissionsStrings []string            `json:"permissions"`
	PermissionsAliases map[string]string   `json:"permissionsAliases"`
	Permissions        []models.Permission `json:"-"`
}

// PermissionsAttributeToEmails are used in PUT /permissions/attribute_to_emails
type PermissionsAttributeToEmails struct {
	Emails             []string            `json:"emails"`
	PermissionsStrings []string            `json:"permissions"`
	PermissionsAliases map[string]string   `json:"permissionsAliases"`
	Permissions        []models.Permission `json:"-"`
}

type permissions struct {
	repo *repositories.All
	ctx  context.Context
}

func (ps permissions) WithContext(ctx context.Context) Permissions {
	return &permissions{ps.repo.WithContext(ctx), ctx}
}

func (ps permissions) Get(id string) (*models.Permission, error) {
	return ps.repo.Permissions.Get(id)
}

func (ps permissions) Delete(id string) error {
	return ps.repo.Permissions.Delete(id)
}

func (ps permissions) CreateRequest(
	saID string, r *models.PermissionRequest,
) error {
	r.State = models.PermissionRequestStates.Created
	return ps.repo.Permissions.CreateRequest(saID, r)
}

func (ps permissions) Create(p *models.Permission) error {
	return ps.repo.Permissions.Create(p)
}

func (ps permissions) GetPermissionRequests(
	saID string,
) ([]models.PermissionRequest, error) {
	return ps.repo.Permissions.GetPermissionRequests(saID)
}

func (ps permissions) Attribute(pa *PermissionsAttribute) error {
	return ps.repo.WithPGTx(ps.ctx, func(repo *repositories.All) error {
		for _, roleID := range pa.RolesIDs {
			for _, permission := range pa.Permissions {
				permission.RoleID = roleID
				if err := repo.Permissions.Create(&permission); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (ps permissions) AttributeToEmails(pa *PermissionsAttributeToEmails) error {
	sas, err := ps.repo.ServiceAccounts.ForEmails(pa.Emails)
	if err != nil {
		return err
	}
	return ps.repo.WithPGTx(ps.ctx, func(repo *repositories.All) error {
		for _, sa := range sas {
			for _, permission := range pa.Permissions {
				permission.RoleID = sa.BaseRoleID
				if err := repo.Permissions.Create(&permission); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// NewPermissions ctor
func NewPermissions(repo *repositories.All) Permissions {
	return &permissions{repo: repo}
}
