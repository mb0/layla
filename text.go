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
	lh, ws, err := l.lineHeight(of)
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
		// TODO handle alignment
		if !markup && li > 0 {
			buf.WriteByte('\n')
		}
		var x float64
		for i, sp := range line.Spans {
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
						Pos: Pos{X: b.X + x, Y: b.Y + y},
						Dim: Dim{W: sp.W, H: lh},
					},
					Font: of,
				})
			} else {
				if i > 0 {
					buf.WriteByte(' ')
				}
				buf.WriteString(sp.Text)
			}
			if x+sp.W > mw {
				mw = x + sp.W
			}
			x += sp.W + ws
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
	n.Calc.W = clamp(n.Calc.W, b.W)
	n.Font = of
	return nil
}

func (l *Layouter) lineHeight(f *Font) (lh, ws float64, _ error) {
	ff, err := l.Styler(l.Manager, *f, mark.Text)
	if err != nil {
		return 0, 0, err
	}
	f.Height = ff.Metrics().Height

	if f.Line <= 0 {
		f.Line = 1.2
	}
	if f.Line < 8 {
		f.Line = math.Ceil(f.Line * l.PtToDot(f.Height))
	}
	wpt := ff.Rune(' ', -1)
	return f.Line, l.PtToDot(wpt), nil
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
		res = append(res, cur)
	}
	return res, nil
}

type line struct {
	Spans []span
	W     float64
}

func (l line) merge(ws float64) line {
	var last *span
	res := make([]span, 0, len(l.Spans))
	for _, s := range l.Spans {
		if last != nil && s.Tag == last.Tag {
			last.Text += " " + s.Text
			last.W += ws + s.W
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

func (s *splitter) spanW(f *font.Face, txt string, space bool, ws float64) (ww, w float64) {
	wpt, _ := f.Text(txt, -1)
	wpt += f.Extra()
	ww = s.PtToDot(wpt)
	w = ww
	if space {
		w += ws
	}
	return ww, w
}
func (s *splitter) spans(f *font.Face, tag mark.Tag, cont string, res []line, cur line) ([]line, line) {
	ws := s.PtToDot(f.Rune(' ', -1))
	for _, txt := range toks(cont) {
		if txt == "" {
			res = append(res, cur.merge(ws))
			cur = line{}
			continue
		}
		space := len(cur.Spans) > 0
		ww, w := s.spanW(f, txt, space, ws)
		mw := s.Max - cur.W
		if w < mw { // normal case: fits in cur line
			cur.W += w
			cur.Spans = append(cur.Spans, span{txt, ww, tag})
			continue
		}
		// check for soft break point
		if d := strings.IndexRune(txt, '-'); d > 0 {
			fst, snd := txt[:d+1], txt[d+1:]
			fww, fw := s.spanW(f, fst, space, ws)
			if fw < mw {
				cur.W += fw
				cur.Spans = append(cur.Spans, span{fst, fww, tag})
				ww, w = s.spanW(f, snd, false, ws)
				txt = snd
			}
		}
		// we nee to break the line
		// if the span does not fit the new line break inside the word until it does
		if ww > s.Max {
			i := 0
			for mw := s.Max - cur.W; ww > mw; mw = s.Max {
				if i > 0 {
					if len(cur.Spans) > 0 {
						res = append(res, cur.merge(ws))
					}
					cur = line{}
				}
				cw, ct, rest := s.splitSpan(f, txt, mw)
				cur.W += cw
				cur.Spans = append(cur.Spans, span{ct, cw, tag})
				ww, w = s.spanW(f, rest, false, ws)
				txt = rest
				i++
			}
		}
		if len(cur.Spans) > 0 {
			res = append(res, cur.merge(ws))
		}
		cur = line{W: ww, Spans: []span{{txt, ww, tag}}}
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
			res = append(res, "")
			space = false
			start = i + 1
		} else if cor.Space(c) {
			if !space {
				if i > start {
					res = append(res, text[start:i])
				}
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
