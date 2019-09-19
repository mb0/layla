package layla

import (
	"math"
	"strings"
)

type xpage struct {
	Org float64
	Box
	res   []*Node
	page  string
	total string
}

func (x *xpage) collect(n *Node, res []*Node, offy float64) []*Node {
	var d *Node
	switch n.Kind {
	case "text", "block":
		d = &Node{Kind: n.Kind, Box: n.Calc, Font: n.Font, Data: n.Data}
		d.Data = strings.ReplaceAll(d.Data, "µP", x.page)
		d.Data = strings.ReplaceAll(d.Data, "µT", x.total)
	case "line":
		d = &Node{Kind: n.Kind, Box: n.Calc, Stroke: n.Stroke}
	case "qrcode", "barcode":
		d = &Node{Kind: n.Kind, Box: n.Calc, Code: n.Code, Data: n.Data}
	case "rect", "ellipse":
		d = &Node{Kind: n.Kind, Box: n.Calc, Stroke: n.Stroke}
		d.Y += offy
		res = append(res, d)
		fallthrough
	case "stage", "group", "vbox", "hbox", "table", "page",
		"extra", "cover", "header", "footer":
		for _, e := range n.List {
			res = x.collect(e, res, offy)
		}
		return res
	}
	d.Y += offy
	return append(res, d)
}

type pager struct {
	*Node
	Extra  *Node
	Cover  *Node
	Header *Node
	Footer *Node
	list   []*xpage
}

func newPager(n *Node) *pager {
	p := &pager{Node: n}
	for _, e := range n.List {
		switch e.Kind {
		case "extra":
			p.Extra = e
		case "cover":
			p.Cover = e
		case "header":
			p.Header = e
		case "footer":
			p.Footer = e
		}
	}
	p.newPage(0)
	return p
}

func (p *pager) newPage(org float64) *xpage {
	b := p.Pad.Inset(Box{Dim: p.Dim})
	top := p.Cover
	if len(p.list) > 0 {
		top = p.Header
	}
	if top != nil {
		h := top.Calc.H
		b.Y += h
		b.H -= b.Y
	}
	if p.Footer != nil {
		b.H -= p.Footer.Calc.H
	}
	x := &xpage{Org: org, Box: b}
	p.list = append(p.list, x)
	return x
}

func (p *pager) draw(n *Node, m *Off) {
	if p.Kind != "page" {
		xp := p.list[0]
		xp.res = append(xp.res, n)
		return
	}
	// find starting page
	for i := len(p.list) - 1; i >= 0; i-- {
		x := p.list[i]
		if x.Org > n.Y {
			continue
		}
		y := n.Y - x.Org
		// simple case fits into the remaining space
		if y+n.H <= x.H {
			n.Y = x.Y + y
			x.res = append(x.res, n)
			return
		}
		switch n.Kind {
		case "text", "block":
			txt := strings.Split(n.Data, "\n")
			lh := n.H / float64(len(txt))
			ah := x.H - y
			hh := 0.0
			for j := 0; len(txt) > 0; j++ {
				lc := int(ah / lh)
				if lc == 0 && j > 0 {
					lc = 1
				}
				if lc > len(txt) {
					lc = len(txt)
				}
				if lc > 0 {
					nn := *n
					nn.Y = x.Y + y
					nn.H = math.Ceil(lh * float64(lc))
					hh += nn.H
					nn.Data = strings.Join(txt[:lc], "\n")
					x.res = append(x.res, &nn)
					txt = txt[lc:]
				}
				if len(txt) == 0 {
					return
				}
				i++
				if i < len(p.list) {
					x = p.list[i]
				} else {
					x = p.newPage(n.Y + hh)
				}
				y = 0
				if m != nil {
					y = m.T
				}
				ah = x.H
			}
			return
		}
		i++
		if i < len(p.list) {
			x = p.list[i]
		} else {
			x = p.newPage(n.Y)
		}
		n.Y = x.Y
		if m != nil {
			n.Y += m.T
		}
		x.res = append(x.res, n)
		return
	}
}
