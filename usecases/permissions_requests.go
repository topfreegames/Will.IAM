package usecases

import (
	"context"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
)

// PermissionsRequests define entrypoints for PermissionsRequests actions
type PermissionsRequests interface {
	Create(*models.PermissionRequest) error
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

func (prs permissionsRequests) Create(pr *models.PermissionRequest) error {
	pr.State = models.PermissionRequestStates.Open
	return prs.repo.PermissionsRequests.Create(pr)
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
