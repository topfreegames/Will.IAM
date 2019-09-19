package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

type serviceAccountIDCtxKeyType string

const serviceAccountIDCtxKey = serviceAccountIDCtxKeyType("serviceAccountID")
const keyPairHeader = "KeyPair"
const bearerTokenHeader = "Bearer"

type authorizationHeader struct {
	Header  string
	Type    string
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
				ctx, err = handleKeyPairAuthorization(request, responseWriter, authHeader, sasUC, logger)
			} else if authHeader.isBearerToken() {
				ctx, err = handleBearerTokenAuthorization(request, responseWriter, authHeader, sasUC, logger)
			} else {
				handleUndefinedAuthorizationType(responseWriter, logger)
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
	return authorizationHeader{
		Header:  header,
		Type:    method,
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
	return strings.EqualFold(auth.Type, keyPairHeader)
}

func (auth authorizationHeader) isBearerToken() bool {
	return strings.EqualFold(auth.Type, bearerTokenHeader)
}

func handleKeyPairAuthorization(
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

func handleBearerTokenAuthorization(
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

func handleUndefinedAuthorizationType(
	responseWriter http.ResponseWriter,
	logger logrus.FieldLogger,
) {
	logger.WithError(errors.NewInvalidAuthorizationTypeError()).Error("auth failed")
	responseWriter.WriteHeader(http.StatusUnauthorized)
	return
}
