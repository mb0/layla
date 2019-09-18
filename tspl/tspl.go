// Package tspl implements a layla renderer for TSC thermal label printer using TSPL.
// This package specifically targets the TSC DA-200 printer, which supports both bar and qr-codes.
package tspl

import (
	"fmt"
	"strings"

	"github.com/mb0/dql/bfr"
	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
)

func dot(f float64) int {
	return int(f * 8)
}

// RenderBfr renders the node n as TSPL to b or returns an error.
func RenderBfr(b bfr.B, man *font.Manager, n *layla.Node, extra ...string) error {
	draw, err := layla.Layout(man, n)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, "SIZE %f mm, %f mm\n", n.W, n.H)
	fmt.Fprintf(b, "GAP %f mm, 0 mm\n", n.Gap)
	b.WriteString("DIRECTION 1,0\nCODEPAGE UTF-8\n")
	for _, line := range extra {
		b.WriteString(line)
		if len(line) > 0 && line[len(line)-1] != '\n' {
			b.WriteByte('\n')
		}
	}
	b.WriteString("CLS\n")
	for _, d := range draw {
		switch d.Kind {
		case "ellipse":
			fmt.Fprintf(b, "ELLIPSE %d,%d,%d,%d,%d\n",
				dot(d.X), dot(d.Y), dot(d.W), dot(d.H),
				dot(d.Stroke))
		case "rect":
			fmt.Fprintf(b, "BOX %d,%d,%d,%d,%d\n",
				dot(d.X), dot(d.Y), dot(d.X+d.W), dot(d.Y+d.H),
				dot(d.Stroke))
		case "text":
			fsize := fontSize(d)
			fmt.Fprintf(b, "TEXT %d,%d,\"0\",0,%d,%d,%d,%q\n",
				dot(d.X), dot(d.Y),
				fsize, fsize, d.Align, d.Data)
		case "block":
			fsize := fontSize(d)
			fmt.Fprintf(b, "BLOCK %d,%d,%d,%d,\"0\",0,%d,%d,%d,%d,%q\n",
				dot(d.X), dot(d.Y), dot(d.W), dot(d.H),
				fsize, fsize, dot(d.Sub.H), d.Align, d.Data)
		case "barcode":
			fmt.Fprintf(b, "BARCODE %d,%d,%s,%d,%d,0,%d,%d,%q\n",
				dot(d.X), dot(d.Y), strings.ToUpper(d.Code.Name), dot(d.H),
				d.Code.Human, dot(d.Code.Wide), d.Align, d.Data)
		case "qrcode":
			fmt.Fprintf(b, "BARCODE %d,%d,%s,%d,A,0,M2,S7,%q\n",
				dot(d.X), dot(d.Y), strings.ToUpper(d.Code.Name),
				dot(d.Code.Wide), d.Data)
		}
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
