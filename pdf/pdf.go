// Package pdf implements a layla renderer for PDF using the jung-kurt/gofpdf package.
package pdf

import (
	"bytes"
	"fmt"
	"image/gif"

	"github.com/boombuler/barcode"
	"github.com/jung-kurt/gofpdf"
	"github.com/mb0/layla"
	"github.com/mb0/layla/bcode"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Doc = gofpdf.Fpdf

func Render(n *layla.Node) (*Doc, error) {
	d := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{n.W, n.H},
	})
	d.AddPage()
	d.SetFont("Helvetica", "", 8)
	draw, err := layla.Layout(n)
	if err != nil {
		return nil, err
	}
	for _, dn := range draw {
		err = renderNode(d, dn)
		if err != nil {
			return nil, err
		}
	}
	return d, d.Error()
}

func renderNode(d *Doc, n *layla.Node) error {
	switch n.Kind {
	case "ellipse":
		rx, ry := n.W/2, n.H/2
		d.Ellipse(n.X+rx, n.Y+ry, rx, ry, 0, "D")
	case "rect":
		d.Rect(n.X, n.Y, n.W, n.H, "D")
	case "text", "block", "styled":
		res, err := enc(n.Data)
		if err != nil {
			return err
		}
		d.SetXY(n.X, n.Y)
		d.MultiCell(n.W, 2.5, res, "", "LA", false)
	case "barcode", "qrcode":
		bc, err := bcode.Barcode(n)
		if err != nil {
			return err
		}
		bc, err = barcode.Scale(bc, int(n.W*8), int(n.H*8))
		if err != nil {
			return fmt.Errorf("%v %g %g", err, n.W, n.H)
			//return err
		}
		var b bytes.Buffer
		err = gif.Encode(&b, bc, nil)
		if err != nil {
			return err
		}
		name := n.Kind + ":" + n.Data
		iopt := gofpdf.ImageOptions{ImageType: "GIF"}
		d.RegisterImageOptionsReader(name, iopt, bytes.NewReader(b.Bytes()))
		d.ImageOptions(name, n.X, n.Y, n.W, n.H, false, iopt, 0, "")
	}
	return nil
}

var win1252Enc = charmap.Windows1252.NewEncoder()

func enc(str string) (string, error) {
	res, _, err := transform.String(win1252Enc, str)
	return res, err
}
