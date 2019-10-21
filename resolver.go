package layla

import (
	"io"
	"strings"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/prx"
	"github.com/mb0/xelf/std"
	"github.com/mb0/xelf/typ"
	"github.com/mb0/xelf/utl"
)

// Env is the default resolver environment for layla.
var Env = exp.Builtin{
	NodeLookup,
	utl.StrLib.Lookup(),
	utl.TimeLib.Lookup(),
	std.Core, std.Decl,
}

// ExecuteString parses and executes the expression string s and returns a node or error.
func ExecuteString(env exp.Env, s string) (*Node, error) {
	return Execute(env, strings.NewReader(s))
}

// Execute parses and executes the expression from reader r and returns a node or error.
func Execute(env exp.Env, rr io.Reader) (*Node, error) {
	x, err := exp.Read(rr)
	if err != nil {
		return nil, err
	}
	r, err := exp.Eval(env, x)
	if err != nil {
		return nil, err
	}
	n := getNode(r.(*exp.Atom).Lit)
	if n == nil {
		return nil, cor.Errorf("expected *layla.Node got %T", r)
	}
	return n, nil
}

// NodeLookup is the resolver lookup for layla node resolvers
func NodeLookup(sym string) *exp.Spec {
	if f := forms[sym]; f != nil {
		return f
	}
	return nil
}

var forms map[string]*exp.Spec

func init() {
	t, err := prx.Reflect((*Node)(nil))
	if err != nil {
		panic(err)
	}
	nodeSig := []typ.Param{{Name: "tags?"}, {Name: "tail?"}, {Type: t}}
	listNodes := []string{"stage", "rect", "ellipse", "box", "vbox", "hbox", "table",
		"page", "extra", "cover", "header", "footer"}
	dataNodes := []string{"line", "text", "qrcode", "barcode"}
	forms = make(map[string]*exp.Spec, len(listNodes)+len(dataNodes))
	for _, n := range listNodes {
		forms[n] = &exp.Spec{typ.Form(n, nodeSig),
			utl.NewNodeResolver(listRules, &Node{Kind: n})}
	}
	for _, n := range dataNodes {
		forms[n] = &exp.Spec{typ.Form(n, nodeSig),
			utl.NewNodeResolver(dataRules, &Node{Kind: n})}
	}
}

var dataRules = utl.NodeRules{
	Tail: utl.KeyRule{
		KeyPrepper: utl.DynPrepper,
		KeySetter: func(n utl.Node, _ string, el lit.Lit) error {
			_, err := n.SetKey("data", el)
			return err
		},
	},
}

var listRules = utl.NodeRules{
	Tail: utl.KeyRule{
		KeyPrepper: utl.ListPrepper,
		KeySetter: func(n utl.Node, _ string, list lit.Lit) error {
			o := getNode(n)
			for _, el := range list.(*lit.List).Data {
				c := getNode(el)
				if el.IsZero() {
					continue
				}
				if c == nil {
					return cor.Errorf("not a layla node %T", el)
				}
				if c.Align == 0 {
					c.Align = o.Align
				}
				if c.Font == nil {
					c.Font = o.Font
				}
				o.List = append(o.List, c)
			}
			return nil
		},
	},
}

func getNode(e lit.Lit) *Node {
	if a, ok := e.(utl.Node); ok {
		n, _ := a.Ptr().(*Node)
		return n
	}
	return nil
}
