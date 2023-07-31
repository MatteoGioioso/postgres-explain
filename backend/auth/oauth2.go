package auth

import (
	"context"
	"fmt"
	"github.com/borealisdb/commons/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
	"net/http"
)

const Oauth2Type = "oauth2"

type Oauth2 struct {
	idTokenVerifier *oidc.IDTokenVerifier
	Params
	Log *logrus.Entry
}

func (d *Oauth2) Init(ctx context.Context, params Params) error {
	d.Log.Infof("authentication type is %v", Oauth2Type)
	_, idTokenVerifier, err := auth.InitializeAuth(ctx, params.IssuerUrl, params.ClientID)
	if err != nil {
		return fmt.Errorf("could not InitializeAuth: %v", err)
	}

	d.idTokenVerifier = idTokenVerifier

	return nil
}

func (d *Oauth2) verifyToken(ctx context.Context, rawIDToken string) (auth.IDTokenClaims, error) {
	idToken, err := d.idTokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return auth.IDTokenClaims{}, err
	}

	claims := auth.IDTokenClaims{}
	if err := idToken.Claims(&claims); err != nil {
		return auth.IDTokenClaims{}, err
	}

	return claims, err
}

func (d *Oauth2) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		borealisCookie, err := r.Cookie("borealis")
		if err != nil {
			borealisCookie = &http.Cookie{} // TODO, maybe this is not the best
		}
		if _, err := d.verifyToken(r.Context(), borealisCookie.Value); err != nil {
			w.WriteHeader(401)
			return
		}

		h.ServeHTTP(w, r)
	})
}
