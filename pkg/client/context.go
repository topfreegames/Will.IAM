package client

import (
	"context"
)

type contextKey int

const (
	authKey contextKey = iota
	williamKey
)

func Auth(ctx context.Context) *AuthInfo {
	if ctx == nil {
		return nil
	}

	return auth(ctx)
}

func auth(ctx context.Context) *AuthInfo {
	if a := ctx.Value(authKey); a != nil {
		return a.(*AuthInfo)
	}

	return nil
}

func Get(ctx context.Context) William {
	if ctx == nil {
		return nil
	}

	return get(ctx)
}

func get(ctx context.Context) William {
	if a := ctx.Value(williamKey); a != nil {
		return a.(William)
	}

	return nil
}

func setAuth(ctx context.Context, a *AuthInfo) context.Context {
	if a == nil {
		return ctx
	}

	return context.WithValue(ctx, authKey, a)
}

func setWilliam(ctx context.Context, will William) context.Context {
	if will == nil {
		return ctx
	}

	return context.WithValue(ctx, williamKey, will)
}
