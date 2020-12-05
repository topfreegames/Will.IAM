package api

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/Will.IAM/constants"
	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/oauth2"
	"github.com/topfreegames/Will.IAM/repositories"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/jaeger"
	"github.com/topfreegames/extensions/middleware"
	"github.com/topfreegames/extensions/router"
)

// App struct
type App struct {
	address         string
	config          *viper.Viper
	logger          logrus.FieldLogger
	router          *mux.Router
	server          *http.Server
	metricsReporter middleware.MetricsReporter
	storage         *repositories.Storage
	oauth2Provider  oauth2.Provider
}

// NewApp creates a new app
func NewApp(
	host string, port int, config *viper.Viper, logger logrus.FieldLogger,
	storageOrNil *repositories.Storage,
) (*App, error) {
	mr, err := middleware.NewDogStatsD(config)
	if err != nil {
		return nil, err
	}
	if storageOrNil == nil {
		storageOrNil = repositories.NewStorage()
	}
	a := &App{
		config:          config,
		address:         fmt.Sprintf("%s:%d", host, port),
		logger:          logger,
		metricsReporter: mr,
		storage:         storageOrNil,
	}
	err = a.configureApp()
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) configureApp() error {
	if err := a.configureJaeger(); err != nil {
		return err
	}
	if err := a.configurePG(); err != nil {
		return err
	}

	a.configureOAuth2Provider()
	a.configureServer()

	return nil
}

func (a *App) configureServer() {
	a.router = a.GetRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"x-access-token", "x-email"},
		AllowCredentials: false,
	})
	handler := c.Handler(a.router)
	a.server = &http.Server{
		Addr:    a.address,
		Handler: wrapHandlerWithResponseWriter(handler),
	}
}

func (a *App) configurePG() error {
	if a.storage != nil && a.storage.PG != nil {
		return nil
	}
	return a.storage.ConfigurePG(a.config)
}

func (a *App) configureJaeger() error {
	opts := jaeger.Options{
		Disabled:    a.config.GetBool("jaeger.disabled"),
		Probability: a.config.GetFloat64("jaeger.samplingProbability"),
		ServiceName: a.config.GetString("jaeger.serviceName"),
	}
	_, err := jaeger.Configure(opts)
	if err != nil {
		a.logger.WithError(err).Error("Failed to initialize Jaeger")
	}
	return err
}

func (a *App) configureOAuth2Provider() {
	repo := repositories.New(a.storage)
	provider := oauth2.GetOAuthProvider(a.config, repo)

	a.SetOAuth2Provider(provider)
}

// SetOAuth2Provider sets a provider in App
func (a *App) SetOAuth2Provider(provider oauth2.Provider) {
	a.oauth2Provider = provider
}

// GetRouter returns App's *mux.Router reference
func (a *App) GetRouter() *mux.Router {
	r := router.NewRouter()
	r.Use(middleware.Version(constants.AppInfo.Version))
	r.Use(middleware.Logging(a.logger))
	r.Use(middleware.Metrics(a.metricsReporter))

	repo := repositories.New(a.storage)

	r.HandleFunc("/healthcheck", healthcheckHandler(
		usecases.NewHealthcheck(repo),
	)).Methods("GET").Name("healthcheck")

	r.HandleFunc("/sso/auth/do",
		authenticationBuildURLHandler(a.oauth2Provider),
	).Methods("GET").Name("ssoAuthDo")

	psUC := usecases.NewPermissions(repo)
	sasUC := usecases.NewServiceAccounts(repo, a.oauth2Provider)

	r.HandleFunc("/sso/auth/done",
		authenticationExchangeCodeHandler(a.oauth2Provider, sasUC),
	).Methods("GET").Name("ssoAuthDone")

	r.HandleFunc("/sso/auth/valid",
		authenticationValidHandler(sasUC),
	).Methods("GET").Name("ssoAuthValid")

	ssUC := usecases.NewServices(repo)
	authMiddle := authMiddleware(sasUC)

	r.Handle("/sso/auth",
		authMiddle(http.HandlerFunc(authenticationHandler)),
	).Methods("GET").Name("ssoAuth")

	r.PathPrefix("/sso").Handler(http.StripPrefix("/sso", http.FileServer(
		http.Dir("./assets/sso/")),
	)).Methods("GET").Name("sso")

	r.Handle(
		"/services",
		authMiddle(http.HandlerFunc(
			servicesListHandler(ssUC),
		)),
	).
		Methods("GET").Name("servicesListHandler")

	hasPermissionMiddle := hasPermissionMiddlewareBuilder(sasUC)

	r.Handle(
		"/services/{id}",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"EditService", "{id}",
		), http.HandlerFunc(
			servicesGetHandler(ssUC),
		))),
	).
		Methods("GET").Name("servicesGetHandler")

	r.Handle(
		"/services",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"CreateServices", "*",
		), http.HandlerFunc(
			servicesCreateHandler(ssUC),
		))),
	).
		Methods("POST").Name("servicesCreateHandler")

	r.Handle(
		"/services/{id}",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"EditService", "{id}",
		), http.HandlerFunc(
			servicesUpdateHandler(ssUC),
		))),
	).
		Methods("PUT").Name("servicesUpdateHandler")

	r.Handle(
		"/service_accounts",
		authMiddle(http.HandlerFunc(serviceAccountsListHandler(sasUC))),
	).
		Methods("GET").Name("serviceAccountsListHandler")

	r.Handle(
		"/service_accounts/with_permission",
		authMiddle(http.HandlerFunc(serviceAccountsListWithPermissionHandler(sasUC))),
	).
		Methods("GET").Name("serviceAccountsListWithPermissionHandler")

	r.Handle(
		"/service_accounts/search",
		authMiddle(http.HandlerFunc(serviceAccountsSearchHandler(sasUC))),
	).
		Methods("GET").Name("serviceAccountsSearchHandler")

	r.Handle(
		"/service_accounts/{id}",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"EditServiceAccount", "{id}",
		), http.HandlerFunc(
			serviceAccountsGetHandler(sasUC),
		))),
	).
		Methods("GET").Name("serviceAccountsGetHandler")

	r.Handle(
		"/service_accounts",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"CreateServiceAccounts", "*",
		), http.HandlerFunc(
			serviceAccountsCreateHandler(sasUC),
		))),
	).
		Methods("POST").Name("serviceAccountsCreateHandler")

	r.Handle(
		"/service_accounts/{id}",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"EditServiceAccount", "{id}",
		), http.HandlerFunc(
			serviceAccountsUpdateHandler(sasUC),
		))),
	).
		Methods("PUT").Name("serviceAccountsUpdateHandler")

	// roles

	rsUC := usecases.NewRoles(repo)

	r.Handle(
		"/roles/{id}/permissions",
		authMiddle(http.HandlerFunc(
			rolesCreatePermissionHandler(sasUC, rsUC),
		)),
	).
		Methods("POST").Name("rolesCreatePermissionHandler")

	r.Handle(
		"/roles/{id}",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"EditRole", "{id}",
		), http.HandlerFunc(
			rolesUpdateHandler(sasUC, rsUC),
		))),
	).
		Methods("PUT").Name("rolesUpdateHandler")

	r.Handle(
		"/roles",
		authMiddle(http.HandlerFunc(rolesListHandler(rsUC))),
	).
		Methods("GET").Name("rolesListHandler")

	r.Handle(
		"/roles/search",
		authMiddle(http.HandlerFunc(rolesSearchHandler(rsUC))),
	).
		Methods("GET").Name("rolesSearchHandler")

	r.Handle(
		"/roles/{id}",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"EditRole", "{id}",
		), http.HandlerFunc(
			rolesGetHandler(rsUC),
		))),
	).
		Methods("GET").Name("rolesListHandler")


	r.Handle(
		"/roles/{id}/permissions",
		authMiddle(http.HandlerFunc(
			rolesGePermissions(rsUC),
		)),
	).
		Methods("GET").Name("rolesGePermissions")

	r.Handle(
		"/roles",
		authMiddle(hasPermissionMiddle(models.BuildWillIAMPermissionLender(
			"CreateRoles", "*",
		), http.HandlerFunc(
			rolesCreateHandler(sasUC, rsUC),
		))),
	).
		Methods("POST").Name("rolesCreateHandler")

	// permissions

	r.Handle(
		"/permissions/{id}",
		authMiddle(http.HandlerFunc(permissionsDeleteHandler(
			sasUC, psUC,
		))),
	).
		Methods("DELETE").Name("permissionsDeleteHandler")


	r.Handle(
		"/permissions/attribute",
		authMiddle(http.HandlerFunc(
			permissionsAttributeHandler(sasUC, psUC),
		)),
	).
		Methods("PUT").Name("permissionsAttributeHandler")

	r.Handle(
		"/permissions/attribute_to_emails",
		authMiddle(http.HandlerFunc(
			permissionsAttributeToEmailsHandler(sasUC, psUC),
		)),
	).
		Methods("PUT").Name("permissionsAttributeHandler")

	r.Handle(
		"/permissions/has",
		authMiddle(http.HandlerFunc(
			permissionsHasHandler(sasUC),
		)),
	).
		Methods("GET").Name("permissionsHasHandler")

	r.Handle(
		"/permissions/hasMany",
		authMiddle(http.HandlerFunc(
			permissionsHasManyHandler(sasUC),
		)),
	).
		Methods("POST").Name("permissionsHasManyHandler")

	// permissions requests

	prsUC := usecases.NewPermissionsRequests(repo)

	r.Handle(
		"/permissions/requests/open",
		authMiddle(http.HandlerFunc(permissionsRequestsListOpenHandler(prsUC))),
	).
		Methods("GET").Name("permissionsGetPermissionRequestsHandler")

	r.Handle(
		"/permissions/requests",
		authMiddle(http.HandlerFunc(permissionsRequestsCreateHandler(prsUC))),
	).
		Methods("POST").Name("permissionsCreatePermissionRequestHandler")

	r.Handle(
		"/permissions/requests/{id}/grant",
		authMiddle(http.HandlerFunc(permissionsRequestsGrantHandler(prsUC))),
	).
		Methods("PUT").Name("permissionsGetPermissionRequestsGrantHandler")

	r.Handle(
		"/permissions/requests/{id}/deny",
		authMiddle(http.HandlerFunc(permissionsRequestsDenyHandler(prsUC))),
	).
		Methods("PUT").Name("permissionsGetPermissionRequestsDenyHandler")

	amUseCase := usecases.NewAM(repo, rsUC)

	r.Handle(
		"/am",
		authMiddle(http.HandlerFunc(amListHandler(amUseCase))),
	).
		Methods("GET").Name("permissionsHasHandler")

	return r
}

//ListenAndServe requests
func (a *App) ListenAndServe() {
	listener, err := net.Listen("tcp", a.address)
	if err != nil {
		a.logger.WithError(err).Error("Failed to listen HTTP")
	}

	defer listener.Close()

	err = a.server.Serve(listener)
	if err != nil {
		a.logger.WithError(err).Error("Closed http listener")
	}
}
