package models

// PermissionRequest type
type PermissionRequest struct {
	ID                        string                 `json:"id" pg:"id"`
	OwnershipLevel            OwnershipLevel         `json:"ownershipLevel" pg:"ownership_level"`
	Service                   string                 `json:"service" pg:"service"`
	Action                    Action                 `json:"action" pg:"action"`
	ResourceHierarchy         ResourceHierarchy      `json:"resourceHierarchy" pg:"resource_hierarchy"`
	Alias                     string                 `json:"alias" pg:"alias"`
	Message                   string                 `json:"message" pg:"message"`
	State                     PermissionRequestState `json:"state" pg:"state"`
	ServiceAccountID          string                 `json:"serviceAccountId" pg:"service_account_id"`
	RequesterPicture          string                 `json:"requesterPicture" pg:"requester_picture"`
	RequesterName             string                 `json:"requesterName" pg:"requester_name"`
	ModeratorServiceAccountID string                 `json:"moderatorServiceAccountId" pg:"moderator_service_account_id"`
	CreatedUpdatedAt
}

// Permission returns requested Permission from pr
func (pr PermissionRequest) Permission() Permission {
	return Permission{
		Service:           pr.Service,
		Action:            pr.Action,
		OwnershipLevel:    pr.OwnershipLevel,
		ResourceHierarchy: pr.ResourceHierarchy,
		Alias:             pr.Alias,
	}
}

// PermissionRequestState type
type PermissionRequestState string

// PermissionRequestStates possible
var PermissionRequestStates = struct {
	Open    PermissionRequestState
	Granted PermissionRequestState
	Denied  PermissionRequestState
}{
	Open:    "open",
	Granted: "granted",
	Denied:  "denied",
}

// String returns permission request state as string
func (prs PermissionRequestState) String() string {
	return string(prs)
}
