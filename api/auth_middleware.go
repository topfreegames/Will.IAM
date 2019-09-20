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
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			header := request.Header.Get("authorization")
			authHeaderContents := strings.Split(header, " ")
			authHeader := buildAuth(header, authHeaderContents[0], authHeaderContents[1])

			if !authHeader.isValid() {
				responseWriter.WriteHeader(http.StatusUnauthorized)
				return
			}

			var ctx context.Context
			var err error
			logger := middleware.GetLogger(request.Context())

			if authHeader.isKeyPair() {
				ctx, err = handleKeyPairAuth(request, responseWriter, authHeader, sasUC, logger)
			} else if authHeader.isOAuth2() {
				ctx, err = handleOAuth2TokenAuth(request, responseWriter, authHeader, sasUC, logger)
			} else {
				handleUndefinedAuthType(responseWriter, logger)
				return
			}

			if err != nil {
				return
			}

			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}

func buildAuth(header string, method string, content string) authorizationHeader {
	var authType models.AuthenticationType

	if strings.EqualFold(method, keyPairHeader) {
		authType = models.AuthenticationTypes.KeyPair
	} else if strings.EqualFold(method, bearerTokenHeader) {
		authType = models.AuthenticationTypes.OAuth2
	} else {
		authType = ""
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

func (auth authorizationHeader) isKeyPair() bool {
	return strings.EqualFold(auth.Type.String(), models.AuthenticationTypes.KeyPair.String())
}

func (auth authorizationHeader) isOAuth2() bool {
	return strings.EqualFold(auth.Type.String(), models.AuthenticationTypes.OAuth2.String())
}

func handleKeyPairAuth(
	request *http.Request,
	responseWriter http.ResponseWriter,
	authHeader authorizationHeader,
	sasUC usecases.ServiceAccounts,
	logger logrus.FieldLogger,
) (context.Context, error) {

	keyPair := strings.Split(authHeader.Content, ":")
	accessKeyPairAuth, err := sasUC.WithContext(request.Context()).AuthenticateKeyPair(keyPair[0], keyPair[1])

	if err != nil {
		logger.WithError(err).Error("auth failed")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	responseWriter.Header().Set("x-service-account-name", accessKeyPairAuth.Name)

	ctx := context.WithValue(request.Context(), serviceAccountIDCtxKey, accessKeyPairAuth.ServiceAccountID)
	return ctx, nil
}

func handleOAuth2TokenAuth(
	request *http.Request,
	responseWriter http.ResponseWriter,
	authHeader authorizationHeader,
	sasUC usecases.ServiceAccounts,
	logger logrus.FieldLogger,
) (context.Context, error) {

	accessToken := authHeader.Content
	accessTokenAuth, err := sasUC.WithContext(request.Context()).AuthenticateAccessToken(accessToken)

	if err != nil {
		logger.WithError(err).Info("auth failed")

		if _, ok := err.(*errors.EntityNotFoundError); ok {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}

		logger.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	responseWriter.Header().Set("x-email", accessTokenAuth.Email)

	if accessTokenAuth.AccessToken != accessToken {
		responseWriter.Header().Set("x-access-token", accessTokenAuth.AccessToken)
	}

	ctx := context.WithValue(request.Context(), serviceAccountIDCtxKey, accessTokenAuth.ServiceAccountID)
	return ctx, nil
}

func handleUndefinedAuthType(
	responseWriter http.ResponseWriter,
	logger logrus.FieldLogger,
) {
	logger.WithError(errors.NewInvalidAuthorizationTypeError()).Error("auth failed")
	responseWriter.WriteHeader(http.StatusUnauthorized)
	return
}
