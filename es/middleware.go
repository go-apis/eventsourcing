package es

import "net/http"

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func CreateUnit(cli Client) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lrw := NewLoggingResponseWriter(w)

			ctx := r.Context()
			unit, err := cli.Unit(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			defer func() {
				if lrw.statusCode >= 400 {
					unit.Rollback(ctx)
					return
				}
				if _, err := unit.Commit(ctx); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}()

			ctx = SetUnit(ctx, unit)
			h.ServeHTTP(lrw, r.WithContext(ctx))
		})
	}
}
