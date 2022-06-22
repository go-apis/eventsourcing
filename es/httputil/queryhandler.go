package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
)

func NewQueryHandler[T any](cli es.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		unit, err := cli.NewUnit(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		filter := filters.Filter{
			Where: []filters.Where{
				filters.WhereClause{
					Column: "username",
					Op:     "like",
					Args:   []interface{}{"ca%"},
				},
				filters.WhereOr{
					Where: filters.WhereClause{
						Column: "username",
						Op:     "like",
						Args:   []interface{}{"ch%"},
					},
				},
			},
		}

		q := es.NewQuery[T](unit)
		out, err := q.Find(ctx, filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
