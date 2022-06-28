package es

import (
	"encoding/json"
	"net/http"
)

type Commander[T Command] interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type commander[T Command] struct {
	cli Client
}

func (c *commander[T]) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var cmd T
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	unit, err := c.cli.NewUnit(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer unit.Rollback(ctx)

	if err := unit.Dispatch(ctx, cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := unit.Commit(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func NewCommander[T Command](cli Client) Commander[T] {
	return &commander[T]{cli: cli}
}
