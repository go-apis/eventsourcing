package tests

import (
	"testing"

	_ "github.com/go-apis/eventsourcing/es/providers/data/pg"
	_ "github.com/go-apis/eventsourcing/es/providers/stream/noop"
)

func Test(t *testing.T) {
	// tester, err := NewTester()
	// require.NoError(t, err)

	// t.Run("create", func(t *testing.T) {
	// })
}
