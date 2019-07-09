package testing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

// CreateRootServiceAccount creates a root service account with root access
func CreateRootServiceAccount(t *testing.T) *models.ServiceAccount {
	saUC := GetServiceAccountsUseCase(t)
	rootSA, err := saUC.CreateKeyPairType("root")
	if err != nil {
		panic(err)
	}
	p, err := models.BuildPermission("*::RO::*::*")
	if err != nil {
		panic(err)
	}
	err = saUC.CreatePermission(rootSA.ID, &p)
	if err != nil {
		panic(err)
	}
	return rootSA
}

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
