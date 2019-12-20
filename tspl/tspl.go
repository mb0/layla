// Package tspl implements a layla renderer for TSC thermal label printer using TSPL.
// This package specifically targets the TSC DA-200 printer, which supports both bar and qr-codes.
package tspl

import (
	"fmt"
	"strings"

	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
	"github.com/mb0/xelf/bfr"
)

func dot(f float64) int {
	return int(f)
}

// RenderBfr renders the node n as TSPL to b or returns an error.
func RenderBfr(b bfr.B, man *font.Manager, n *layla.Node, extra ...string) error {
	draw, err := layla.Layout(man, n)
	if err != nil {
		return err
	}
	w, h := n.W, n.H
	if n.Rot == 90 {
		w, h = h, w
	}
	fmt.Fprintf(b, "SIZE %g mm, %g mm\n", w/8, h/8)
	fmt.Fprintf(b, "GAP %g mm, 0 mm\n", n.Gap/8)
	b.WriteString("DIRECTION 1,0\nCODEPAGE UTF-8\n")
	for _, line := range extra {
		b.WriteString(line)
		if len(line) > 0 && line[len(line)-1] != '\n' {
			b.WriteByte('\n')
		}
	}
	b.WriteString("CLS\n")
	for _, d := range draw {
		err = renderNode(b, d, n.Rot, n.H)
		if err != nil {
			return err
		}
	}
	return nil
}

func renderNode(b bfr.B, d *layla.Node, rot int, rh float64) error {
	if rot != 0 {
		switch d.Kind {
		case "rect", "line", "ellipse":
			d.X, d.Y = rh-d.Y-d.H, d.X
			d.W, d.H = d.H, d.W
		case "text", "barcode", "qrcode":
			d.X, d.Y = rh-d.Y, d.X
		}
	}
	switch d.Kind {
	case "ellipse":
		fmt.Fprintf(b, "ELLIPSE %d,%d,%d,%d,%d\n",
			dot(d.X), dot(d.Y), dot(d.W), dot(d.H),
			dot(d.Border.W))
	case "rect":
		fmt.Fprintf(b, "BOX %d,%d,%d,%d,%d\n",
			dot(d.X), dot(d.Y), dot(d.X+d.W), dot(d.Y+d.H),
			dot(d.Border.W))
	case "line":
		fmt.Fprintf(b, "LINE %d,%d,%d,%d,%d\n",
			dot(d.X), dot(d.Y), dot(d.X+d.W), dot(d.Y+d.H),
			dot(d.Border.W))
	case "text":
		fsize := fontSize(d)
		data := strings.Replace(fmt.Sprintf("%q", d.Data), "\\n", "\\[L]", -1)
		space := d.Font.Line - (d.Font.Height * 25.4 * 8 / 72)
		x, w := dot(d.X), dot(d.W)
		// TODO fix overflow due to discrepancy between font measuring and printing
		// the reason might be that the tsc printer does not apply kerning?
		switch d.Align {
		case 3:
			x -= 10
		case 2:
			x -= 5
			w += 5
		default:
			w += 10
		}
		fmt.Fprintf(b, "BLOCK %d,%d,%d,%d,\"0\",%d,%d,%d,%d,%d,%s\n",
			dot(d.X-2), dot(d.Y), dot(d.W+4), dot(d.H), rot,
			fsize, fsize, dot(space), d.Align, data)
	case "barcode":
		fmt.Fprintf(b, "BARCODE %d,%d,%q,%d,%d,%d,%d,%d,%q\n",
			dot(d.X), dot(d.Y), strings.ToUpper(d.Code.Name), dot(d.H),
			dot(d.Code.Wide), rot, d.Code.Human, d.Align, d.Data)
	case "qrcode":
		fmt.Fprintf(b, "QRCODE %d,%d,%s,%d,A,%d,M2,S7,%q\n",
			dot(d.X), dot(d.Y), strings.ToUpper(d.Code.Name),
			dot(d.Code.Wide), rot, d.Data)
	default:
		return fmt.Errorf("layout %s not supported", d.Kind)
	}
	return nil
}

func fontSize(n *layla.Node) (res int) {
	if n.Font != nil {
		res = dot(n.Font.Size)
	}
	if res == 0 {
		res = 8
	}
	return res
}
