package font

import (
	"io/ioutil"

	"github.com/golang/freetype/truetype"
	"github.com/mb0/xelf/cor"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Key struct {
	Name string
	Size float64
}

type Src struct {
	*truetype.Font
	Path string
}

type Manager struct {
	ttfs  map[string]*Src
	faces map[Key]font.Face
}

func (m *Manager) RegisterTTF(name string, path string) error {
	_, ok := m.ttfs[name]
	if ok {
		return nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return cor.Errorf("reading file %q: %v", path, err)
	}
	f, err := truetype.Parse(data)
	if err != nil {
		return cor.Errorf("parse file %q: %v", path, err)
	}
	if m.ttfs == nil {
		m.ttfs = make(map[string]*Src)
	}
	m.ttfs[name] = &Src{f, path}
	return nil
}

func (m *Manager) Close() error {
	for _, f := range m.faces {
		f.Close()
	}
	return nil
}

func (m *Manager) Path(name string) (string, error) {
	src, ok := m.ttfs[name]
	if !ok {
		return "", cor.Errorf("unknown font %q", name)
	}
	return src.Path, nil
}

func (m *Manager) Face(name string, size float64) (font.Face, error) {
	key := Key{name, size}
	f, ok := m.faces[key]
	if ok {
		return f, nil
	}
	src, ok := m.ttfs[name]
	if !ok {
		return nil, cor.Errorf("unknown font %q", name)
	}
	f = truetype.NewFace(src.Font, &truetype.Options{Size: size})
	if m.faces == nil {
		m.faces = make(map[Key]font.Face)
	}
	m.faces[key] = f
	return f, nil
}

func Layout(f font.Face, text string, width int) ([]string, int, int, error) {
	mw := fixed.I(width)
	var li, le int
	var sw, lw, mlw fixed.Int26_6
	last := rune('\n')
	var res []string
	for _, t := range tokens(text) {
		if t.len == 0 { // hard break
			if li >= 0 {
				if lw > mlw {
					mlw = lw
				}
				res = append(res, text[li:le])
			} else {
				res = append(res, "")
			}
			li = -1
			sw, lw = 0, 0
			last = '\n'
			continue
		}
		if li < 0 {
			li = t.off
			le = t.off
		}
		tx := text[t.off : t.off+t.len]
		tw, lst := tokWidth(f, tx, last)
		if tx[0] == ' ' { // spaces
			sw = tw
			last = ' '
			continue
		} else if lw+sw+tw > mw {
			if tw > mw { // hard break token
				for i, c := range tx {
					a, ok := f.GlyphAdvance(c)
					if !ok {
						a, _ = f.GlyphAdvance('x')
					}
					ww := a
					if last != -1 && last != '\n' {
						ww += f.Kern(last, c)
					}
					if lw+ww > mw {
						if lw > mlw {
							mlw = lw
						}
						res = append(res, text[li:t.off+i])
						li = t.off + i
						sw, lw = 0, a
					} else {
						sw, lw = 0, lw+sw+a
					}
					last = c
				}
			} else { // soft break
				if lw > 0 {
					if lw > mlw {
						mlw = lw
					}
					res = append(res, text[li:le])
				}
				li, le = t.off, 0
				sw, lw = 0, tw
			}
		} else {
			sw, lw = 0, lw+sw+tw
		}
		le = t.off + t.len
		last = lst
	}
	if lw > 0 {
		if lw > mlw {
			mlw = lw
		}
		res = append(res, text[li:le])
	}
	lh := f.Metrics().Height.Ceil()
	return res, lh, mlw.Ceil(), nil
}

func tokWidth(f font.Face, text string, last rune) (res fixed.Int26_6, _ rune) {
	for _, c := range text {
		a, ok := f.GlyphAdvance(c)
		if !ok {
			a, _ = f.GlyphAdvance('x')
		}
		if last != -1 && last != '\n' {
			res += f.Kern(last, c)
		}
		res += a
		last = c
	}
	return res, last
}

type tok struct {
	off, len int
}

func tokens(text string) (res []tok) {
	var start int
	var space, dash bool
	for i, c := range text {
		if c == '\n' {
			if !space && i > start {
				res = append(res, tok{start, i - start})
			}
			res = append(res, tok{i, 0})
			space, dash = false, false
			start = i + 1
		} else if c == ' ' {
			if !space {
				if i > start {
					res = append(res, tok{start, i - start})
				}
				space, dash = true, false
				start = i
			}
		} else if i > start {
			if space {
				res = append(res, tok{start, i - start})
				space, dash = false, false
				start = i
			} else if dash {
				res = append(res, tok{start, i - start})
				space, dash = false, false
				start = i
			} else if c == '-' {
				dash = true
			}
		}
	}
	if len(text) > start {
		res = append(res, tok{start, len(text) - start})
	}
	return res
}
