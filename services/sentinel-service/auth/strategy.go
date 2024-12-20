package auth

import (
	"context"
	"net/http"
)

type AuthenticationStrategy interface {
	AuthenticateRequest(ctx context.Context, r *http.Request) error
	FetchTokens(ctx context.Context, r *http.Request) (string, string, error)
}
