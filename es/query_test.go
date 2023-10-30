package es

import "testing"

func TestIt(t *testing.T) {
	data := []struct {
		limit    int
		offset   int
		expected int
	}{
		{limit: 10, offset: 0, expected: 1},
		{limit: 10, offset: 10, expected: 2},
		{limit: 10, offset: 15, expected: 2},
		{limit: 10, offset: 20, expected: 3},
		{limit: 10, offset: 30, expected: 4},
		{limit: 10, offset: 40, expected: 5},
	}

	for _, d := range data {
		if toPage(d.limit, d.offset) != d.expected {
			t.Errorf("Expected %d, got %d", d.expected, toPage(d.limit, d.offset))
		}
	}
}
