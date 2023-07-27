package auth

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
)

const DisabledType = "disabled"

type Disabled struct {
	Log *logrus.Entry
}

func (d Disabled) AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func (d Disabled) Init(ctx context.Context, params Params) error {
	d.Log.Warnf("authentication type is %v. Authentication is DISABLED!", DisabledType)
	return nil
}
