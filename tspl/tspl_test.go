package tspl

import (
	"strings"
	"testing"

	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
)

func TestRenderNode(t *testing.T) {
	man := font.NewManager(200, 1, 1).RegisterTTF("", "../testdata/font/Go-Regular.ttf")
	if err := man.Err(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		raw  string
		want string
		rot  string
	}{
		{raw: "(box w:400 h:400 (rect x:100 y:80 w:60 h:40 border.w:1))",
			want: "BOX 100,80,160,120,1\n",
			rot:  "BOX 280,100,320,160,1\n",
		},
		{raw: "(box w:400 h:400 (ellipse x:100 y:80 w:60 h:40 border.w:2))",
			want: "ELLIPSE 100,80,60,40,2\n",
			rot:  "ELLIPSE 280,100,40,60,2\n",
		},
		{raw: "(stage w:5000 h:800 (text align:2 font.size:32 'Smokey Mayonnaise'))",
			want: "BLOCK -5,0,839,108,\"0\",0,32,32,18,2,\"Smokey Mayonnaise\"\n",
		},
		{raw: "(box w:400 h:400 (markup `Test *Test* Test`))", want: "" +
			"BLOCK 0,0,73,41,\"0\",0,8,8,7,0,\"Test\"\n" +
			"BLOCK 71,0,74,41,\"0\",0,8,8,7,0,\"Test\"\n" +
			"BLOCK 72,0,75,41,\"0\",0,8,8,7,0,\"Test\"\n" +
			"BLOCK 143,0,73,41,\"0\",0,8,8,7,0,\"Test\"\n", rot: "" +
			"BLOCK 400,0,73,41,\"0\",90,8,8,7,0,\"Test\"\n" +
			"BLOCK 400,71,74,41,\"0\",90,8,8,7,0,\"Test\"\n" +
			"BLOCK 401,71,75,41,\"0\",90,8,8,7,0,\"Test\"\n" +
			"BLOCK 400,143,73,41,\"0\",90,8,8,7,0,\"Test\"\n",
		},
	}
	for _, test := range tests {
		got, err := render(man, test.raw, false, 400)
		if err != nil {
			t.Errorf("rot %v", err)
			continue
		}
		if got != test.want {
			t.Errorf("want: %s\ngot:  %s", test.want, got)
		}
		if test.rot != "" {
			got, err = render(man, test.raw, true, 400)
			if err != nil {
				t.Errorf("render %v", err)
				continue
			}
			if got != test.rot {
				t.Errorf("want: %s\ngot:  %s", test.rot, got)
			}
		}
	}
}

func render(man *font.Manager, raw string, rot bool, h float64) (string, error) {
	node, err := layla.Execute(layla.Env, strings.NewReader(raw))
	if err != nil {
		return "", err
	}
	lay := &layla.Layouter{man, 'i', layla.FakeBoldStyler}
	draw, err := lay.LayoutAndPage(node)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, d := range draw {
		deg := 0
		if rot {
			deg = 90
		}
		err = renderNode(lay, &b, d, deg, h)
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}
