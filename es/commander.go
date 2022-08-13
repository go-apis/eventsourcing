package es

import (
	"encoding/json"
	"net/http"
)

func NewCommander[T Command]() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		unit, err := GetUnit(ctx)
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

		tx, err := unit.NewTx(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(ctx)

		if err := unit.Dispatch(ctx, cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := tx.Commit(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
