package layla

// Pos is a simple position consisting of x and y coordinates in the base unit.
type Pos struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

// Dim is a simple dimension consisting of width and height in the base unit.
type Dim struct {
	W int `json:"w,omitempty"`
	H int `json:"h,omitempty"`
}

// Box is a simple box consisting of a position and dimension.
type Box struct {
	Pos
	Dim
}

// Off is a box offset consisting of left, top, right and bottom offsets in base unit.
type Off struct {
	L int `json:"l,omitempty"`
	T int `json:"t,omitempty"`
	R int `json:"r,omitempty"`
	B int `json:"b,omitempty"`
}

// Apply returns a box result of b subtracted by o.
func (o *Off) Apply(b Box) Box {
	if o != nil {
		b.X += o.L
		b.Y += o.T
		b.W -= o.L + o.R
		if b.W < 0 {
			b.W = 0
		}
		b.H -= o.T + o.B
		if b.H < 0 {
			b.H = 0
		}
	}
	return b
}

// Font holds font all font data related node data
type Font struct {
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"`
}

// NodeLayout holds all layout related node data
type NodeLayout struct {
	Gap   int   `json:"gap,omitempty"`
	Mar   *Off  `json:"mar,omitempty"`
	Pad   *Off  `json:"pad,omitempty"`
	Rot   int   `json:"rot,omitempty"`
	Align int   `json:"align,omitempty"`
	Sub   Dim   `json:"sub,omitempty"`
	Cols  []int `json:"cols,omitempty"`
}

// Code holds all qr and barcode related node data
type Code struct {
	Name   string `json:"name,omitempty"`
	Human  int    `json:"human,omitempty"`
	Narrow int    `json:"narrow,omitempty"`
	Wide   int    `json:"wide,omitempty"`
}

// Node is a part of the display tree represents all display elements.
type Node struct {
	Kind string `json:"kind"`
	Box
	NodeLayout
	Font *Font   `json:"font,omitempty"`
	Line int     `json:"line,omitempty"`
	Code *Code   `json:"code,omitempty"`
	Data string  `json:"data,omitempty"`
	List []*Node `json:"list,omitempty"`
}
