package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_TagSplit(t *testing.T) {
	var data = []struct {
		tag string
		out []string
	}{
		{
			tag: `name,eq`,
			out: []string{`name`, `eq`},
		},
		{
			tag: `name,eq,again;again2`,
			out: []string{`name`, `eq`, `again`, `again2`},
		},
		{
			tag: `name,eq,again;again2`,
			out: []string{`name`, `eq`, `again`, `again2`},
		},
		{
			tag: `name,eq,string_to_array(version\, '.')::int[];again2;again3`,
			out: []string{`name`, `eq`, `string_to_array(version, '.')::int[]`, `again2`, `again3`},
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("data[%d]", i), func(t *testing.T) {
			out := SplitTag(d.tag)
			if !reflect.DeepEqual(out, d.out) {
				t.Errorf("expected %v, got %v", d.out, out)
			}
		})
	}
}
