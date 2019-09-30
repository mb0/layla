package pdf

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/mb0/layla"
	"github.com/mb0/layla/font"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
)

var man = &font.Manager{}

func init() {
	fonts := []struct {
		name, path string
	}{
		{"regular", "../testdata/Go-Regular.ttf"},
		{"bold", "../testdata/Go-Bold.ttf"},
	}
	for _, f := range fonts {
		err := man.RegisterTTF(f.name, f.path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestPdf(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"letter1.layla"},
		{"measure1.layla"},
	}
	for _, test := range tests {
		runTest(t, test.path)
	}
}

func runTest(t *testing.T, path string) {
	f, err := os.Open(filepath.Join("../testdata", path))
	if err != nil {
		t.Errorf("error opening file %q: %v", path, err)
		return
	}
	defer f.Close()
	param := lit.RecFromKeyed([]lit.Keyed{})
	env := &exp.ParamEnv{exp.NewScope(layla.Env), param}
	n, err := layla.Execute(env, f)
	if err != nil {
		t.Fatalf("exec %q error: %v", path, err)
	}
	doc, err := Render(man, n)
	if err != nil {
		t.Fatalf("render %q error: %v", path, err)
	}
	ext := filepath.Ext(path)
	err = doc.OutputFileAndClose(filepath.Join("../testdata", path[:len(path)-len(ext)]+".pdf"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
}
