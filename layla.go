package layla

// Pos is a simple position consisting of x and y coordinates in mm.
type Pos struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
}

// Dim is a simple dimension consisting of width and height in mm.
type Dim struct {
	W float64 `json:"w,omitempty"`
	H float64 `json:"h,omitempty"`
}

// Box is a simple box consisting of a position and dimension.
type Box struct {
	Pos
	Dim
}

// Off is a box offset consisting of left, top, right and bottom offsets in mm.
type Off struct {
	L float64 `json:"l,omitempty"`
	T float64 `json:"t,omitempty"`
	R float64 `json:"r,omitempty"`
	B float64 `json:"b,omitempty"`
}

// Inset returns a box result of b with o substracted.
func (o *Off) Inset(b Box) Box {
	if o != nil {
		b.X += o.L
		b.Y += o.T
		b.W -= o.L + o.R
		b.H -= o.T + o.B
		if b.W < 0 {
			b.W = 0
		}
		if b.H < 0 {
			b.H = 0
		}
	}
	return b
}

// Outset returns a box result of b with o added.
func (o *Off) Outset(b Box) Box {
	if o != nil {
		b.X -= o.L
		b.Y -= o.T
		b.W += o.L + o.R
		b.H += o.T + o.B
	}
	return b
}

// Font holds font all font data related node data
type Font struct {
	Name string  `json:"name,omitempty"`
	Size float64 `json:"size,omitempty"`
	Line float64 `json:"line,omitempty"`
}

// NodeLayout holds all layout related node data
type NodeLayout struct {
	Gap   float64   `json:"gap,omitempty"`
	Mar   *Off      `json:"mar,omitempty"`
	Pad   *Off      `json:"pad,omitempty"`
	Rot   int       `json:"rot,omitempty"`
	Align int       `json:"align,omitempty"`
	Sub   Dim       `json:"sub,omitempty"`
	Cols  []float64 `json:"cols,omitempty"`
}

// Code holds all qr and barcode related node data
type Code struct {
	Name  string  `json:"name,omitempty"`
	Human int     `json:"human,omitempty"`
	Wide  float64 `json:"wide,omitempty"`
}

// Node is a part of the display tree represents all display elements.
type Node struct {
	Kind string `json:"kind"`
	Box
	NodeLayout
	Font   *Font   `json:"font,omitempty"`
	Stroke float64 `json:"stroke,omitempty"`
	Code   *Code   `json:"code,omitempty"`
	Data   string  `json:"data,omitempty"`
	List   []*Node `json:"list,omitempty"`
	Calc   Box     `json:"-"`
}
