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
	err = lay.collect(n, p)
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

func (l *layouter) collect(n *Node, p *pager) (err error) {
	var d *Node
	switch n.Kind {
	case "text", "block":
		d = &Node{Kind: n.Kind, Box: n.Calc, Font: n.Font, Data: n.Data}
	case "line":
		d = &Node{Kind: n.Kind, Box: n.Calc, Stroke: n.Stroke}
	case "qrcode", "barcode":
		d = &Node{Kind: n.Kind, Box: n.Calc, Code: n.Code, Data: n.Data}
	case "rect", "ellipse":
		d = &Node{Kind: n.Kind, Box: n.Calc, Stroke: n.Stroke}
		p.draw(d, n.Mar)
		fallthrough
	case "stage", "group", "vbox", "hbox", "table", "page":
		for _, e := range n.List {
			err = l.collect(e, p)
			if err != nil {
				return err
			}
		}
		return nil
	case "extra", "cover", "header", "footer":
		return nil
	}
	p.draw(d, n.Mar)
	return nil
}

// layout sets the calculated absolute box inside the available bounds a and returns
// the required area including margins.
// The passed in dimension can be unbounded vertically by setting h <= 0
func (l *layouter) layout(n *Node, a Box, stack []*Node) (_ Box, err error) {
	if a.W <= 0 {
		return n.Calc, cor.Errorf("measure always needs availible width")
	}
	m := getMargin(n)
	ab := m.Inset(a)
	nb := Box{Pos: ab.Pos, Dim: n.Dim}
	nb.W = clampFill(ab.W, nb.W)
	nb.H = clamp(ab.H, nb.H)
	n.Calc = nb
	switch n.Kind {
	case "text", "block":
		err = l.textLayout(n, stack)
	case "line":
	case "qrcode":
		if nb.H == 0 || nb.W < nb.H {
			n.Calc.H = nb.W
		} else if nb.H > 0 && nb.W > nb.H {
			n.Calc.W = nb.H
		}
	case "barcode":
	case "group", "rect", "ellipse":
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
	cwpt := n.Calc.W * 72 / (25.4 * 8)
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
		lh = lf * float64(h) * 25.4 * 8 / 72
	}
	n.Calc.H = clamp(n.Calc.H, math.Ceil(lh*float64(len(txt))))
	n.Calc.W = clamp(n.Calc.W, math.Ceil(float64(mw)*25.4*8/72))
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
	var w, h float64
	for i, e := range n.List {
		ab := a
		if n.Sub.H > 0 && e.H <= 0 {
			e.H = n.Sub.H
		}
		eb, err := l.layout(e, ab, stack)
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
		if eb.W > w {
			w = eb.W
		}
	}
	if n.Calc.W <= 0 {
		n.Calc.W = clamp(n.Calc.W, w)
	}
	if n.Calc.H <= 0 {
		n.Calc.H = clamp(n.Calc.H, h)
	}
	return nil
}
func (l *layouter) hboxLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
	a := n.Pad.Inset(n.Calc)
	var w, h float64
	for i, e := range n.List {
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
	if n.Calc.W <= 0 {
		n.Calc.W = clamp(n.Calc.W, w)
	}
	if n.Calc.H <= 0 {
		n.Calc.H = clamp(n.Calc.H, h)
	}
	return nil
}
func (l *layouter) tableLayout(n *Node, stack []*Node) error {
	stack = append(stack, n)
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
	a := n.Calc
	var h, rh float64
	for i, e := range n.List {
		a.X = n.Calc.X
		ci := i % len(n.Cols)
		if ci == 0 && i > 0 {
			if i < len(n.List)-1 {
				rh += n.Gap
			}
			a.Y += rh
			a.H -= rh
			h += rh
			rh = 0
		} else {
			for _, c := range n.Cols[:ci] {
				a.X += c
			}
		}
		a.W = n.Cols[ci]
		eb, err := l.layout(e, a, stack)
		if err != nil {
			return err
		}
		if eb.H > rh {
			rh = eb.H
		}
	}
	h += rh
	if n.Calc.H <= 0 {
		n.Calc.H = clamp(n.Calc.H, h)
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
	f := Font{"", 0, 0}
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
