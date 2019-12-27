package font

import (
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Pt = fixed.Int26_6

var PtI = fixed.I

func PtF(f float64) Pt {
	return Pt(f * 64)
}

func PtToF(pt Pt) float64 {
	res := float64(pt >> 6)
	res += float64(pt&63) / 64
	return res
}

type Face struct {
	font.Face
	Add Pt
}

func (f *Face) Extra() Pt { return f.Add }

func (f *Face) Text(text string, last rune) (res Pt, _ rune) {
	for _, r := range text {
		a := f.Rune(r, last)
		res += a
		last = r
	}
	return res, last
}

func (f *Face) Rune(r, last rune) (res Pt) {
	if last != -1 && last != '\n' {
		res += f.Kern(last, r)
	}
	a, ok := f.GlyphAdvance(r)
	if !ok {
		a, _ = f.GlyphAdvance('x')
	}
	res += a
	return res
}
