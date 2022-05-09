package types

import (
	"encoding/json"
	"testing"
)

func Test_SliceMarshal(t *testing.T) {
	var s Slice[int]
	s = append(s, 1)
	s = append(s, 2)
	s = append(s, 3)
	s = append(s, 4)

	items, err := json.Marshal(s)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Out %s", items)
}
func Test_SliceUnmarshal(t *testing.T) {
	raw := []byte(`[1,2,3,4]`)
	var s Slice[int]

	if err := json.Unmarshal(raw, &s); err != nil {
		t.Error(err)
		return
	}

	t.Logf("Out %v", s)
}

func Test_SliceUnmarshalItems(t *testing.T) {
	items := [][]byte{
		[]byte(`{"value":5,"index":5}`),
		[]byte(`{"value":3,"index":3}`),
		[]byte(`{"value":1,"index":1}`),
		[]byte(`{"value":2,"index":2}`),
		[]byte(`{"value":4,"index":4}`),
		[]byte(`{"value":6,"index":5}`),
	}

	var s Slice[int]

	for _, raw := range items {
		if err := json.Unmarshal(raw, &s); err != nil {
			t.Error(err)
			return
		}
	}

	t.Logf("Out %v", s)
}
