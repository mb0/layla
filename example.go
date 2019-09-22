// +build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"time"

	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
	"github.com/mb0/layla/html"
	"github.com/mb0/layla/pdf"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

var raw = `
(stage :w 464 :h 480 :gap 32 :font.size 8 :pad [32 32 0 0]
	(vbox :w 300 :sub.h 72
		(box (text 'Produkt:')
			(text :y 24 :font.size 10 $title))
		(box (text 'Anbieter:')
			(text :y 24 :font.size 10 $vendor))
		(box (text 'Batch:')
			(text :y 24 :font.size 10 $batch))
		(box (text 'Datum:')
			(text :y 24 :font.size 10 (time:date_long $now)))
	)
	(qrcode :x 300 :y 166 :code ['H' 0 4]
		'https://vendor.url/' $batch)
	(barcode :x 9 :y 320 :h 124.4 :code ['ean128' 2 1]
		'10' $batch)
)`

func main() {
	m := &font.Manager{}
	err := m.RegisterTTF("", "testdata/Go-Regular.ttf")
	if err != nil {
		log.Fatalf("parse font %v", err)
	}
	prog := &exp.ParamEnv{exp.NewScope(layla.Env), lit.RecFromKeyed([]lit.Keyed{
		{"now", lit.Time(time.Now())},
		{"title", lit.Str("Produkt")},
		{"vendor", lit.Str("Firma GmbH")},
		{"batch", lit.Str("AB19020501")},
	})}
	n, err := layla.ExecuteString(prog, raw)
	if err != nil {
		log.Fatalf("exec error: %v", err)
	}
	{ // write html
		var b bytes.Buffer
		b.WriteString(`<body style="background-color: grey">`)
		err = html.RenderBfr(&b, m, n)
		if err != nil {
			log.Fatalf("render html error: %v", err)
		}
		b.WriteString(`</body>`)
		err = ioutil.WriteFile("example.html", b.Bytes(), 0644)
		if err != nil {
			log.Fatalf("write html error: %v", err)
		}
	}
	{ // write pdf
		doc, err := pdf.Render(m, n)
		if err != nil {
			log.Fatalf("render error: %v", err)
		}
		err = doc.OutputFileAndClose("example.pdf")
		if err != nil {
			log.Fatalf("write error: %v", err)
		}
	}
}
