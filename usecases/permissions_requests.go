package usecases

import (
	"context"
	"fmt"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// PermissionsRequests define entrypoints for PermissionsRequests actions
type PermissionsRequests interface {
	Create(*models.PermissionRequest) error
	Deny(string, string) error
	Grant(string, string) error
	ListOpenRequestsVisibleTo(
		*repositories.ListOptions, string,
	) ([]models.PermissionRequest, int64, error)
	WithContext(context.Context) PermissionsRequests
}

type permissionsRequests struct {
	repo *repositories.All
	ctx  context.Context
}

func (prs permissionsRequests) WithContext(ctx context.Context) PermissionsRequests {
	return &permissionsRequests{prs.repo.WithContext(ctx), ctx}
}

// Create checks if pr.saID has open request OR has permission, if not it opens a permission request
func (prs permissionsRequests) Create(pr *models.PermissionRequest) error {
	return prs.repo.WithPGTx(prs.ctx, func(repo *repositories.All) error {
		pr.State = models.PermissionRequestStates.Open
		has, err := repo.ServiceAccounts.HasPermission(pr.ServiceAccountID, pr.Permission())
		if err != nil {
			return err
		}
		if has {
			// TODO: replace by proper error
			return fmt.Errorf("user already has requested permission")
		}
		return repo.PermissionsRequests.Create(pr)
	})
}

// Deny will check if saID (moderator_service_account_id) is owner of the permission
// requested in prID, and if so will DENY it to the pr.ServiceAccountID base role
func (prs permissionsRequests) Deny(saID, prID string) error {
	return prs.repo.WithPGTx(prs.ctx, func(repo *repositories.All) error {
		pr, err := repo.PermissionsRequests.Get(prID)
		if err != nil {
			return err
		}
		if pr.State != models.PermissionRequestStates.Open {
			// TODO: replace by proper error
			return fmt.Errorf("permission request is closed")
		}
		ownerPermission := pr.Permission()
		ownerPermission.OwnershipLevel = models.OwnershipLevels.Owner
		has, err := repo.ServiceAccounts.HasPermission(saID, ownerPermission)
		if err != nil {
			return err
		}
		if !has {
			// TODO: replace by proper error
			return fmt.Errorf("user isn't owner of permission")
		}
		return repo.PermissionsRequests.Deny(saID, prID)
	})
}

// Grant will check if saID (moderator_service_account_id) is owner of the permission
// requested in prID, and if so will GRANT it to the pr.ServiceAccountID base role
func (prs permissionsRequests) Grant(saID, prID string) error {
	return prs.repo.WithPGTx(prs.ctx, func(repo *repositories.All) error {
		pr, err := repo.PermissionsRequests.Get(prID)
		if err != nil {
			return err
		}
		if pr.State != models.PermissionRequestStates.Open {
			// TODO: replace by proper error
			return fmt.Errorf("permission request is closed")
		}
		ownerPermission := pr.Permission()
		ownerPermission.OwnershipLevel = models.OwnershipLevels.Owner
		has, err := repo.ServiceAccounts.HasPermission(saID, ownerPermission)
		if err != nil {
			return err
		}
		if !has {
			// TODO: replace by proper error
			return fmt.Errorf("user isn't owner of permission")
		}
		p := pr.Permission()
		if err := createPermissionForServiceAccount(repo, pr.ServiceAccountID, &p); err != nil {
			return err
		}
		return repo.PermissionsRequests.Grant(saID, prID)
	})
}

func (prs permissionsRequests) ListOpenRequestsVisibleTo(
	lo *repositories.ListOptions, saID string,
) ([]models.PermissionRequest, int64, error) {
	ors, err := prs.repo.PermissionsRequests.ListOpenRequestsVisibleTo(lo, saID)
	if err != nil {
		return nil, 0, err
	}
	count, err := prs.repo.PermissionsRequests.ListOpenRequestsVisibleToCount(saID)
	if err != nil {
		return nil, 0, err
	}
	return ors, count, nil
}

// NewPermissionsRequests ctor
func NewPermissionsRequests(repo *repositories.All) PermissionsRequests {
	return &permissionsRequests{repo: repo}
}
