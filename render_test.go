package layla_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
	"github.com/mb0/layla/html"
	"github.com/mb0/layla/pdf"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

func man() *font.Manager {
	m := font.NewManager(72, 2, 4).
		RegisterTTF("regular", "testdata/font/Go-Regular.ttf").
		RegisterTTF("bold", "testdata/font/Go-Bold.ttf")
	if err := m.Err(); err != nil {
		log.Fatal(err)
	}
	return m
}

var testFiles = []string{
	"lines",
	"textbox",
	"pages",
	"label1",
	"label2",
}

func TestHtml(t *testing.T) {
	m := man()
	for _, name := range testFiles {
		n, err := read(name)
		if err != nil {
			t.Errorf("error reading test file %q: %v", name, err)
			continue
		}
		var b bytes.Buffer
		b.WriteString("<body style=\"background-color: grey\">\n")
		err = html.RenderBfr(&b, m, n)
		if err != nil {
			t.Errorf("render html error: %v", err)
			continue
		}
		b.WriteString(`</body>`)
		err = ioutil.WriteFile(path(name, ".html"), b.Bytes(), 0644)
		if err != nil {
			t.Errorf("write html error: %v", err)
		}
	}
}

func TestPdf(t *testing.T) {
	m := man()
	for _, name := range testFiles {
		n, err := read(name)
		if err != nil {
			t.Errorf("error reading test file %q: %v", name, err)
			continue
		}
		doc, err := pdf.Render(m, n)
		if err != nil {
			t.Errorf("render %q error: %v", name, err)
			continue
		}
		err = doc.OutputFileAndClose(path(name, ".pdf"))
		if err != nil {
			t.Errorf("write error: %v", err)
		}
	}
}

func read(name string) (*layla.Node, error) {
	f, err := os.Open(path(name, ".layla"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	now := time.Date(2019, time.October, 5, 23, 0, 0, 0, time.UTC)
	param := lit.RecFromKeyed([]lit.Keyed{
		{"now", lit.Time(now)},
		{"title", lit.Str("Produkt")},
		{"vendor", lit.Str("Firma GmbH")},
		{"batch", lit.Str("AB19020501")},
		{"ingreds", lit.Str("list of all the ingredients, like suger and spice and everthing nice.")},
	})
	env := &exp.ParamEnv{layla.Env, param}
	return layla.Execute(env, f)
}

func path(name, ext string) string {
	return filepath.Join("testdata", name+ext)
}
