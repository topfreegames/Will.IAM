package models

// PermissionRequest type
type PermissionRequest struct {
	ID                string                 `json:"id" pg:"id"`
	OwnershipLevel    OwnershipLevel         `json:"ownershipLevel" pg:"ownership_level"`
	Service           string                 `json:"service" pg:"service"`
	Action            Action                 `json:"action" pg:"action"`
	ResourceHierarchy ResourceHierarchy      `json:"resourceHierarchy" pg:"resource_hierarchy"`
	Message           string                 `json:"message" pg:"message"`
	State             PermissionRequestState `json:"state" pg:"state"`
	ServiceAccountID  string                 `json:"serviceAccountId" pg:"service_account_id"`
	CreatedUpdatedAt
}

// Permission returns requested Permission from pr
func (pr PermissionRequest) Permission() Permission {
	return Permission{
		Service:           pr.Service,
		Action:            pr.Action,
		OwnershipLevel:    pr.OwnershipLevel,
		ResourceHierarchy: pr.ResourceHierarchy,
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
