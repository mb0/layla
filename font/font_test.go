package font

import (
	"reflect"
	"testing"

	"golang.org/x/image/math/fixed"
)

func TestLayout(t *testing.T) {
	m := &Manager{}
	err := m.RegisterTTF("go", "../testdata/font/Go-Regular.ttf")
	if err != nil {
		t.Fatalf("parse font %v", err)
	}
	f, err := m.Face("go", 10)
	if err != nil {
		t.Fatalf("find font %v", err)
	}
	lineheight := f.Metrics().Height
	if lineheight != fixed.I(10) {
		t.Errorf("lineheight %s", lineheight)
	}
	tests := []struct {
		text  string
		width int
		lines []string
	}{
		{"Hello world", 60, []string{"Hello world"}},
		{"Hello world", 30, []string{"Hello", "world"}},
		{"Hello world", 20, []string{"Hell", "o w", "orld"}},
		{"Tobeornottobe", 20, []string{"Tob", "eor", "nott", "obe"}},
		{"To be or not to be", 50, []string{"To be or", "not to be"}},
		{"To be or_not to be", 50, []string{"To be", "or_not to", "be"}},
		{"To be or-not to be", 50, []string{"To be or-", "not to be"}},
		{"To be\nor not\nto be", 50, []string{"To be", "or not", "to be"}},
	}
	for i, test := range tests {
		lines, _, _, err := Layout(f, test.text, test.width)
		if err != nil {
			t.Errorf("layout error: %v", err)
			continue
		}
		if !reflect.DeepEqual(test.lines, lines) {
			t.Errorf("test %d want lines %q got %q", i, test.lines, lines)
		}
	}
}

func TestTokens(t *testing.T) {
	tests := []struct {
		text string
		toks []tok
	}{
		{"x y", []tok{{0, 1}, {1, 1}, {2, 1}}},
		{"x  \n  y", []tok{{0, 1}, {3, 0}, {4, 2}, {6, 1}}},
		{"x  \ny", []tok{{0, 1}, {3, 0}, {4, 1}}},
		{"x  \n  \n  y", []tok{{0, 1}, {3, 0}, {6, 0}, {7, 2}, {9, 1}}},
		{"foo  bar", []tok{{0, 3}, {3, 2}, {5, 3}}},
		{"-o-  bar", []tok{{0, 3}, {3, 2}, {5, 3}}},
		{"foo-bar", []tok{{0, 4}, {4, 3}}},
	}
	for _, test := range tests {
		toks := tokens(test.text)
		if !reflect.DeepEqual(test.toks, toks) {
			t.Errorf("test %q want toks %v got %v", test.text, test.toks, toks)
		}
	}
}
