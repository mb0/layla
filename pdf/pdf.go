// Package pdf implements a layla renderer for PDF using the jung-kurt/gofpdf package.
package pdf

import (
	"bytes"
	"fmt"
	"image/gif"
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

func Render(m *font.Manager, n *layla.Node) (*Doc, error) {
	return RenderTo(NewDoc(n), m, n)
}

func RenderTo(d *Doc, m *font.Manager, n *layla.Node) (*Doc, error) {
	d.AddPage()
	draw, err := layla.Layout(m, n)
	if err != nil {
		return nil, err
	}
	addFonts(m, d, draw)
	for _, dn := range draw {
		err = renderNode(m, d, dn)
		if err != nil {
			return nil, err
		}
	}
	return d, d.Error()
}

func addFonts(m *font.Manager, d *Doc, ns []*layla.Node) error {
	fs := make(map[string]bool, 8)
	for _, n := range ns {
		if n.Font == nil {
			continue
		}
		if fs[n.Font.Name] {
			continue
		}
		path, err := m.Path(n.Font.Name)
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

func renderNode(m *font.Manager, d *Doc, n *layla.Node) error {
	switch n.Kind {
	case "ellipse":
		rx, ry := n.W/16, n.H/16
		d.Ellipse(n.X/8+rx, n.Y/8+ry, rx, ry, 0, "D")
	case "line":
		d.SetDrawColor(0, 0, 0)
		d.SetLineWidth(n.Stroke / 8)
		d.SetLineCapStyle("square")
		x, y := n.X/8, n.Y/8
		d.Line(x, y, x+n.W/8, y+n.H/8)
	case "rect":
		d.Rect(n.X/8, n.Y/8, n.W/8, n.H/8, "D")
	case "text", "block":
		d.SetFont(n.Font.Name, "", n.Font.Size)
		res, err := enc(n.Data)
		if err != nil {
			return err
		}
		d.SetXY(n.X/8, n.Y/8)
		_, lh := d.GetFontSize()
		if n.Font.Line > 0 {
			lh *= n.Font.Line
		}
		d.MultiCell(n.W/8, lh, res, "", "LA", false)
		// d.Rect(n.X/8, n.Y/8, n.W/8, n.H/8, "D")
	case "barcode", "qrcode":
		bc, err := bcode.Barcode(n)
		if err != nil {
			return err
		}
		bc, err = barcode.Scale(bc, int(n.W), int(n.H))
		if err != nil {
			return fmt.Errorf("%v %g %g", err, n.W, n.H)
		}
		var b bytes.Buffer
		err = gif.Encode(&b, bc, nil)
		if err != nil {
			return err
		}
		name := n.Kind + ":" + n.Data
		iopt := gofpdf.ImageOptions{ImageType: "GIF"}
		d.RegisterImageOptionsReader(name, iopt, bytes.NewReader(b.Bytes()))
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
