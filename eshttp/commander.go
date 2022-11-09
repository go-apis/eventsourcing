package eshttp

import (
	"encoding/json"
	"net/http"

	"github.com/contextcloud/eventstore/es"
	"go.opentelemetry.io/otel"
)

func NewCommander[T es.Command]() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		pctx, span := otel.Tracer("es").Start(ctx, "NewCommander")
		defer span.End()

		var cmd T
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := es.Dispatch(pctx, cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
