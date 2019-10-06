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
	fmt.Fprintf(b, "SIZE %g mm, %g mm\n", n.W/8, n.H/8)
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
			fmt.Fprintf(b, "BLOCK %d,%d,%d,%d,\"0\",0,%d,%d,%d,%d,%q\n",
				dot(d.X), dot(d.Y), dot(d.W), dot(d.H),
				fsize, fsize, dot(d.Font.Line), d.Align, d.Data)
		case "barcode":
			fmt.Fprintf(b, "BARCODE %d,%d,%s,%d,%d,0,%d,%d,%q\n",
				dot(d.X), dot(d.Y), strings.ToUpper(d.Code.Name), dot(d.H),
				d.Code.Human, dot(d.Code.Wide), d.Align, d.Data)
		case "qrcode":
			fmt.Fprintf(b, "BARCODE %d,%d,%s,%d,A,0,M2,S7,%q\n",
				dot(d.X), dot(d.Y), strings.ToUpper(d.Code.Name),
				dot(d.Code.Wide), d.Data)
		case "page":
			return fmt.Errorf("paging not supported")
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
