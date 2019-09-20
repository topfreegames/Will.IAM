package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

type serviceAccountIDCtxKeyType string

const serviceAccountIDCtxKey = serviceAccountIDCtxKeyType("serviceAccountID")
const keyPairHeader = "KeyPair"
const bearerTokenHeader = "Bearer"

type authorizationHeader struct {
	Header  string
	Type    models.AuthenticationType
	Content string
}

func getServiceAccountID(ctx context.Context) (string, bool) {
	v := ctx.Value(serviceAccountIDCtxKey)
	vv, ok := v.(string)
	if !ok {
		return "", false
	}
	return vv, true
}

// authMiddleware authenticates either access_token or key pair
func authMiddleware(sasUC usecases.ServiceAccounts) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("authorization")
			authHeaderContents := strings.Split(header, " ")
			authHeader := buildAuth(header, authHeaderContents[0], authHeaderContents[1])
			if !authHeader.isValid() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			var ctx context.Context
			var err error
			logger := middleware.GetLogger(r.Context())

			switch authHeader.Type {
			case models.AuthenticationTypes.KeyPair:
				ctx, err = handleKeyPairAuth(r, w, authHeader, sasUC)
			case models.AuthenticationTypes.OAuth2:
				ctx, err = handleOAuth2TokenAuth(r, w, authHeader, sasUC)
			default:
				handleUndefinedAuthType(w, logger)
				return
			}

			if err != nil {
				logger.WithError(err).Error("auth failed")
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func buildAuth(header, method, content string) authorizationHeader {
	var authType models.AuthenticationType

	if strings.EqualFold(method, keyPairHeader) {
		authType = models.AuthenticationTypes.KeyPair
	} else if strings.EqualFold(method, bearerTokenHeader) {
		authType = models.AuthenticationTypes.OAuth2
	} else {
		authType = models.AuthenticationTypes.Unknown
	}

	return authorizationHeader{
		Header:  header,
		Type:    authType,
		Content: content,
	}
}

func (auth authorizationHeader) isValid() bool {
	if auth.Header == "" || auth.Type == "" || auth.Content == "" {
		return false
	}
	return true
}

func handleKeyPairAuth(
	r *http.Request,
	w http.ResponseWriter,
	authHeader authorizationHeader,
	sasUC usecases.ServiceAccounts,
) (context.Context, error) {
	keyPair := strings.Split(authHeader.Content, ":")
	accessKeyPairAuth, err := sasUC.WithContext(r.Context()).AuthenticateKeyPair(keyPair[0], keyPair[1])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	w.Header().Set("x-service-account-name", accessKeyPairAuth.Name)
	ctx := context.WithValue(r.Context(), serviceAccountIDCtxKey, accessKeyPairAuth.ServiceAccountID)

	return ctx, nil
}

func handleOAuth2TokenAuth(
	r *http.Request,
	w http.ResponseWriter,
	authHeader authorizationHeader,
	sasUC usecases.ServiceAccounts,
) (context.Context, error) {
	accessToken := authHeader.Content
	accessTokenAuth, err := sasUC.WithContext(r.Context()).AuthenticateAccessToken(accessToken)

	if err != nil {
		if _, ok := err.(*errors.EntityNotFoundError); ok {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}

		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	w.Header().Set("x-email", accessTokenAuth.Email)

	if accessTokenAuth.AccessToken != accessToken {
		w.Header().Set("x-access-token", accessTokenAuth.AccessToken)
	}

	ctx := context.WithValue(r.Context(), serviceAccountIDCtxKey, accessTokenAuth.ServiceAccountID)
	return ctx, nil
}

func handleUndefinedAuthType(
	w http.ResponseWriter,
	logger logrus.FieldLogger,
) {
	logger.WithError(errors.NewInvalidAuthorizationTypeError()).Error("auth failed")
	w.WriteHeader(http.StatusUnauthorized)
	return
}
