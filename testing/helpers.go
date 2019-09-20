package testing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/Will.IAM/api"
	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/oauth2"
	"github.com/topfreegames/Will.IAM/repositories"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/Will.IAM/utils"
)

// GetConfig gets config for tests
func GetConfig(t *testing.T, path ...string) *viper.Viper {
	t.Helper()
	filePath := "./../testing/config.yaml"
	if len(path) > 0 {
		filePath = path[0]
	}
	config, err := utils.GetConfig(filePath)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

// GetLogger gets config for tests
func GetLogger(t *testing.T) logrus.FieldLogger {
	t.Helper()
	return utils.GetLogger("0.0.0.0", 4040, 0, true)
}

// GetApp is a helper to create an *api.App
func GetApp(t *testing.T) *api.App {
	app, err := api.NewApp("0.0.0.0", 4040, GetConfig(t), GetLogger(t), nil)
	app.SetOAuth2Provider(oauth2.NewProviderBlankMock())
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return app
}

// DoRequest executes req over handler and returns a recorder
func DoRequest(
	t *testing.T, req *http.Request, handler http.Handler,
) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// GetStorage returns a *repositories.Storage
func GetStorage(t *testing.T) *repositories.Storage {
	t.Helper()
	s := repositories.NewStorage()
	if err := s.ConfigurePG(GetConfig(t)); err != nil {
		panic(err)
	}
	return s
}

// GetRepo return an instance of *repositories.All
func GetRepo(t *testing.T) *repositories.All {
	t.Helper()
	return repositories.New(GetStorage(t))
}

// GetRolesUseCase returns a usecases.Roles
func GetRolesUseCase(t *testing.T) usecases.Roles {
	t.Helper()
	return usecases.NewRoles(GetRepo(t)).WithContext(context.Background())
}

// GetServiceAccountsUseCase returns a usecases.ServiceAccounts
func GetServiceAccountsUseCase(t *testing.T) usecases.ServiceAccounts {
	t.Helper()
	repo := GetRepo(t)
	providerBlankMock := oauth2.NewProviderBlankMock()
	return usecases.NewServiceAccounts(repo, providerBlankMock).
		WithContext(context.Background())
}

// GetServicesUseCase returns a usecases.Services
func GetServicesUseCase(t *testing.T) usecases.Services {
	t.Helper()
	return usecases.NewServices(GetRepo(t)).WithContext(context.Background())
}

// GetPermissionsRequestsUseCase returns a usecases.PermissionsRequests
func GetPermissionsRequestsUseCase(t *testing.T) usecases.PermissionsRequests {
	t.Helper()
	return usecases.NewPermissionsRequests(GetRepo(t)).WithContext(context.Background())
}

// CreateRootServiceAccountWithKeyPair creates a root service account with root access using KeyPair
func CreateRootServiceAccountWithKeyPair(t *testing.T) *models.ServiceAccount {
	t.Helper()
	return CreateServiceAccountWithPermissions(t, "root", "root@test.com", models.AuthenticationTypes.KeyPair, "*::RO::*::*")
}

// CreateRootServiceAccountWithOAuth creates a root service account with root access using OAuth
func CreateRootServiceAccountWithOAuth(t *testing.T) *models.ServiceAccount {
	t.Helper()

	serviceAccount := CreateServiceAccountWithPermissions(t, "root", "root@test.com", models.AuthenticationTypes.OAuth2, "*::RO::*::*")
	token := &models.Token{
		AccessToken:  uuid.Must(uuid.NewV4()).String(),
		RefreshToken: uuid.Must(uuid.NewV4()).String(),
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour * 24 * 10),
		Email:        serviceAccount.Email,
	}

	repo := GetRepo(t)
	repo.Tokens.Save(token)

	return serviceAccount
}

// CreateServiceAccountWithPermissions create an account with a list of permissions
func CreateServiceAccountWithPermissions(t *testing.T, name string, email string, method models.AuthenticationType, permissions ...string) *models.ServiceAccount {
	saUC := GetServiceAccountsUseCase(t)

	var rootSA *models.ServiceAccount
	var err error

	if method == models.AuthenticationTypes.KeyPair {
		rootSA, err = saUC.CreateKeyPairType(name)
	} else if method == models.AuthenticationTypes.OAuth2 {
		rootSA, err = saUC.CreateOAuth2Type(name, email)
	}

	if err != nil {
		panic(err)
	}

	for _, permission := range permissions {
		p, err := models.BuildPermission(permission)
		if err != nil {
			panic(err)
		}
		err = saUC.CreatePermission(rootSA.ID, &p)
		if err != nil {
			panic(err)
		}
	}
	return rootSA
}

// CleanupPG clears the databse data between tests
func CleanupPG(t *testing.T) {
	t.Helper()
	storage := GetStorage(t)
	rels := []string{
		"permissions_requests",
		"permissions",
		"role_bindings",
		"roles",
		"service_accounts",
		"services",
	}
	for _, rel := range rels {
		if _, err := storage.PG.DB.Exec(fmt.Sprintf("DELETE FROM %s;", rel)); err != nil {
			panic(err)
		}
	}
}
