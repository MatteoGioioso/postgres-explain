package auth

import (
	"context"
	"net/http"
)

type Auth interface {
	AuthMiddleware(h http.Handler) http.Handler
	Init(ctx context.Context, params Params) error
}

type Factory struct {
	Providers map[string]Auth
}

func (f Factory) Get(provider string) Auth {
	return f.Providers[provider]
}

type Params struct {
	IssuerUrl string
	ClientID  string
}
