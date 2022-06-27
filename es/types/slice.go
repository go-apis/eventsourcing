package types

import "encoding/json"

type Slice[T any] []T

func (n *Slice[T]) UnmarshalJSON(buf []byte) error {
	if len(buf) > 0 && buf[0] == '{' {
		// get the index.
		s := struct {
			Index  int             `json:"index"`
			Value  json.RawMessage `json:"value"`
			Delete bool            `json:"delete"`
		}{}
		if err := json.Unmarshal(buf, &s); err != nil {
			return err
		}

		if s.Delete {
			*n = append((*n)[:s.Index], (*n)[s.Index+1:]...)
			return nil
		}

		// grow the slice?
		len := s.Index - len(*n) + 1
		if len > 0 {
			*n = append(*n, make([]T, len)...)
		}
		if err := json.Unmarshal(s.Value, &(*n)[s.Index]); err != nil {
			return err
		}
		return nil
	}

	var items []T
	if err := json.Unmarshal(buf, &items); err != nil {
		return err
	}
	*n = items
	return nil
}

type SliceItem[T any] struct {
	Index int `json:"index"`
	Value T   `json:"value"`
}

type SliceDelete struct {
	Index  int  `json:"index"`
	Delete bool `json:"delete"`
}
