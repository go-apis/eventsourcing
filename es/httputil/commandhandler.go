package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/contextcloud/eventstore/es"
)

func NewCommandHandler[T es.Command](cli es.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var cmd T
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		unit, err := cli.NewUnit(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := unit.Dispatch(ctx, cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := unit.Commit(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
