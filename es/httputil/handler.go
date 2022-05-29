package httputil

import (
	"encoding/json"
	"eventstore/es"
	"net/http"
)

func NewCommandHandler[T es.Command](store es.EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var cmd T
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// get the transaction
		tx, err := store.GetTx(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := tx.Dispatch(ctx, cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
