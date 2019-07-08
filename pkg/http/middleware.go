package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	ehttp "github.com/topfreegames/extensions/http"
)

// Middleware calls Will.IAM before the execution of every route
// that contains this middleware.
type Middleware struct {
	logger     logrus.FieldLogger
	permission *permission
	resource   func(*http.Request) string
	iamURL     string
	enabled    bool

	next http.Handler
}

var client *http.Client

// NewMiddleware returns a callback that returns *Middleware,
// which implements http.Handler.
// Permission is used to build the Will.IAM permission string.
func NewMiddleware(
	logger logrus.FieldLogger,
	cnf *config,
	ownershipLevel, action string,
	resource func(*http.Request) string,
) func(http.Handler) http.Handler {
	if client == nil {
		client = &http.Client{
			Transport: getHTTPTransport(cnf),
			Timeout:   cnf.HTTP.Timeout,
		}

		ehttp.Instrument(client)
	}

	return func(next http.Handler) http.Handler {
		return &Middleware{
			logger: logger,
			permission: &permission{
				Service:        cnf.Permission.Service,
				OwnershipLevel: ownershipLevel,
				Action:         action,
				Resource:       resource,
			},
			iamURL:  cnf.URL,
			enabled: cnf.Middleware.Enabled,
			next:    next,
		}
	}
}

// ServeHTTP calls Will.IAM with permission, sets on writer the returned
// token and calls next http.Handler.
func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !m.enabled {
		m.next.ServeHTTP(w, r)
		return
	}

	token := accessTokenFromHeader(r)
	if token == "" {
		m.logger.Error("request with empty access token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	status, err := m.getAuthStatus(m.permission.build(r), token)
	if err != nil {
		m.logger.WithError(err).Error("failed to get auth status")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if status.code != http.StatusOK {
		m.logger.
			WithField("statusCode", status.code).
			Error("received invalid status code")
		w.WriteHeader(status.code)
		return
	}

	if status.token != "" && status.token != token {
		w.Header().Set("x-access-token", status.token)
	}

	if status.email != "" {
		w.Header().Set("x-email", status.email)
	}

	m.next.ServeHTTP(w, r)
}

type auth struct {
	code  int
	token string
	email string
}

func (m *Middleware) getAuthStatus(permission, token string) (*auth, error) {
	url := fmt.Sprintf("%s/permissions/has?permission=%s",
		m.iamURL, url.QueryEscape(permission))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &auth{
		code:  res.StatusCode,
		token: res.Header.Get("x-access-token"),
		email: res.Header.Get("x-email"),
	}, nil
}

// Permission holds information to build the full Will.IAM permission
// string.
// For the description on service, ownershipLevel and action
// check the README on https://github.com/topfreegames/Will.IAM.
// The resource function is supposed to return a resource hierarchy
// (also described on the README) from the request: in this case,
// you choose to return a resource according to the request's path,
// body, etc.
type permission struct {
	Service        string
	OwnershipLevel string
	Action         string
	Resource       func(*http.Request) string
}

func (p *permission) build(r *http.Request) string {
	return fmt.Sprintf(
		"%s::%s::%s::%s",
		p.Service, p.OwnershipLevel, p.Action, p.Resource(r))
}

func accessTokenFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}
