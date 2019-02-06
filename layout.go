package layla

import "fmt"

// Layout returns a slice of visible nodes with the layouts applied or an error
func Layout(o *Node) ([]*Node, error) {
	lay := &layouter{draw: make([]*Node, 0, 8)}
	err := lay.layout(o, Box{Dim: o.Dim}, nil)
	if err != nil {
		return nil, err
	}
	return lay.draw, nil
}

// layouter implements the layout routine and holds required context
type layouter struct {
	draw []*Node
}

// layout a node with child nodes. all data must already be resolved
func (l *layouter) layout(o *Node, b Box, stack []*Node) error {
	// subtract child margin from free box to get the area box
	a := o.Mar.Apply(b)
	c := o.Box
	if c.X > 0 {
		a.X += c.X
	}
	if c.Y > 0 {
		a.Y += c.Y
	}
	// w is clamped to or fills available w
	if a.W > 0 && (c.X+c.W > a.W || c.W <= 0) {
		c.W = a.W - c.X
	}
	// h is clamped to or fills available h
	if a.H > 0 && (c.H+c.Y > a.H || c.H <= 0) {
		c.H = a.H - c.Y
	}
	c = Box{a.Pos, c.Dim}
	switch o.Kind {
	case "text", "block":
		l.draw = append(l.draw, &Node{Kind: o.Kind, Box: c, Font: o.Font, Data: o.Data})
		return nil
	case "qrcode":
		if c.H < c.W {
			c.W = c.H
		} else {
			c.H = c.W
		}
		fallthrough
	case "barcode":
		l.draw = append(l.draw, &Node{Kind: o.Kind, Box: c, Code: o.Code, Data: o.Data})
		return nil
	case "rect", "ellipse":
		l.draw = append(l.draw, &Node{Kind: o.Kind, Box: c, Line: o.Line})
	case "stage":
		if c.W == 0 || c.H == 0 {
			return fmt.Errorf("stage need both width and height set")
		}
	}
	// substract the padding to get the child area box
	a = o.Pad.Apply(c)
	for i, e := range o.List {
		// layout child box according to layout and available box
		switch o.Kind {
		case "stage", "group": // free layout?
			l.layout(e, a, append(stack, o))
		case "vbox": // vertical layout
			// child h is respected or falls back parent sub h if available
			if e.H <= 0 {
				e.H = o.Sub.H
			}
			l.layout(e, a, append(stack, o))
			a.Y += e.H + o.Gap
			a.H -= e.H + o.Gap
		case "hbox": // horizontal layout
			// child w is respected or falls back to parent sub w if available
			if e.W <= 0 {
				e.W = o.Sub.W
			}
			l.layout(e, a, append(stack, o))
			a.X += e.W + o.Gap
			a.W -= e.W + o.Gap
		case "table":
			ca := a
			cols := o.Cols
			if len(cols) == 0 {
				cols = []float64{ca.W}
			}
			ci := i % len(cols)
			for _, cw := range cols[:ci] {
				ca.X += cw
			}
			ca.H = o.Sub.H
			ca.W = cols[ci]
			l.layout(e, ca, append(stack, o))
			if ci == len(cols)-1 {
				a.Y += ca.H + o.Gap
				a.H -= ca.H + o.Gap
			}
		}
	}
	switch o.Kind {
	case "vbox":
		if o.H <= 0 {
			o.H = a.Y - o.Gap
		}
	case "table":
		if o.W <= 0 {
			o.W = 0
			for _, cw := range o.Cols {
				c.W += cw
			}
		}
		if o.H <= 0 {
			o.H = a.Y - o.Gap
		}
	case "hbox":
		if o.W <= 0 {
			o.W = a.X - o.Gap
		}
	}
	return nil
}
