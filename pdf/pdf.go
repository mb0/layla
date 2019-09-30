// Package pdf implements a layla renderer for PDF using the jung-kurt/gofpdf package.
package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"

	"github.com/boombuler/barcode"
	"github.com/jung-kurt/gofpdf"
	"github.com/mb0/layla"
	"github.com/mb0/layla/bcode"
	"github.com/mb0/layla/font"
	"github.com/mb0/xelf/cor"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Doc = gofpdf.Fpdf

func NewDoc(n *layla.Node) *Doc {
	return gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{n.W / 8, n.H / 8},
	})
}
func NewA4() *Doc { return gofpdf.New("P", "mm", "A4", "") }

func Render(m *font.Manager, n *layla.Node) (*Doc, error) {
	return Renderer{m, nil}.RenderTo(NewDoc(n), n)
}

type colorhack struct{ image.Image }

func (colorhack) ColorModel() color.Model { return color.GrayModel }

func DefaultBarcoder(n *layla.Node) (image.Image, error) {
	bc, err := bcode.Barcode(n)
	if err != nil {
		return nil, err
	}
	bc, err = barcode.Scale(bc, int(n.W), int(n.H))
	if err != nil {
		return nil, fmt.Errorf("%v %g %g", err, n.W, n.H)
	}
	return colorhack{bc}, nil
}

type Renderer struct {
	*font.Manager
	Barcoder func(*layla.Node) (image.Image, error)
}

func (r Renderer) RenderTo(d *Doc, n *layla.Node) (*Doc, error) {
	return r.RenderSubjTo(d, n, "")
}

func (r Renderer) RenderSubjTo(d *Doc, n *layla.Node, subj string) (*Doc, error) {
	d.AddPage()
	if subj != "" {
		d.Bookmark(subj, 0, 0)
	}
	draw, err := layla.Layout(r.Manager, n)
	if err != nil {
		return nil, err
	}
	r.addFonts(d, draw)
	for _, dn := range draw {
		err = r.renderNode(d, dn)
		if err != nil {
			return nil, err
		}
	}
	return d, d.Error()
}

func (r Renderer) addFonts(d *Doc, ns []*layla.Node) error {
	fs := make(map[string]bool, 8)
	for _, n := range ns {
		if n.Font == nil {
			continue
		}
		if fs[n.Font.Name] {
			continue
		}
		path, err := r.Path(n.Font.Name)
		if err != nil {
			return err
		}
		dir, fname := filepath.Split(path)
		d.SetFontLocation(dir)
		ext := filepath.Ext(fname)
		descf := fmt.Sprintf("%s.json", fname[:len(fname)-len(ext)])
		d.AddFont(n.Font.Name, "", descf)
		fs[n.Font.Name] = true
	}
	return nil
}

func setupBorder(d *Doc, bw float64, c *layla.Color) float64 {
	bw = bw / 8
	d.SetLineWidth(bw)
	if c == nil {
		d.SetDrawColor(0, 0, 0)
	} else {
		d.SetDrawColor(c.R, c.G, c.B)
	}
	return bw
}
func drawBorder(d *Doc, b layla.Box, br layla.Border, c *layla.Color) {
	if br == (layla.Border{}) {
		return
	}
	x1, y1 := b.X/8, b.Y/8
	x2, y2 := (b.X+b.W)/8, (b.Y+b.H)/8
	if br.L > 0 {
		bw := setupBorder(d, br.L, c) / 2
		d.Line(x1+bw, y1, x1+bw, y2)
	}
	if br.T > 0 {
		bw := setupBorder(d, br.T, c) / 2
		d.Line(x1, y1+bw, x2, y1+bw)
	}
	if br.R > 0 {
		bw := setupBorder(d, br.R, c) / 2
		d.Line(x2-bw, y1, x2-bw, y2)
	}
	if br.B > 0 {
		bw := setupBorder(d, br.B, c) / 2
		d.Line(x1, y2-bw, x2, y2-bw)
	}
}

func (r Renderer) renderNode(d *Doc, n *layla.Node) error {
	switch n.Kind {
	case "ellipse":
		b := n.Border.Default(1.6)
		d.SetLineWidth(b.W / 8)
		rx, ry := n.W/16, n.H/16
		d.Ellipse(n.X/8+rx, n.Y/8+ry, rx, ry, 0, "D")
	case "line":
		b := n.Border.Default(1.6)
		d.SetDrawColor(0, 0, 0)
		d.SetLineWidth(b.W / 8)
		x, y := n.X/8, n.Y/8
		d.Line(x, y, x+n.W/8, y+n.H/8)
	case "rect":
		b := n.Border.Default(1.6)
		drawBorder(d, n.Box, b, nil)
	case "text":
		br := n.Border.Default(0)
		drawBorder(d, n.Box, br, nil)
		d.SetFont(n.Font.Name, "", n.Font.Size)
		b := n.Pad.Inset(n.Box)
		res, err := enc(n.Data)
		if err != nil {
			return err
		}
		d.SetXY((b.X-8)/8, b.Y/8)
		_, lh := d.GetFontSize()
		if n.Font.Line > 0 {
			lh *= n.Font.Line
		} else {
			lh *= 1.2
		}
		align := "LB"
		switch n.Align {
		case layla.AlignRight:
			align = "RB"
		case layla.AlignCenter:
			align = "CB"
		}
		d.MultiCell((b.W+16)/8, lh, res, "", align, false)
	case "barcode", "qrcode":
		coder := r.Barcoder
		if coder == nil {
			coder = DefaultBarcoder
		}
		bc, err := coder(n)
		if err != nil {
			return err
		}
		var b bytes.Buffer
		err = png.Encode(&b, bc)
		if err != nil {
			return err
		}
		name := n.Kind + ":" + n.Data
		iopt := gofpdf.ImageOptions{ImageType: "PNG"}
		d.RegisterImageOptionsReader(name, iopt, &b)
		d.ImageOptions(name, n.X/8, n.Y/8, n.W/8, n.H/8, false, iopt, 0, "")
	case "page":
		d.AddPage()
	default:
		return cor.Errorf("unexpected node kind %q", n.Kind)
	}
	return nil
}

var win1252Enc = charmap.Windows1252.NewEncoder()

func enc(str string) (string, error) {
	res, _, err := transform.String(win1252Enc, str)
	return res, err
}
