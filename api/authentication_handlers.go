package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ghostec/Will.IAM/models"
	"github.com/ghostec/Will.IAM/oauth2"
	"github.com/ghostec/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

func authenticationBuildURLHandler(
	provider oauth2.Provider,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		qs := r.URL.Query()
		if len(qs["referer"]) == 0 {
			Write(
				w, http.StatusUnprocessableEntity,
				`{ "error": "querystrings.referer is required" }`,
			)
			return
		}
		authURL := provider.WithContext(r.Context()).BuildAuthURL(qs["referer"][0])
		http.Redirect(w, r, authURL, http.StatusSeeOther)
	}
}

func authenticationExchangeCodeHandler(
	provider oauth2.Provider, sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		qs := r.URL.Query()
		if len(qs["code"]) == 0 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if len(qs["state"]) == 0 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		code := qs["code"][0]
		authResult, err := provider.WithContext(r.Context()).ExchangeCode(code)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sa := &models.ServiceAccount{
			Name:    authResult.Email,
			Email:   authResult.Email,
			Picture: authResult.Picture,
		}
		if err = sasUC.WithContext(r.Context()).Create(sa); err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		v := url.Values{}
		v.Add("accessToken", authResult.AccessToken)
		v.Add("email", authResult.Email)
		v.Add("referer", qs["state"][0])
		redirectTo := fmt.Sprintf("/sso?%s", v.Encode())
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}
}

func authenticationValidHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		qs := r.URL.Query()
		if len(qs["referer"]) == 0 {
			Write(
				w, http.StatusUnprocessableEntity,
				`{ "error": "querystrings.referer is required" }`,
			)
			return
		}
		if len(qs["accessToken"]) == 0 {
			Write(
				w, http.StatusUnprocessableEntity,
				`{ "error": "querystrings.accessToken is required" }`,
			)
			return
		}
		authResult, err := sasUC.WithContext(r.Context()).
			AuthenticateAccessToken(qs["accessToken"][0])
		if err != nil {
			l.Error(err)
			v := url.Values{}
			v.Add("referer", qs["referer"][0])
			http.Redirect(
				w, r, fmt.Sprintf("/sso/auth/do?%s", v.Encode()), http.StatusSeeOther,
			)
			return
		}
		v := url.Values{}
		v.Add("referer", qs["referer"][0])
		v.Add("accessToken", authResult.AccessToken)
		sep := "?"
		if strings.Contains(qs["referer"][0], "?") {
			sep = "&"
		}
		http.Redirect(
			w, r, fmt.Sprintf("%s%s%s", qs["referer"][0], sep, v.Encode()),
			http.StatusSeeOther,
		)
	}
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	// Work is in authMiddleware
	w.WriteHeader(200)
}
