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
(stage :w 464 :h 480 :gap 32 :font.size 8 :pad [30 30 00]
	(vbox :w 300 :sub.h 72
		(group (text 'Produkt:')
			(text :y 24 :font.size 16 $title))
		(group (text 'Anbieter:')
			(text :y 24 :font.size 16 $vendor))
		(group (text 'Batch:')
			(text :y 24 :font.size 16 $batch))
		(group (text 'Datum:')
			(text :y 24 :font.size 16 (time.date_long $now)))
	)
	(qrcode :x 300 :y 166 :code ['H' 00 4]
		'https://vendor.url/' $batch)
	(barcode :x 10 :y 320 :h 124 :code ['ean128' 2 2 0]
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
	err = html.RenderBfr(b, n)
	if err != nil {
		log.Fatalf("render error: %v", err)
	}
	err = b.Flush()
	if err != nil {
		log.Fatalf("flush error: %v", err)
	}
}
