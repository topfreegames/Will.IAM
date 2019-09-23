package models

import "github.com/gofrs/uuid"

// ServiceAccount type
type ServiceAccount struct {
	ID                 string             `json:"id" pg:"id"`
	Name               string             `json:"name" pg:"name"`
	KeyID              string             `json:"keyId" pg:"key_id"`
	KeySecret          string             `json:"keySecret" pg:"key_secret"`
	Email              string             `json:"email" pg:"email"`
	Picture            string             `json:"picture" pg:"picture"`
	BaseRoleID         string             `json:"baseRoleId" pg:"base_role_id"`
	AuthenticationType AuthenticationType `json:"authenticationType" pg:"-"`
	CreatedUpdatedAt
}

// AuthenticationType type
type AuthenticationType string

// AuthenticationTypes are the supported authentication types
var AuthenticationTypes = struct {
	OAuth2  AuthenticationType
	KeyPair AuthenticationType
	Unknown AuthenticationType
}{
	OAuth2:  "oauth2",
	KeyPair: "keypair",
	Unknown: "unknown",
}

func (authType AuthenticationType) String() string {
	return string(authType)
}

// Valid checks if at is a possible value
func (authType AuthenticationType) Valid() bool {
	if authType.String() == AuthenticationTypes.OAuth2.String() || authType.String() == AuthenticationTypes.KeyPair.String() {
		return true
	}
	return false
}

// BuildKeyPairServiceAccount generates random KeyID and KeySecret
func BuildKeyPairServiceAccount(name string) *ServiceAccount {
	return &ServiceAccount{
		Name:      name,
		KeyID:     uuid.Must(uuid.NewV4()).String(),
		KeySecret: uuid.Must(uuid.NewV4()).String(),
	}
}

// BuildOAuth2ServiceAccount generates a ServiceAccount with Name and Email
func BuildOAuth2ServiceAccount(name, email string) *ServiceAccount {
	return &ServiceAccount{
		Name:  name,
		Email: email,
	}
}
