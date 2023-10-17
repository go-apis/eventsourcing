package tests

import (
	"testing"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/noop"
)

func Test(t *testing.T) {
	// tester, err := NewTester()
	// require.NoError(t, err)

	// t.Run("create", func(t *testing.T) {
	// })
}
