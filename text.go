package layla

import (
	"bytes"
	"math"
	"strings"

	"github.com/mb0/layla/font"
	"github.com/mb0/layla/mark"
	"github.com/mb0/xelf/cor"
)

func (l *Layouter) lineLayout(n *Node, stack []*Node) (err error) {
	markup := n.Kind == "markup"
	var els []mark.El
	if markup {
		els, err = mark.Inline(n.Data)
		if err != nil {
			return err
		}
	} else {
		els = []mark.El{{Cont: n.Data}}
	}
	stack = append(stack, n)
	of := getFont(stack)
	b := n.Pad.Inset(n.Calc)
	lh, err := l.lineHeight(of)
	if err != nil {
		return err
	}
	s := &splitter{Layouter: l, Font: *of, Max: b.W}
	res, err := s.lines(els)
	if err != nil {
		return err
	}
	if markup {
		n.List = make([]*Node, 0, len(res))
	}
	var buf bytes.Buffer
	var y, mw float64
	for li, line := range res {
		bx := b.X
		switch n.Align {
		case 2: // center
			bx += math.Floor((b.W - line.W) / 2)
		case 3: // right
			bx += math.Floor(b.W - line.W)
		}
		if !markup && li > 0 {
			buf.WriteByte('\n')
		}
		var x float64
		for i, sp := range line.Spans {
			w := math.Ceil(sp.W)
			if markup {
				of := of
				if sp.Tag != 0 {
					ofv := *of
					ofv.Style = sp.Tag
					of = &ofv
				}
				n.List = append(n.List, &Node{
					Kind: "text",
					Data: sp.Text,
					Calc: Box{
						Pos: Pos{X: bx + math.Ceil(x), Y: b.Y + y},
						Dim: Dim{W: w, H: lh},
					},
					Font: of,
				})
			} else {
				if i > 0 {
					buf.WriteByte(' ')
				}
				buf.WriteString(sp.Text)
			}
			if x+w > mw {
				mw = x + w
			}
			x += w
		}
		y += lh
	}
	if !markup {
		n.Data = buf.String()
	}
	b.H = math.Ceil(y)
	b.W = math.Ceil(mw)
	b = n.Pad.Outset(b)
	n.Calc.H = clamp(n.Calc.H, b.H)
	if n.W > 0 {
		n.Calc.W = clamp(n.Calc.W, n.W)
	} else {
		n.Calc.W = clamp(n.Calc.W, b.W)
	}
	n.Font = of
	return nil
}

func (l *Layouter) lineHeight(f *Font) (lh float64, _ error) {
	ff, err := l.Styler(l.Manager, *f, mark.Text)
	if err != nil {
		return 0, err
	}
	f.Height = ff.Metrics().Height

	if f.Line <= 0 {
		f.Line = 1.2
	}
	if f.Line < 8 {
		f.Line = math.Round(f.Line * l.PtToDot(f.Height))
	}
	return f.Line, nil
}

type splitter struct {
	*Layouter
	Font
	Max float64
}

func (s *splitter) lines(els []mark.El) (res []line, err error) {
	var cur line
	res = make([]line, 0, len(els)/8)
	for _, el := range els {
		// select face
		f, err := s.Styler(s.Manager, s.Font, el.Tag)
		if err != nil {
			return res, err
		}
		res, cur = s.spans(f, el.Tag, el.Cont, res, cur)
	}
	if len(cur.Spans) > 0 {
		res = append(res, cur.merge())
	}
	return res, nil
}

type line struct {
	Spans []span
	W     float64
}

func (l line) merge() line {
	var last *span
	res := make([]span, 0, len(l.Spans))
	for _, s := range l.Spans {
		if last != nil && s.Tag == last.Tag {
			last.Text += s.Text
			last.W += s.W
			continue
		}
		res = append(res, s)
		last = &res[len(res)-1]
	}
	return line{res, l.W}
}

type span struct {
	Text string
	W    float64
	Tag  mark.Tag
}

func (s *splitter) splitSpan(f *font.Face, txt string, mw float64) (w float64, _, rest string) {
	mp := s.DotToPt(mw)
	pt := f.Extra()
	last := rune(-1)
	for i, r := range txt {
		wr := f.Rune(r, last)
		if i > 0 && pt+wr > mp {
			return s.PtToDot(pt), txt[:i], txt[i:]
		}
		pt += wr
		last = r
	}
	return s.PtToDot(pt), txt, ""
}

func (s *splitter) spanW(f *font.Face, txt string) float64 {
	wpt, _ := f.Text(txt, -1)
	wpt += f.Extra()
	return math.Ceil(s.PtToDot(wpt))
}
func (s *splitter) spans(f *font.Face, tag mark.Tag, cont string, res []line, cur line) ([]line, line) {
	var space bool
	sdot := s.PtToDot(f.Rune(' ', -1))
	for _, txt := range toks(cont) {
		switch txt {
		case "":
			res = append(res, cur.merge())
			cur = line{}
			space = false
			continue
		case " ":
			space = true
			continue
		}
		ww := s.spanW(f, txt)
		var ws float64
		if space {
			ws = sdot
			space = false
		}
		mw := s.Max - cur.W
		if ww+ws < mw { // normal case: fits in cur line
			if ws > 0 {
				cur.Spans = append(cur.Spans, span{" ", ws, tag})
			}
			cur.W += math.Ceil(ws + ww)
			cur.Spans = append(cur.Spans, span{txt, ww, tag})
			continue
		}
		// check for soft break point
		if d := strings.IndexRune(txt, '-'); d > 0 {
			fst, snd := txt[:d+1], txt[d+1:]
			wf := s.spanW(f, fst)
			if ws+wf < mw {
				if ws > 0 {
					cur.Spans = append(cur.Spans, span{" ", ws, tag})
				}
				cur.W += math.Ceil(ws + wf)
				cur.Spans = append(cur.Spans, span{fst, wf, tag})
				ww, ws = s.spanW(f, snd), 0
				txt = snd
			}
		}
		// we nee to break the line
		// if the span does not fit the new line break inside the word until it does
		if ww > s.Max {
			i := 0
			for mw := s.Max - cur.W; ws+ww > mw; mw = s.Max {
				if i > 0 {
					if len(cur.Spans) > 0 {
						res = append(res, cur.merge())
					}
					cur = line{}
				}
				cw, ct, rest := s.splitSpan(f, txt, mw-ws)
				cur.W += math.Ceil(ws + cw)
				if ws > 0 {
					cur.Spans = append(cur.Spans, span{" ", ws, tag})
					ws = 0
				}
				cur.Spans = append(cur.Spans, span{ct, cw, tag})
				ww = s.spanW(f, rest)
				txt = rest
				i++
			}
		}
		if len(cur.Spans) > 0 {
			res = append(res, cur.merge())
		}
		cur = line{W: ww, Spans: []span{{txt, ww, tag}}}
	}
	if space {
		cur.Spans = append(cur.Spans, span{" ", sdot, tag})
		cur.W = sdot
	}
	return res, cur
}

func toks(text string) (res []string) {
	var start int
	var space bool
	for i, c := range text {
		if c == '\n' {
			if !space && i > start {
				res = append(res, text[start:i])
			}
			if space && len(res) > 0 && res[len(res)-1] == " " {
				res[len(res)-1] = ""
			} else {
				res = append(res, "")
			}
			space = true
			start = i + 1
		} else if cor.Space(c) {
			if !space {
				if i > start {
					res = append(res, text[start:i])
				}
				res = append(res, " ")
				space = true
			}
			start = i + 1
		} else if i > start {
			space = false
		}
	}
	if len(text) > start {
		res = append(res, text[start:])
	}
	return res
}
