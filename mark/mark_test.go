package mark

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		raw string
		res []El
	}{
		{"test", []El{
			{Tag: P, Els: []El{{Cont: "test"}}},
		}},
		{"test\ntest", []El{
			{Tag: P, Els: []El{{Cont: "test"}, {Cont: "test"}}},
		}},
		{"test\n\ntest", []El{
			{Tag: P, Els: []El{{Cont: "test"}}},
			{Tag: P, Els: []El{{Cont: "test"}}},
		}},
		{"test\n------\ntest", []El{
			{Tag: P, Els: []El{{Cont: "test"}}},
			{Tag: HR},
			{Tag: P, Els: []El{{Cont: "test"}}},
		}},
		{"#title\n###title\n########title\ntest", []El{
			{Tag: H1, Cont: "title"},
			{Tag: H3, Cont: "title"},
			{Tag: H4, Cont: "title"},
			{Tag: P, Els: []El{{Cont: "test"}}},
		}},
		{"test [Link](url) *test* _test_ `test` test", []El{
			{Tag: P, Els: []El{
				{Cont: "test "},
				{Tag: A, Cont: "url", Els: []El{{Cont: "Link"}}},
				{Cont: " "},
				{Tag: B, Cont: "test"},
				{Cont: " "},
				{Tag: I, Cont: "test"},
				{Cont: " "},
				{Tag: M, Cont: "test"},
				{Cont: " test"},
			}},
		}},
	}
	for _, test := range tests {
		res, err := Parse(test.raw)
		if err != nil {
			t.Errorf("parse %q err: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(res, test.res) {
			t.Errorf("parse %q\nwant %v\n got %v", test.raw, test.res, res)
		}
	}
}
