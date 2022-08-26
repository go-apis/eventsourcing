package es

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
)

func NewCommander[T Command]() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		pctx, span := otel.Tracer("es").Start(ctx, "NewCommander")
		defer span.End()

		unit, err := GetUnit(pctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var cmd T
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		tx, err := unit.NewTx(pctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(pctx)

		if err := unit.Dispatch(pctx, cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := tx.Commit(pctx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
