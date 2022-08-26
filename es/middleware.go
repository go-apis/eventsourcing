package es

import "net/http"

func CreateUnit(cli Client) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			unit, err := cli.Unit(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ctx = SetUnit(ctx, unit)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
