package layla

import (
	"fmt"
	"math"
	"strings"

	"github.com/mb0/layla/font"
	"github.com/mb0/xelf/cor"
)

// Layout returns a slice of visible nodes with the layouts applied or an error
func Layout(m *font.Manager, n *Node) ([]*Node, error) {
	lay := newLayouter(m, n)
	_, err := lay.layout(n, n.Box, nil)
	if err != nil {
		return nil, err
	}
	p := newPager(n)
	err = p.collect(n)
	if err != nil {
		return nil, err
	}
	var res []*Node
	total := fmt.Sprint(len(p.list))
	for i, x := range p.list {
		if i > 0 {
			res = append(res, &Node{Kind: "page"})
		}
		x.page = fmt.Sprint(i + 1)
		x.total = total
		if p.Extra != nil {
			res = x.collect(p.Extra, res, 0)
		}
		top := p.Cover
		if i > 0 {
			top = p.Header
		}
		if top != nil {
			res = x.collect(top, res, 0)
		}
		if p.Footer != nil {
			offy := x.Y + x.H
			res = x.collect(p.Footer, res, offy)
		}
		res = append(res, x.res...)
	}
	return res, nil
}

// layouter implements the layout routine and holds required context
type layouter struct {
	*font.Manager
	overflow []*Node
}

func newLayouter(m *font.Manager, o *Node) *layouter {
	return &layouter{Manager: m}
}

// layout sets the calculated absolute box inside the available bounds a and returns
// the required area including margins.
// The passed in dimension can be unbounded vertically by setting h <= 0
func (l *layouter) layout(n *Node, a Box, stack []*Node) (_ Box, err error) {
	if a.W <= 0 {
		return n.Calc, cor.Errorf("layout always needs availible width")
	}
	m := getMargin(n)
	ab := m.Inset(a)
	nb := Box{Pos: ab.Pos, Dim: n.Dim}
	nb.W = clampFill(ab.W, nb.W)
	if nb.W < ab.W {
		switch n.Align {
		case AlignRight:
			nb.X += math.Floor(ab.W - nb.W)
		case AlignCenter:
			nb.X += math.Ceil((ab.W - nb.W) / 2)
		}
	}
	nb.H = clamp(ab.H, nb.H)
	n.Calc = nb
	switch n.Kind {
	case "text":
		err = l.textLayout(n, stack)
	case "line":
		n.Calc.W = n.W
	case "qrcode":
		if nb.H == 0 || nb.W < nb.H {
			n.Calc.H = nb.W
		} else if nb.H > 0 && nb.W > nb.H {
			n.Calc.W = nb.H
		}
	case "barcode":
	case "box", "rect", "ellipse":
		n.Calc.H = clampFill(ab.H, nb.H)
		err = l.freeLayout(n, stack)
	case "extra", "cover", "header", "footer":
		err = l.freeLayout(n, stack)
	case "stage":
		err = l.freeLayout(n, stack)
	case "page":
		n.Calc.H = 0
		err = l.freeLayout(n, stack)
	case "vbox":
		err = l.vboxLayout(n, stack)
	case "hbox":
		err = l.hboxLayout(n, stack)
	case "table":
		err = l.tableLayout(n, stack)
	}
	if err != nil {
		return Box{}, err
	}
	return m.Outset(n.Calc), nil
}
func (l *layouter) textLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
	of := getFont(stack)
	f, err := l.Face(of.Name, of.Size)
	if err != nil {
		return cor.Errorf("%v for %q\n\t%d", err, n.Data, len(stack))
	}
	b := n.Pad.Inset(n.Calc)
	cwpt := b.W * 72 / (25.4 * 8)
	txt, h, mw, err := font.Layout(f, n.Data, int(cwpt))
	if err != nil {
		return err
	}
	lf := of.Line
	if lf <= 0 {
		lf = 1.2
	}
	lh := lf
	if lh < 8 {
		lh = math.Ceil(lf * float64(h) * 25.4 * 8 / 72)
		of.Line = lh
	}
	b.H = lh * float64(len(txt))
	b.W = math.Ceil(float64(mw) * 25.4 * 8 / 72)
	b = n.Pad.Outset(b)
	n.Calc.H = clamp(n.Calc.H, b.H)
	n.Calc.W = clamp(n.Calc.W, b.W)
	n.Data = strings.Join(txt, "\n")
	n.Font = of
	return nil
}

func (l *layouter) freeLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
	a := n.Pad.Inset(n.Calc)
	var h float64
	for _, e := range n.List {
		eb, err := l.layout(e, a, stack)
		if err != nil {
			return err
		}
		if y := eb.Y + eb.H; y > h {
			h = y
		}
	}
	if n.Mar != nil {
		h += n.Mar.B
	}
	if n.Calc.H <= 0 {
		n.Calc.H = h
	}
	return nil
}
func (l *layouter) vboxLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
	a := n.Pad.Inset(n.Calc)
	var h float64
	for i, e := range n.List {
		if n.Sub.H > 0 && e.H <= 0 {
			e.H = n.Sub.H
		}
		max := a.W
		if e.Mar != nil {
			max -= e.Mar.L + e.Mar.R
		}
		if e.W > max {
			e.W = max
		}
		eb, err := l.layout(e, a, stack)
		if err != nil {
			return err
		}
		y := eb.H
		if i < len(n.List)-1 {
			y += n.Gap
		}
		a.Y += y
		a.H -= y
		h += y
		if e.W > 0 {
			e.Calc.W = e.W
		} else {
			e.Calc.W = max
		}
	}
	n.Calc.H = clamp(n.Calc.H, h)
	return nil
}
func (l *layouter) hboxLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
	a := n.Pad.Inset(n.Calc)
	var w, h float64
	for i, e := range n.List {
		if n.Sub.W > 0 && e.W <= 0 {
			e.W = n.Sub.W
		}
		max := a.H
		if e.Mar != nil {
			max -= e.Mar.T + e.Mar.B
		}
		if e.H > max {
			e.H = max
		}
		eb, err := l.layout(e, a, stack)
		if err != nil {
			return err
		}
		x := eb.W
		if i < len(n.List)-1 {
			x += n.Gap
		}
		a.X += x
		a.W -= x
		w += x
		if eb.H > h {
			h = eb.H
		}
	}
	n.Calc.W = clamp(n.Calc.W, w)
	return nil
}
func tableCols(n *Node) {
	aw := n.Calc.W
	nw := 0.0
	for _, c := range n.Cols {
		if c <= 0 {
			nw++
		} else {
			aw -= c
		}
	}
	if nw > 0 {
		for i, c := range n.Cols {
			if c <= 0 {
				n.Cols[i] = aw / nw
			}
		}
		aw = 0
	}
	if aw > 0 {
		n.Calc.W -= aw
	}
}

func (l *layouter) tableLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
	tableCols(n)
	a := n.Calc
	for i := 0; i < len(n.List); i += len(n.Cols) {
		r := n.List[i:]
		if len(n.Cols) < len(r) {
			r = r[:len(n.Cols)]
		}
		var rw, rh float64
		for i, c := range r {
			b := a
			b.X += rw
			b.W = n.Cols[i]
			rw += b.W
			eb, err := l.layout(c, b, stack)
			if err != nil {
				return err
			}
			c.Calc.W = b.W
			if eb.H > rh {
				rh = c.Calc.H
			}
		}
		for _, c := range r {
			c.Calc.H = rh
		}
		rh += n.Gap
		a.Y += rh
		a.H -= rh
	}
	if n.Calc.H <= 0 {
		n.Calc.H = clamp(n.Calc.H, a.Y-n.Calc.Y)
	}
	return nil
}

func clampFill(a, c float64) float64 {
	if a > 0 && c > a || c <= 0 {
		return a
	}
	return c
}

func clamp(a, c float64) float64 {
	if a > 0 && c > a {
		return a
	}
	return c
}

func getMargin(n *Node) Off {
	var m Off
	if n.Mar != nil {
		m = *n.Mar
	}
	if n.X > 0 {
		m.L = n.X
	}
	if n.Y > 0 {
		m.T = n.Y
	}
	return m
}

func getFont(stack []*Node) *Font {
	f := Font{}
	for i := len(stack) - 1; i >= 0; i-- {
		nf := stack[i].Font
		if nf == nil {
			continue
		}
		if f.Name == "" {
			f.Name = nf.Name
		}
		if f.Size == 0 {
			f.Size = nf.Size
		}
		if f.Line == 0 {
			f.Line = nf.Line
		}
		if f.Size != 0 && f.Name != "" && f.Line != 0 {
			break
		}
	}
	return &f
}
