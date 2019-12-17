package layla

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mb0/layla/font"
	"github.com/mb0/xelf/prx"
)

func TestLayla(t *testing.T) {
	man := &font.Manager{}
	err := man.RegisterTTF("", "testdata/font/Go-Regular.ttf")
	if err != nil {
		t.Fatalf("register font error: %v", err)
	}
	tests := []struct {
		raw  string
		want string
	}{
		{`(stage w:360 h:360 (rect))`, `{kind:'rect' w:360 h:360}`},
		{`(rect w:360 h:360 (text 'Hello'))`, `{kind:'rect' w:360 h:360}` +
			`{kind:'text' w:82 h:41 font:{line:41} data:'Hello'}`},
		{`(stage w:360 h:360 (rect h:100))`, `{kind:'rect' w:360 h:100}`},
		{`(stage w:360 h:360 pad:[5 5 5 5] (rect h:100))`,
			`{kind:'rect' x:5 y:5 w:350 h:100}`},
		{`(stage w:360 h:360 pad:[5 5 5 5] (rect h:100 mar:[3 3 3 3]))`,
			`{kind:'rect' x:8 y:8 w:344 h:100}`},
		{`(vbox w:360 h:360 sub.h:36 (rect)(rect h:72)(rect))`, "" +
			`{kind:'rect' w:360 h:36}` +
			`{kind:'rect' y:36 w:360 h:72}` +
			`{kind:'rect' y:108 w:360 h:36}`},
		{`(vbox w:360 h:360 sub.h:36 (rect)(if false (rect h:72))(rect))`, "" +
			`{kind:'rect' w:360 h:36}` +
			`{kind:'rect' y:36 w:360 h:36}`},
		{`(vbox w:300 h:200 (rect w:200 h:100))`, `{kind:'rect' w:200 h:100}`},
		{`(vbox w:300 h:200 align:1 (rect w:200 h:100))`,
			`{kind:'rect' x:100 w:200 h:100}`},
		{`(vbox w:300 h:200 align:2 (rect w:200 h:100))`,
			`{kind:'rect' x:50 w:200 h:100}`},
		{`(hbox w:300 h:200 (rect w:200 h:100))`, `{kind:'rect' w:200 h:100}`},
		{`(vbox w:300 (table sub.h:41 cols:[100,200]` +
			`(text 'a:') (text '1')` +
			`(text 'b:') (text '2'))` +
			`(text h:30 'end'))`, "" +
			`{kind:'text' w:100 h:41 font:{line:41} data:'a:'}` +
			`{kind:'text' x:100 w:200 h:41 font:{line:41} data:'1'}` +
			`{kind:'text' y:41 w:100 h:41 font:{line:41} data:'b:'}` +
			`{kind:'text' x:100 y:41 w:200 h:41 font:{line:41} data:'2'}` +
			`{kind:'text' y:82 w:300 h:30 font:{line:41} data:'end'}`},
		{`(vbox w:300 h:300 list:(list (text 'Hello') (text 'World')))`, "" +
			`{kind:'text' w:300 h:41 font:{line:41} data:'Hello'}` +
			`{kind:'text' y:41 w:300 h:41 font:{line:41} data:'World'}`},
		{`(page w:200 h:41 (vbox (text 'Page1') (text 'Page2') (text 'Page3')))`, "" +
			`{kind:'text' w:200 h:41 font:{line:41} data:'Page1'}` +
			`{kind:'page'}{kind:'text' w:200 h:41 font:{line:41} data:'Page2'}` +
			`{kind:'page'}{kind:'text' w:200 h:41 font:{line:41} data:'Page3'}`},
		{`(page w:200 h:41 (text 'Page1\nPage2\nPage3'))`, "" +
			`{kind:'text' w:99 h:41 font:{line:41} data:'Page1'}` +
			`{kind:'page'}{kind:'text' w:99 h:41 font:{line:41} data:'Page2'}` +
			`{kind:'page'}{kind:'text' w:99 h:41 font:{line:41} data:'Page3'}`},
		{`(page w:200 h:100 (text 'Page1\nPage2\nPage3'))`, "" +
			`{kind:'text' w:99 h:82 font:{line:41} data:'Page1\nPage2'}` +
			`{kind:'page'}{kind:'text' w:99 h:41 font:{line:41} data:'Page3'}`},
		{`(page w:200 h:41 (text 'Hello World\nHallo Welt'))`, "" +
			`{kind:'text' w:181 h:41 font:{line:41} data:'Hello World'}` +
			`{kind:'page'}{kind:'text' w:181 h:41 font:{line:41} data:'Hallo Welt'}`},
	}
	for _, test := range tests {
		n, err := ExecuteString(Env, test.raw)
		if err != nil {
			t.Errorf("exec %s error: %+v", test.raw, err)
			continue
		}
		draw, err := Layout(man, n)
		if err != nil {
			t.Errorf("layout err: %v\n%v", err, n)
			continue
		}
		var b strings.Builder
		for _, d := range draw {
			dl, err := prx.Adapt(d)
			if err != nil {
				t.Errorf("could not adapt %v, error: %v", d, err)
			}
			b.WriteString(dl.String())
		}
		if got := b.String(); test.want != "" && got != test.want {
			t.Errorf("for %s\nwant: %s\n got: %s", test.raw, test.want, got)
		}
	}
}

func TestMeasure(t *testing.T) {
	man := &font.Manager{}
	err := man.RegisterTTF("", "testdata/font/Go-Regular.ttf")
	if err != nil {
		t.Fatalf("register font error: %v", err)
	}
	tests := []struct {
		raw  string
		want string
		calc string
	}{
		{"(stage w:10 h:20)", `{"w":10,"h":20}`, ``},
		{"(rect w:10 h:20)", `{"w":10,"h":20}`, ``},
		{"(text 'Hello')", `{"w":82,"h":41}`, ``},
		{"(text 'World')", `{"w":91,"h":41}`, ``},
		{"(text 'Hello World')", `{"w":91,"h":82}`, ``},
		{"(text mar:[1 2 3 4] 'Hello')", `{"w":86,"h":47}`, `{"x":1,"y":2,"w":82,"h":41}`},
	}
	for _, test := range tests {
		n, err := ExecuteString(Env, test.raw)
		if err != nil {
			t.Errorf("exec %s error: %+v", test.raw, err)
			continue
		}
		lay := newLayouter(man, n)
		b, err := lay.layout(n, Box{Dim: Dim{120, 0}}, nil)
		if err != nil {
			t.Errorf("measure %s error: %+v", test.raw, err)
			continue
		}
		want := test.want
		bb, _ := json.Marshal(b)
		if got := string(bb); got != want {
			t.Errorf("for %s\nwant res: %s\n got: %s", test.raw, want, got)
		}
		if test.calc != "" {
			want = test.calc
		}
		bb, _ = json.Marshal(n.Calc)
		if got := string(bb); got != want {
			t.Errorf("for %s\nwant calc: %s\n got: %s", test.raw, want, got)
		}
	}
}
