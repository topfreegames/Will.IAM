package usecases

import (
	"context"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// Services contract
type Services interface {
	List() ([]models.Service, error)
	Get(string) (*models.Service, error)
	Create(*models.Service) error
	Update(*models.Service) error
	WithContext(context.Context) Services
}

type services struct {
	repo *repositories.All
	ctx  context.Context
}

func (ss services) WithContext(ctx context.Context) Services {
	return &services{ss.repo.WithContext(ctx), ctx}
}

// Create a new service with unique name and permission name
// Also creates an associate Service Account with full access
// and attributes full access to creator
func (ss services) Create(service *models.Service) error {
	creatorSA, err := ss.repo.ServiceAccounts.Get(service.CreatorServiceAccountID)
	if err != nil {
		return err
	}
	return ss.repo.WithPGTx(ss.ctx, func(repo *repositories.All) error {
		sa := models.BuildKeyPairServiceAccount(service.Name)
		if err := createServiceAccount(sa, repo); err != nil {
			return err
		}
		service.ServiceAccountID = sa.ID
		if err := repo.Services.Create(service); err != nil {
			return err
		}
		buildFullAccessPermissionForRoleID := func(
			roleID string,
		) *models.Permission {
			return &models.Permission{
				Service:           service.PermissionName,
				OwnershipLevel:    models.OwnershipLevels.Owner,
				Action:            models.Action("*"),
				ResourceHierarchy: models.ResourceHierarchy("*"),
				RoleID:            roleID,
			}
		}
		repo.Permissions.Create(
			buildFullAccessPermissionForRoleID(sa.BaseRoleID),
		)
		repo.Permissions.Create(
			buildFullAccessPermissionForRoleID(creatorSA.BaseRoleID),
		)
		return nil
	})
}

func (ss services) List() ([]models.Service, error) {
	return ss.repo.Services.List()
}

func (ss services) Get(id string) (*models.Service, error) {
	return ss.repo.Services.Get(id)
}

func (ss services) Update(service *models.Service) error {
	return ss.repo.Services.Update(service)
}

// NewServices services' ctor
func NewServices(repo *repositories.All) Services {
	return &services{repo: repo}
}
