package tspl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
)

func TestRenderBfr(t *testing.T) {
	man := font.NewManager(200, 1, 1).RegisterTTF("", "../testdata/font/Go-Regular.ttf")
	if err := man.Err(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		raw  string
		want string
		rot  string
	}{
		{raw: "(rect x:100 y:80 w:60 h:40 border.w:1)",
			want: "BOX 100,80,160,120,1\n",
			rot:  "BOX 280,100,320,160,1\n",
		},
		{raw: "(ellipse x:100 y:80 w:60 h:40 border.w:2)",
			want: "ELLIPSE 100,80,60,40,2\n",
			rot:  "ELLIPSE 280,100,40,60,2\n",
		},
	}
	for _, test := range tests {
		node, err := layla.Execute(layla.Env, strings.NewReader(test.raw))
		if err != nil {
			t.Errorf("execute %v", err)
			continue
		}
		var b bytes.Buffer
		err = renderNode(&b, node, 0, 400)
		if err != nil {
			t.Errorf("render %v", err)
			continue
		}
		if got := b.String(); got != test.want {
			t.Errorf("want: %s\ngot:  %s", test.want, got)
		}
		if test.rot != "" {
			b.Reset()
			err = renderNode(&b, node, 90, 400)
			if err != nil {
				t.Errorf("render %v", err)
				continue
			}
			if got := b.String(); got != test.rot {
				t.Errorf("want: %s\ngot:  %s", test.rot, got)
			}
		}
	}
}
