package repositories

import "github.com/topfreegames/Will.IAM/models"

// PermissionsRequests repository
type PermissionsRequests interface {
	Clone() PermissionsRequests
	Create(*models.PermissionRequest) error
	Deny(string, string) error
	Get(string) (*models.PermissionRequest, error)
	Grant(string, string) error
	ListOpenRequestsVisibleTo(*ListOptions, string) ([]models.PermissionRequest, error)
	ListOpenRequestsVisibleToCount(string) (int64, error)
	setStorage(*Storage)
}

type permissionsRequests struct {
	*withStorage
}

func (prs *permissionsRequests) Clone() PermissionsRequests {
	return NewPermissionsRequests(prs.storage.Clone())
}

func (prs *permissionsRequests) Create(pr *models.PermissionRequest) error {
	_, err := prs.storage.PG.DB.Query(
		pr, `INSERT INTO permissions_requests (service, ownership_level, action, resource_hierarchy,
    message, state, service_account_id) VALUES (?service, ?ownership_level, ?action,
    ?resource_hierarchy, ?message, ?state, ?service_account_id)
    ON CONFLICT (service, ownership_level, action, resource_hierarchy, service_account_id)
    WHERE state = 'open' DO NOTHING RETURNING id`, pr,
	)
	return err
}

func (prs *permissionsRequests) Deny(saID, prID string) error {
	_, err := prs.storage.PG.DB.Exec(
		`UPDATE permissions_requests SET state = ?, moderator_service_account_id = ?, updated_at = now()
    WHERE id = ?`, models.PermissionRequestStates.Denied, saID, prID,
	)
	return err
}

func (prs *permissionsRequests) Get(prID string) (*models.PermissionRequest, error) {
	var pr models.PermissionRequest
	if _, err := prs.storage.PG.DB.Query(
		&pr, ` SELECT * FROM permissions_requests WHERE id = ? `, prID,
	); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (prs *permissionsRequests) Grant(saID, prID string) error {
	_, err := prs.storage.PG.DB.Exec(
		`UPDATE permissions_requests SET state = ?, moderator_service_account_id = ?, updated_at = now()
    WHERE id = ?`, models.PermissionRequestStates.Granted, saID, prID,
	)
	return err
}

func (prs *permissionsRequests) ListOpenRequestsVisibleTo(
	lo *ListOptions, saID string,
) ([]models.PermissionRequest, error) {
	var prSl []models.PermissionRequest
	if _, err := prs.storage.PG.DB.Query(
		// TODO: this query can probably be optimized if we group saop by service, action, rh and try to
		// keep only higher scopes
		// eg: service::action::*, service::action::x::*, ... => service::action::x
		&prSl, `
    SELECT DISTINCT pr.id, pr.service, pr.ownership_level, pr.action, pr.resource_hierarchy,
    pr.service_account_id, sas.picture AS requester_picture, sas.name AS requester_name, pr.state,
    pr.message
    FROM permissions_requests pr
    CROSS JOIN (SELECT service, action, resource_hierarchy FROM permissions
        WHERE role_id = ANY (SELECT role_id FROM role_bindings WHERE service_account_id = ?)
        AND ownership_level = 'RO') saop
    INNER JOIN service_accounts sas ON sas.id = pr.service_account_id
    WHERE state = 'open'
      AND CASE WHEN saop.service = '*' THEN true ELSE pr.service = saop.service END
      AND CASE WHEN saop.action = '*' THEN true ELSE pr.action = saop.action END
      AND CASE WHEN saop.resource_hierarchy = '*'
        THEN true
        ELSE pr.resource_hierarchy LIKE CONCAT(REPLACE(saop.resource_hierarchy, '*', ''), '%')
      END
    ORDER BY pr.service, pr.action, pr.resource_hierarchy ASC LIMIT ? OFFSET ?
    `, saID, lo.Limit(), lo.Offset(),
	); err != nil {
		return nil, err
	}
	return prSl, nil
}

func (prs *permissionsRequests) ListOpenRequestsVisibleToCount(
	saID string,
) (int64, error) {
	var count int64
	if _, err := prs.storage.PG.DB.Query(
		&count, `
    SELECT COUNT(DISTINCT pr.id) FROM permissions_requests pr
    CROSS JOIN (SELECT service, action, resource_hierarchy FROM permissions
        WHERE role_id = ANY (SELECT role_id FROM role_bindings WHERE service_account_id = ?)
        AND ownership_level = 'RO') saop
    WHERE state = 'open'
      AND CASE WHEN saop.service = '*' THEN true ELSE pr.service = saop.service END
      AND CASE WHEN saop.action = '*' THEN true ELSE pr.action = saop.action END
      AND CASE WHEN saop.resource_hierarchy = '*'
        THEN true
        ELSE pr.resource_hierarchy LIKE CONCAT(REPLACE(saop.resource_hierarchy, '*', ''), '%')
      END
    `, saID,
	); err != nil {
		return 0, err
	}
	return count, nil
}

// NewPermissionsRequests users ctor
func NewPermissionsRequests(s *Storage) PermissionsRequests {
	return &permissionsRequests{&withStorage{storage: s}}
}
