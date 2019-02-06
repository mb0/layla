// +build ignore

package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/mb0/layla"
	"github.com/mb0/layla/html"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/utl"
)

var raw = `
(stage :w 58 :h 60 :gap 4 :font.size 1 :pad [4 4 00]
	(vbox :w 37.5 :sub.h 9
		(group (text 'Produkt:')
			(text :y 3 :font.size 2 $title))
		(group (text 'Anbieter:')
			(text :y 3 :font.size 2 $vendor))
		(group (text 'Batch:')
			(text :y 3 :font.size 2 $batch))
		(group (text 'Datum:')
			(text :y 3 :font.size 2 (time.date_long $now)))
	)
	(qrcode :x 37.5 :y 20.75 :code ['H' 0 0.5]
		'https://vendor.url/' $batch)
	(barcode :x 1.125 :y 40 :h 15.5 :code ['ean128' 2 0.25]
		'10' $batch)
)`

func main() {
	prog := utl.Prog(&lit.Dict{[]lit.Keyed{
		{"now", lit.Time(time.Now())},
		{"title", lit.Str("Produkt")},
		{"vendor", lit.Str("Firma GmbH")},
		{"batch", lit.Str("AB19020501")},
	}}, layla.Env...)
	n, err := layla.ExecuteString(prog, raw)
	if err != nil {
		log.Fatalf("exec error: %v", err)
	}
	b := bufio.NewWriter(os.Stdout)
	b.WriteString(`<body style="background-color: grey">`)
	err = html.RenderBfr(b, n)
	if err != nil {
		log.Fatalf("render error: %v", err)
	}
	b.WriteString(`</body>`)
	err = b.Flush()
	if err != nil {
		log.Fatalf("flush error: %v", err)
	}
}
