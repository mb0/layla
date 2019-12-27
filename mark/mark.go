// Package mark provides a simple markdown text format.
package mark

import (
	"strings"

	"github.com/mb0/xelf/cor"
)

type Tag uint

const (
	B Tag = 1 << iota
	I
	M
	A

	H1
	H2
	H3
	H4
	HR
	P

	Text   = Tag(0)
	Style  = B | I | M | A
	Header = H1 | H2 | H3 | H4
	Block  = HR | P
	All    = Style | Header | Block
)

type El struct {
	Tag  Tag
	Cont string
	Els  []El
}

func Parse(txt string) ([]El, error)  { return All.Parse(txt) }
func Inline(txt string) ([]El, error) { return All.Inline(txt) }

func (tag Tag) Parse(txt string) (res []El, err error) {
	var line string
	var cont bool
	for len(txt) > 0 {
		line, txt = readLine(txt)
		var el *El
		switch {
		case tag&Header != 0 && strings.HasPrefix(line, "#"):
			var c int
			for c = 1; c < len(line); c++ {
				if line[c] != '#' {
					break
				}
			}
			el = &El{Cont: strings.TrimSpace(line[c:])}
			if c > 4 {
				c = 4
			}
			el.Tag = H1 << (c - 1)
			cont = false
		case tag&HR != 0 && strings.HasPrefix(line, "---"):
			el = &El{Tag: HR}
			line = strings.TrimLeft(line, "-")
			el.Cont = strings.TrimSpace(line)
			cont = false
		default:
			line = strings.TrimSpace(line)
			if line == "" {
				cont = false
				continue
			}
			els, err := tag.Inline(line)
			if err != nil {
				return res, err
			}
			if cont {
				el = &res[len(res)-1]
				el.Els = append(el.Els, els...)
				continue
			}
			el = &El{Tag: P, Els: els}
			cont = true
		}
		res = append(res, *el)
	}
	return
}

func readLine(txt string) (line, rest string) {
	end := strings.IndexByte(txt, '\n')
	if end < 0 {
		return txt, ""
	}
	line, rest = txt[:end], txt[end+1:]
	if end > 0 && line[end-1] == '\r' {
		line = line[:end-1]
	}
	return
}

func (tag Tag) Inline(txt string) (res []El, _ error) {
	var start, i int
	for i < len(txt) {
		c := rune(txt[i])
		tag, end, ok := tag.inlineStart(c)
		switch ok {
		case true:
			cont, n := consumeSpan(txt[i:], end)
			if n == 0 {
				break
			}
			var link string
			var nn int
			if tag == A {
				ii := i + n
				ii += skipSpace(txt[ii:])
				if ii >= len(txt) || txt[ii] != '(' {
					break
				}
				link, nn = consumeSpan(txt[ii:], ')')
				if nn == 0 {
					break
				}
				nn += ii - i - n
			}
			if start < i {
				cont := txt[start:i]
				res = append(res, El{Cont: cont})
			}
			i += n + nn
			el := El{Tag: tag, Cont: cont}
			if tag == A {
				el.Els = []El{{Cont: el.Cont}}
				el.Cont = link
			}
			start = i
			res = append(res, el)
			continue

		}
		for _, c := range txt[i:] {
			i++
			if cor.Space(c) {
				break
			}
		}
		i += skipSpace(txt[i:])
	}
	if start < len(txt) {
		res = append(res, El{Cont: txt[start:]})
	}
	return res, nil
}

func (tag Tag) inlineStart(c rune) (Tag, rune, bool) {
	switch {
	case tag&A != 0 && c == '[': // link
		return A, ']', true
	case tag&B != 0 && c == '*': // emphasis
		return B, c, true
	case tag&I != 0 && c == '_':
		return I, c, true
	case tag&B != 0 && c == '`':
		return M, c, true
	}
	return Text, 0, false
}

func skipSpace(s string) (n int) {
	for _, c := range s {
		if !cor.Space(c) {
			return
		}
		n++
	}
	return
}

func consumeSpan(txt string, end rune) (string, int) {
	var esc bool
	for i, r := range txt {
		if i == 0 {
			continue
		}
		if i == 1 && r == ' ' {
			break
		}
		if esc {
			esc = false
			continue
		}
		switch r {
		case '\\':
			esc = true
		case end:
			return txt[1:i], i + 1
		}
	}
	return "", 0
}
