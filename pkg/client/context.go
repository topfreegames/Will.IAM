package client

import (
	"context"
	"net/http"
)

type contextKey int

const (
	authKey contextKey = iota
	williamKey
)

func Auth(r *http.Request) *AuthInfo {
	if r == nil {
		return nil
	}

	return auth(r)
}

func auth(r *http.Request) *AuthInfo {
	if a := r.Context().Value(authKey); a != nil {
		return a.(*AuthInfo)
	}

	return nil
}

func Get(r *http.Request) William {
	if r == nil {
		return nil
	}

	return get(r)
}

func get(r *http.Request) William {
	if a := r.Context().Value(williamKey); a != nil {
		return a.(William)
	}

	return nil
}

func setAuth(r *http.Request, a *AuthInfo) *http.Request {
	if a == nil {
		return r
	}

	return r.WithContext(context.WithValue(r.Context(), authKey, a))
}

func setWilliam(r *http.Request, wi William) *http.Request {
	if wi == nil {
		return r
	}

	return r.WithContext(context.WithValue(r.Context(), williamKey, wi))
}
