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
			{Tag: "p", Els: []El{{Cont: "test"}}},
		}},
		{"test\ntest", []El{
			{Tag: "p", Els: []El{{Cont: "test"}, {Cont: "test"}}},
		}},
		{"test\n\ntest", []El{
			{Tag: "p", Els: []El{{Cont: "test"}}},
			{Tag: "p", Els: []El{{Cont: "test"}}},
		}},
		{"test\n------\ntest", []El{
			{Tag: "p", Els: []El{{Cont: "test"}}},
			{Tag: "hr"},
			{Tag: "p", Els: []El{{Cont: "test"}}},
		}},
		{"#title\n###title\n########title\ntest", []El{
			{Tag: "h1", Cont: "title"},
			{Tag: "h3", Cont: "title"},
			{Tag: "h6", Cont: "title"},
			{Tag: "p", Els: []El{{Cont: "test"}}},
		}},
		{"test [Link](url) *test* _test_ `test` test", []El{
			{Tag: "p", Els: []El{
				{Cont: "test "},
				{Tag: "a", Cont: "url", Els: []El{{Cont: "Link"}}},
				{Cont: " "},
				{Tag: "b", Cont: "test"},
				{Cont: " "},
				{Tag: "i", Cont: "test"},
				{Cont: " "},
				{Tag: "code", Cont: "test"},
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
