package es

import "testing"

func Test_Builder(t *testing.T) {
	b := NewBuilder().
		AddEntity(&demoEntity{})

	cfg, err := b.Build()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cfg)
}
