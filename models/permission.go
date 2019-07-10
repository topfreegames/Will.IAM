package models

import (
	"fmt"
	"strings"

	"github.com/topfreegames/Will.IAM/constants"
)

// OwnershipLevel defines the holding rights about a resource
type OwnershipLevel string

// OwnershipLevels are all possible holding rights
// Owner: can exercise the action over the resource and provide the exact same
// ownership:action:resource rights to other parties
// Lender: can only exercise the action over the resource
var OwnershipLevels = struct {
	Owner  OwnershipLevel
	Lender OwnershipLevel
}{
	Owner:  "RO",
	Lender: "RL",
}

// Less returns true if o < oo; Lender < Owner
func (o OwnershipLevel) Less(oo OwnershipLevel) bool {
	if o == OwnershipLevels.Lender && oo == OwnershipLevels.Owner {
		return true
	}
	return false
}

// String returns string fmt
func (o OwnershipLevel) String() string {
	return string(o)
}

// Action is defined by IAM clients
type Action string

// BuildAction from string
func BuildAction(str string) Action {
	return Action(str)
}

// All checks if action matches any
func (a Action) All() bool {
	return a.String() == "*"
}

// String returns string representation
func (a Action) String() string {
	return string(a)
}

// ResourceHierarchy is either a complete or an open hierarchy to something
// Eg:
// Complete: maestro::sniper-3d::na::sniper3d-red
// Open: maestro::sniper-3d::stag::*
type ResourceHierarchy string

// BuildResourceHierarchy returns an instance of ResourceHierarchy from a string
func BuildResourceHierarchy(rh string) ResourceHierarchy {
	return ResourceHierarchy(rh)
}

// All checks if rh matches any
func (rh ResourceHierarchy) All() bool {
	return string(rh) == "*"
}

// PermissionMatches returns a slice of all possible ways to check if a service account has
// any level of permission over this ResourceHierarchy
// Example: rh = "x::y::z".PermissionMatches() = ["*", "x::*", "x::y::*", "x::y::z"]
func (rh ResourceHierarchy) PermissionMatches() []string {
	whole := string(rh)
	if whole == "*" {
		return []string{whole}
	}
	parts := strings.Split(whole, "::")
	ret := []string{"*"}
	for i := 1; i < len(parts); i++ {
		if parts[i] == "*" {
			break
		}
		match := strings.Join(parts[:i], "::")
		ret = append(ret, match+"::*")
	}
	return append(ret, whole)
}

type resourceHierarchy struct {
	size      int
	hierarchy []string
}

func buildResourceHierarchy(rh ResourceHierarchy) resourceHierarchy {
	parts := strings.Split(string(rh), "::")
	size := len(parts)
	return resourceHierarchy{size: size, hierarchy: parts}
}

// Contains checks whether rh.Hierarchy contains orh.Hierarchy
func (rh ResourceHierarchy) Contains(orh ResourceHierarchy) bool {
	rhh := buildResourceHierarchy(rh)
	orhh := buildResourceHierarchy(orh)
	if rhh.size > orhh.size {
		return false
	}
	for i := range orhh.hierarchy {
		if i >= len(rhh.hierarchy) {
			return false
		}
		if rhh.hierarchy[i] == "*" {
			return true
		}
		if orhh.hierarchy[i] != rhh.hierarchy[i] {
			return false
		}
	}
	return true
}

// String returns string representation
func (rh ResourceHierarchy) String() string {
	return string(rh)
}

// Permission is bound to a role and
// defines the onwership level of an action over a resource
type Permission struct {
	ID                string            `json:"id" pg:"id"`
	RoleID            string            `json:"roleId" pg:"role_id"`
	Service           string            `json:"service" pg:"service"`
	OwnershipLevel    OwnershipLevel    `json:"ownershipLevel" pg:"ownership_level"`
	Action            Action            `json:"action" pg:"action"`
	ResourceHierarchy ResourceHierarchy `json:"resourceHierarchy" pg:"resource_hierarchy"`
	Alias             string            `json:"alias" pg:"alias"`
}

// ValidatePermission validates a permission in string format
func ValidatePermission(str string) (bool, error) {
	// Format: OwnershipLevel::Action::Service::{ResourceHierarchy}
	parts := strings.Split(str, "::")
	if len(parts) < 4 {
		return false, fmt.Errorf(
			"Incomplete permission. Expected format: " +
				"Service::OwnershipLevel::Action::{ResourceHierarchy}",
		)
	}
	ol := OwnershipLevel(parts[1])
	if ol != OwnershipLevels.Owner && ol != OwnershipLevels.Lender {
		return false, fmt.Errorf("OwnershipLevel needs to be RO or RL")
	}
	for _, part := range parts {
		if part == "" {
			return false, fmt.Errorf("No parts can be empty")
		}
	}
	return true, nil
}

// BuildPermission transforms a string into struct
func BuildPermission(str string) (Permission, error) {
	if valid, err := ValidatePermission(str); !valid {
		return Permission{}, err
	}
	parts := strings.Split(str, "::")
	service := parts[0]
	ol := OwnershipLevel(parts[1])
	action := BuildAction(parts[2])
	rh := ResourceHierarchy(strings.Join(parts[3:], "::"))
	return Permission{
		Service:           service,
		OwnershipLevel:    ol,
		Action:            action,
		ResourceHierarchy: rh,
	}, nil
}

// BuildPermissions calls BuildPermission for all strings
func BuildPermissions(ps []string) ([]Permission, error) {
	pSl := make([]Permission, len(ps))
	var err error
	for i := range ps {
		pSl[i], err = BuildPermission(ps[i])
		if err != nil {
			return nil, err
		}
	}
	return pSl, nil
}

// IsPresent checks if a permission is satisfied in a slice
func (p Permission) IsPresent(permissions []Permission) bool {
	for _, pp := range permissions {
		if (pp.Service != "*" && pp.Service != p.Service) ||
			(pp.Action != "*" && pp.Action != p.Action) ||
			pp.OwnershipLevel.Less(p.OwnershipLevel) {
			continue
		}
		if pp.ResourceHierarchy.Contains(p.ResourceHierarchy) {
			return true
		}
	}
	return false
}

// String converts a permission to it's equivalent string format
func (p Permission) String() string {
	return fmt.Sprintf(
		"%s::%s::%s::%s", p.Service, p.OwnershipLevel, p.Action,
		string(p.ResourceHierarchy),
	)
}

// HasServiceFullAccess checks if permission allows it's role
// to execute any action over any resourch hierarchy under it's service
func (p Permission) HasServiceFullAccess() bool {
	if !p.Action.All() {
		return false
	}
	return p.ResourceHierarchy.All()
}

// HasServiceFullOwnership checks if permission allows it's role
// to execute any action over any resourch hierarchy under it's service
// and it's RO
func (p Permission) HasServiceFullOwnership() bool {
	return p.HasServiceFullAccess() && p.OwnershipLevel == OwnershipLevels.Owner
}

// buildWillIAMPermission builds a permission in the format expected
// by WillIAM handlers
func buildWillIAMPermission(ro OwnershipLevel, action, rh string) string {
	return fmt.Sprintf(
		"%s::%s::%s::%s", constants.AppInfo.Name, string(ro), action, rh,
	)
}

// BuildWillIAMPermissionLender builds a permission in the format expected
// by WillIAM handlers
func BuildWillIAMPermissionLender(action, rh string) string {
	return buildWillIAMPermission(OwnershipLevels.Lender, action, rh)
}

// BuildWillIAMPermissionOwner builds a permission in the format expected
// by WillIAM handlers
func BuildWillIAMPermissionOwner(action, rh string) string {
	return buildWillIAMPermission(OwnershipLevels.Owner, action, rh)
}

// PermissionRequest type
type PermissionRequest struct {
	ID                string                 `json:"id" pg:"id"`
	Service           string                 `json:"service" pg:"service"`
	Action            Action                 `json:"action" pg:"action"`
	ResourceHierarchy ResourceHierarchy      `json:"resourceHierarchy" pg:"resource_hierarchy"`
	Message           string                 `json:"message" pg:"message"`
	State             PermissionRequestState `json:"state" pg:"state"`
	CreatedUpdatedAt
}

// ToLenderString returns a permission formatted string with lender
// ownership level
func (r PermissionRequest) ToLenderString() string {
	return fmt.Sprintf("%s::%s::%s::%s", r.Service,
		OwnershipLevels.Lender.String(), r.Action.String(),
		r.ResourceHierarchy.String())
}

// PermissionRequestState can be: created:0 granted:1 denied:2
type PermissionRequestState int

// PermissionRequestStates possible
var PermissionRequestStates = struct {
	Created PermissionRequestState
	Granted PermissionRequestState
	Denied  PermissionRequestState
}{
	Created: 0,
	Granted: 1,
	Denied:  2,
}

// String returns permission request state as string
func (prs PermissionRequestState) String() string {
	i := int(prs)
	switch i {
	case 0:
		return "created"
	case 1:
		return "granted"
	case 2:
		return "denied"
	default:
		return "unknown"
	}
}
