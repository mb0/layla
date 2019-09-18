// Package html implements a layla renderer for html previews.
// Both barcodes and qrcodes are generated as images and embedded as data urls.
package html

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/gif"
	"log"

	"github.com/boombuler/barcode"
	"github.com/mb0/layla"
	"github.com/mb0/layla/bcode"
	"github.com/mb0/layla/font"
	"github.com/mb0/xelf/bfr"
)

// RenderBfr renders the node n as HTML to b or returns an error.
func RenderBfr(b bfr.B, man *font.Manager, n *layla.Node) error {
	draw, err := layla.Layout(man, n)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, `<div class="layla" style="position:relative;background-color:white;width:%fmm;height:%fmm">`, n.W/8, n.H/8)
	b.WriteString(`<style>.layla div { position: absolute; box-sizing: border-box; font-size: 8pt }</style>`)
	for _, d := range draw {
		b.WriteString(`<div style="`)
		fmt.Fprintf(b, "left:%fmm;", d.X/8)
		fmt.Fprintf(b, "top:%fmm;", d.Y/8)
		fmt.Fprintf(b, "width:%fmm;", d.W/8)
		fmt.Fprintf(b, "height:%fmm;", d.H/8)
		switch d.Kind {
		case "ellipse":
			fmt.Fprintf(b, "border:%fmm solid black;", d.Stroke/8)
			x, y := d.W/16+float64(d.Stroke), d.H/16+float64(d.Stroke/8)
			fmt.Fprintf(b, "border-radius:%fmm / %fmm;", x, y)
			b.WriteString(`">`)
		case "rect":
			fmt.Fprintf(b, "border:%fmm solid black;", d.Stroke/8)
			b.WriteString(`">`)
		case "text", "block", "styled":
			b.WriteString(`">`)
			b.WriteString(d.Data)
		case "barcode", "qrcode":
			b.WriteString(`">`)
			err = writeBarcode(b, d)
			if err != nil {
				return err
			}
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	return nil
}

func writeBarcode(b bfr.B, d *layla.Node) error {
	img, err := bcode.Barcode(d)
	if err != nil {
		return err
	}
	img, err = barcode.Scale(img, int(d.W), int(d.H))
	if err != nil {
		log.Printf("scale barcode %f %f", d.W, d.H)
		return err
	}
	fmt.Fprintf(b, `<img style="width:%fmm; height:%fmm" src="`, d.W/8, d.H/8)
	err = writeDataURL(b, img)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, `" alt="%s">`, d.Kind)
	return nil
}

// writeDataURL writes the given img as monochrome gif data url to b
func writeDataURL(b bfr.B, img image.Image) error {
	b.WriteString("data:image/gif;base64,")
	enc := base64.NewEncoder(base64.RawStdEncoding, b)
	err := gif.Encode(enc, img, nil)
	if err != nil {
		return err
	}
	return nil
}
