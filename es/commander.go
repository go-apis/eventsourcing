package es

import (
	"encoding/json"
	"net/http"
)

type Commander[T Command] interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type commander[T Command] struct {
	unit Unit
}

func (c *commander[T]) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var cmd T
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := c.unit.NewTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	if err := c.unit.Dispatch(ctx, cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := tx.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func NewCommander[T Command](unit Unit) Commander[T] {
	return &commander[T]{unit: unit}
}
