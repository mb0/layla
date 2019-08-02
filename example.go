// +build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"time"

	"github.com/mb0/layla"
	"github.com/mb0/layla/html"
	"github.com/mb0/layla/pdf"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

var raw = `
(stage :w 58 :h 60 :gap 4 :font.size 1 :pad [4 4 0 0]
	(vbox :w 37.5 :sub.h 9
		(group (text 'Produkt:')
			(text :y 3 :font.size 2 $title))
		(group (text 'Anbieter:')
			(text :y 3 :font.size 2 $vendor))
		(group (text 'Batch:')
			(text :y 3 :font.size 2 $batch))
		(group (text 'Datum:')
			(text :y 3 :font.size 2 (time:date_long $now)))
	)
	(qrcode :x 37.5 :y 20.75 :code ['H' 0 0.5]
		'https://vendor.url/' $batch)
	(barcode :x 1.125 :y 40 :h 15.5 :code ['ean128' 2 0.25]
		'10' $batch)
)`

func main() {
	prog := &exp.ParamScope{exp.NewScope(layla.Env), lit.RecFromKeyed([]lit.Keyed{
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
		err = html.RenderBfr(&b, n)
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
		doc, err := pdf.Render(n)
		if err != nil {
			log.Fatalf("render error: %v", err)
		}
		err = doc.OutputFileAndClose("example.pdf")
		if err != nil {
			log.Fatalf("write error: %v", err)
		}
	}
}
