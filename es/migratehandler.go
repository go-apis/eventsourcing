package es

import (
	"encoding/json"
	"net/http"
)

// MigrateDbHandler returns an http.Handler that runs Client.MigrateDb on
// POST. Services that disable migrate-on-boot (it makes scale-from-zero
// cold starts slow) mount this and have their deploy pipeline call it
// after each rollout; AutoMigrate is idempotent, so repeated calls are
// safe. The handler does no authentication — callers are expected to
// protect the route with platform-level auth (e.g. Cloud Run IAM).
func MigrateDbHandler(cli Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := cli.MigrateDb(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"migrated": true})
	})
}
