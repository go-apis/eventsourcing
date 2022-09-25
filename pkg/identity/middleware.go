package identity

import (
	"net/http"

	"github.com/contextcloud/eventstore/es"
)

type Fetch func(r *http.Request) (User, error)

func Middleware(fn Fetch) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, err := fn(r)
			if err != nil {
				// Do error handling!
			}
			if user != nil {
				ctx = SetUser(ctx, user)
				ctx = es.SetNamespace(ctx, user.GetAudience())
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
