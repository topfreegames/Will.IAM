package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

type serviceAccountIDCtxKeyType string

const serviceAccountIDCtxKey = serviceAccountIDCtxKeyType("serviceAccountID")

func getServiceAccountID(ctx context.Context) (string, bool) {
	v := ctx.Value(serviceAccountIDCtxKey)
	vv, ok := v.(string)
	if !ok {
		return "", false
	}
	return vv, true
}

// authMiddleware authenticates either access_token or key pair
func authMiddleware(
	sasUC usecases.ServiceAccounts,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("authorization")
			parts := strings.Split(authorization, " ")
			if authorization == "" || len(parts) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var ctx context.Context
			l := middleware.GetLogger(r.Context())
			if strings.EqualFold(parts[0], "KeyPair") {
				keyPair := strings.Split(parts[1], ":")
				accessKeyPairAuth, err := sasUC.WithContext(r.Context()).
					AuthenticateKeyPair(keyPair[0], keyPair[1])
				if err != nil {
					l.WithError(err).Error("auth failed")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("x-service-account-name", accessKeyPairAuth.Name)
				ctx = context.WithValue(r.Context(), serviceAccountIDCtxKey, accessKeyPairAuth.ServiceAccountID)
			} else if strings.EqualFold(parts[0], "Bearer") {
				accessToken := parts[1]
				accessTokenAuth, err := sasUC.WithContext(r.Context()).
					AuthenticateAccessToken(accessToken)
				if err != nil {
					l.WithError(err).Info("auth failed")
					if _, ok := err.(*errors.EntityNotFoundError); ok {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					l.Error(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("x-email", accessTokenAuth.Email)
				if accessTokenAuth.AccessToken != accessToken {
					w.Header().Set("x-access-token", accessTokenAuth.AccessToken)
				}
				ctx = context.WithValue(
					r.Context(), serviceAccountIDCtxKey, accessTokenAuth.ServiceAccountID,
				)
			} else {
				l.WithError(errors.NewInvalidAuthorizationTypeError()).Error("auth failed")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
