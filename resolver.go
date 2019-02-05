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
	l, tail, err := utl.ParseNode(c, env, e, o)
	if err != nil {
		return e, err
	}
	// resolve tail
	if hasData(o.Kind) {
		o.Data, err = resolveStr(c, env, tail)
		if err != nil {
			return e, fmt.Errorf("%v on resolveStr", err)
		}
	} else if hasList(o.Kind) {
		for _, child := range tail {
			cl, err := c.Resolve(env, child)
			if err != nil {
				return e, fmt.Errorf("%v, on resolve child %s", err, child)
			}

			co, ok := getPtr(cl).(*Node)
			if !ok {
				return e, fmt.Errorf("child is not a layla object %T", cl)
			}
			inheritAttr(o, co)
			o.List = append(o.List, co)
		}
	} else if len(tail) > 0 {
		return e, exp.ErrRogueTail
	}
	return l, nil
}

func getPtr(e exp.El) interface{} {
	if a, ok := e.(lit.Assignable); ok {
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
