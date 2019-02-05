// Package html implements a layla renderer for html previews.
// Both barcodes and qrcodes are generated as images and embedded as data urls.
package html

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/gif"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/qr"
	"github.com/mb0/layla"
	"github.com/mb0/xelf/bfr"
)

// RenderBfr renders the node n as HTML to b or returns an error.
func RenderBfr(b bfr.B, n *layla.Node) error {
	draw, err := layla.Layout(n)
	if err != nil {
		return err
	}
	b.WriteString(`<div class="layla">`)
	b.WriteString(`<style>
.layla { position: relative }
.layla div { position: absolute; box-sizing: border-box }
</style>`)
	for _, d := range draw {
		b.WriteString(`<div style="`)
		fmt.Fprintf(b, "left:%dpx;", d.X)
		fmt.Fprintf(b, "top:%dpx;", d.Y)
		fmt.Fprintf(b, "width:%dpx;", d.W)
		fmt.Fprintf(b, "height:%dpx;", d.H)
		switch d.Kind {
		case "ellipse":
			fmt.Fprintf(b, "border:%dpx solid black;", d.Line)
			x, y := d.W/2+d.Line, d.H/2+d.Line
			fmt.Fprintf(b, "border-radius:%dpx / %dpx;", x, y)
			b.WriteString(`">`)
		case "rect":
			fmt.Fprintf(b, "border:%dpx solid black;", d.Line)
			b.WriteString(`">`)
		case "text", "block", "styled":
			b.WriteString(`">`)
			b.WriteString(d.Data)
		case "barcode", "qrcode":
			b.WriteString(`">`)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	return nil
}

func writeBarcode(b bfr.B, d *layla.Node) error {
	img, err := getBarcode(d)
	if err != nil {
		return err
	}
	img, err = barcode.Scale(img, d.W, d.H)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, `<img width="%d" height="%d" src="`, d.W, d.H)
	err = writeDataURL(b, img)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, `" alt="%s">`, d.Kind)
	return nil
}

func getBarcode(d *layla.Node) (barcode.Barcode, error) {
	switch d.Code.Name {
	case "ean128":
		return code128.Encode(d.Data)
	case "ean8", "ean13":
		return ean.Encode(d.Data)
	}
	if d.Kind != "qrcode" {
		return nil, fmt.Errorf("unknown code name %q", d.Code.Name)
	}
	ec := getErrorCorrection(d.Code.Name)
	return qr.Encode(d.Data, ec, qr.Auto)
}

func getErrorCorrection(name string) qr.ErrorCorrectionLevel {
	switch name {
	case "l":
		return qr.L
	case "m":
		return qr.M
	case "q":
		return qr.Q
	}
	return qr.H
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
