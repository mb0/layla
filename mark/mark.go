// Package mark provides a simple markdown text format.
package mark

import (
	"fmt"
	"strings"

	"github.com/mb0/xelf/cor"
)

type El struct {
	Tag  string
	Cont string
	Els  []El
}

func Parse(txt string) (res []El, err error) {
	var line string
	var cont bool
	for len(txt) > 0 {
		line, txt = readLine(txt)
		var el *El
		switch {
		case strings.HasPrefix(line, "#"):
			var c int
			for c = 1; c < len(line); c++ {
				if line[c] != '#' {
					break
				}
			}
			el = &El{Cont: strings.TrimSpace(line[c:])}
			if c > 6 {
				c = 6
			}
			el.Tag = fmt.Sprintf("h%d", c)
			cont = false
		case strings.HasPrefix(line, "---"):
			el = &El{Tag: "hr"}
			line = strings.TrimLeft(line, "-")
			el.Cont = strings.TrimSpace(line)
			cont = false
		default:
			line = strings.TrimSpace(line)
			if line == "" {
				cont = false
				continue
			}
			els, err := Inline(line)
			if err != nil {
				return res, err
			}
			if cont {
				el = &res[len(res)-1]
				el.Els = append(el.Els, els...)
				continue
			}
			el = &El{Tag: "p", Els: els}
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

func Inline(txt string) (res []El, _ error) {
	var start, i int
	for i < len(txt) {
		c := rune(txt[i])
		tag, end, ok := inlineStart(c)
		switch ok {
		case true:
			cont, n := consumeSpan(txt[i:], end)
			if n == 0 {
				break
			}
			var link string
			var nn int
			if tag == "a" {
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
			if tag == "a" {
				el.Els = []El{{Cont: el.Cont}}
				el.Cont = link
			}
			start = i
			res = append(res, el)
			continue

		}
		for _, c := range txt[i:] {
			i++
			if c == ' ' {
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

func skipSpace(s string) (n int) {
	for _, c := range s {
		if !cor.Space(c) {
			return
		}
		n++
	}
	return
}

func inlineStart(c rune) (string, rune, bool) {
	switch c {
	case '[': // link
		return "a", ']', true
	case '*': // emphasis
		return "b", c, true
	case '_':
		return "i", c, true
	case '`':
		return "code", c, true
	}
	return "", 0, false
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
