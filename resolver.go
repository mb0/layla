package layla

import (
	"errors"
	"fmt"
	"time"

	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/utl"
)

// Env is the default resolver environment for layla.
var Env = exp.Builtin{
	NodeLookup,
	utl.StrLib.Lookup(),
	utl.TimeLib.Lookup(),
	exp.Std, exp.Core,
}

// NodeLookup is the resolver lookup for layla node resolvers
func NodeLookup(sym string) exp.Resolver {
	if hasData(sym) || hasList(sym) {
		return exp.ExprResolverFunc(resolveNode)
	}
	return nil
}

// ExecuteString parses and executes the expression string s and returns a node or error.
func ExecuteString(env exp.Env, s string) (*Node, error) {
	x, err := exp.ParseString(s)
	if err != nil {
		return nil, err
	}
	r, err := exp.Execute(env, x)
	if err != nil {
		return nil, err
	}
	n, ok := getPtr(r).(*Node)
	if !ok {
		return nil, fmt.Errorf("expected *layla.Node got %T", r)
	}
	return n, nil
}

func fDateLong(t time.Time) string             { return t.Format("2006-01-02") }
func fAddDays(t time.Time, days int) time.Time { return t.AddDate(0, 0, days) }

func hasList(sym string) bool {
	switch sym {
	case "stage", "rect", "ellipse", "group", "vbox", "hbox", "table":
		return true
	}
	return false
}

func hasData(sym string) bool {
	switch sym {
	case "text", "block", "qrcode", "barcode":
		return true
	}
	return false
}

func resolveNode(c *exp.Ctx, env exp.Env, e *exp.Expr) (exp.El, error) {
	o := &Node{Kind: e.Name}
	var r utl.NodeRules
	if hasData(e.Name) {
		r.Tail.KeySetter = func(n utl.Node, _ string, el lit.Lit) error {
			return lit.AssignTo(el, &o.Data)
		}
	} else if hasList(e.Name) {
		r.Tail.KeyPrepper = utl.ListPrepper
		r.Tail.KeySetter = func(n utl.Node, _ string, list lit.Lit) error {
			for _, el := range list.(lit.List) {
				co, ok := getPtr(el).(*Node)
				if !ok {
					return fmt.Errorf("not a layla node %T", el)
				}
				inheritAttr(o, co)
				o.List = append(o.List, co)
			}
			return nil
		}
	}
	err := utl.ParseNode(c, env, e.Args, o, r)
	if err != nil {
		return e, err
	}
	return utl.GetNode(o)
}

func getPtr(e exp.El) interface{} {
	if a, ok := e.(utl.Node); ok {
		return a.Ptr()
	}
	return nil
}

func inheritAttr(o, c *Node) {
	if c.Align == 0 {
		c.Align = o.Align
	}
	if c.Font == nil {
		c.Font = o.Font
	}
}

func resolveStr(c *exp.Ctx, env exp.Env, xs []exp.El) (string, error) {
	if len(xs) == 0 {
		return "", nil
	}
	x := xs[0]
	if len(xs) > 1 {
		x = exp.Dyn(xs)
	}
	dl, err := c.Resolve(env, x)
	if err != nil {
		return "", err
	}
	ch, ok := dl.(lit.Charer)
	if !ok {
		return "", errors.New("data is not a char literal")
	}
	return ch.Char(), nil
}
