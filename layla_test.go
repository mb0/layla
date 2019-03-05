package layla

import (
	"strings"
	"testing"
	"time"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

var testLabel1 = `(stage :w 360 :h 360 :align 2 :gap 30 :font.size 7 :pad [30 40 30 0]
	(text :font.size 12 $title)
	(block :mar [6 0 6 0] :y 70 :h 76 "Zutaten: " $ingreds)
	(vbox :y 146 :sub.h 20
		(text 'Verpackt am: ' (time:date_long $now))
		(text 'ungeöffnet haltbar bis: ' (time:date_long (time:add_days $now $bbdays)))
		(text 'Hergestellt für: ' $vendor)
		(text 'Straße Nr, PLZ Ort')
	)
	(ellipse :y 246 :w 100 :h 66 :line 2
		(vbox :y 9 :sub.h 18 :font.size 6
			(text "DE")
			(text $stamp)
			(text "EG")
		)
	)
)`

var testLabel2 = `(stage :w 464 :h 480 :gap 32 :font.size 8 :pad [32 40 32 40]
(table :w 300 :sub.h 40 :cols [100,200]
	(text 'Produkt:')  (text $title)
	(text 'Anbieter:') (text $vendor)
	(text 'Quelle:')   (text $batch)
	(text 'Datum:')    (text (time:date_long $now))
)
(qrcode :x 268 :y 120 :code.name 'h' :code.wide 4 (str:upper ('A' 'http://vendor.url/' $batch)))
(barcode :x 8 :y 320 :h 124 :code.name 'ean128' :code.human 2 :code.wide 2 '10' $batch)
)`

func TestLayla(t *testing.T) {
	prog := &exp.ParamScope{exp.NewScope(Env), lit.ObjFromKeyed([]lit.Keyed{
		{"now", lit.Time(time.Now())},
		{"title", lit.Str("Produkt")},
		{"vendor", lit.Str("Firma GmbH")},
		{"batch", lit.Str("AB19020501")},
		{"ingreds", lit.Str("Zutaten: Zucker, Essig, Salz, Gewürze")},
		{"bbdays", lit.Int(99)},
		{"stamp", lit.Str("XY 12345")},
	})}
	tests := []struct {
		raw  string
		want string
	}{
		{testLabel1, ""},
		{testLabel2, ""},
		{`(stage :w 360 :h 360 (rect))`, `{kind:'rect' w:360 h:360}`},
		{`(stage :w 360 :h 360 (rect :h 100))`, `{kind:'rect' w:360 h:100}`},
		{`(stage :w 360 :h 360 :pad [5 5 5 5] (rect :h 100))`,
			`{kind:'rect' x:5 y:5 w:350 h:100}`},
		{`(stage :w 360 :h 360 :pad [5 5 5 5] (rect :h 100 :mar [3 3 3 3]))`,
			`{kind:'rect' x:8 y:8 w:344 h:100}`},
		{`(vbox :w 360 :h 360 :sub.h 36 (rect)(rect :h 72)(rect))`, "" +
			`{kind:'rect' w:360 h:36}` +
			`{kind:'rect' y:36 w:360 h:72}` +
			`{kind:'rect' y:108 w:360 h:36}`},
		{`(vbox :w 300 (table :sub.h 40 :cols [100,200]` +
			`(text 'a:') (text '1')` +
			`(text 'b:') (text '2'))` +
			`(text :h 30 'end'))`, "" +
			`{kind:'text' w:100 h:40 data:'a:'}` +
			`{kind:'text' x:100 w:200 h:40 data:'1'}` +
			`{kind:'text' y:40 w:100 h:40 data:'b:'}` +
			`{kind:'text' x:100 y:40 w:200 h:40 data:'2'}` +
			`{kind:'text' y:80 w:300 h:30 data:'end'}`},
	}
	for _, test := range tests {
		n, err := ExecuteString(prog, test.raw)
		if err != nil {
			t.Errorf("exec %s error: %+v", test.raw, err)
			continue
		}
		draw, err := Layout(n)
		if err != nil {
			t.Errorf("layout err: %v\n%v", err, n)
			continue
		}
		var b strings.Builder
		for _, d := range draw {
			dl, err := lit.Adapt(d)
			if err != nil {
				t.Errorf("could not adapt %v, error: %v", d, err)
			}
			b.WriteString(dl.String())
		}
		if got := b.String(); test.want != "" && got != test.want {
			t.Errorf("for %s want %s got %s", test.raw, test.want, got)
		}
	}
}
