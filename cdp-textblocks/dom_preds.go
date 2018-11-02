// Prolog predicates for querying the DOM tree
package main

import (
	"log"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/term"
	//"github.com/chromedp/cdproto/dom"
)

type (
	Machine          = golog.Machine
	Term             = term.Term
	ForeignReturn    = golog.ForeignReturn
	ForeignPredicate = golog.ForeignPredicate
)

func NewDOMMachine(c *CDP, ctx Context) Machine {
	m := golog.NewMachine()
	return m.RegisterForeign(map[string]ForeignPredicate{
		// cdp(-GoValue<*CDP>)
		"cdp/1": func(m Machine, args []Term) ForeignReturn {
			if r := args[0]; term.IsVariable(r) {
				return golog.ForeignUnify(r, GoValue{c})
			}
			panic("cdp/1: argument must a variable")
		},

		// cdpcontext(-GoValue<Context>)
		"cdpcontext/1": func(m Machine, args []Term) ForeignReturn {
			if r := args[0]; term.IsVariable(r) {
				return golog.ForeignUnify(r, GoValue{ctx})
			}
			panic("cdpcontext/1: argument must a variable")
		},

		//
		"node_content/2":     NodeContent2,
		"node_value/2":       NodeContent2,
		"node_boxmodel/2":    NodeBoxModel2,
		"node_contentquad/2": NodeContentQuad2,
	})
}

func getCDPContext(m Machine) (cdp *CDP, ctx Context) {
	vcdp := term.NewVar("_")
	vctx := term.NewVar("_")
	g := term.NewCallable(",",
		term.NewCallable("cdp", vcdp),
		term.NewCallable("cdpcontext", vctx))

	bs := m.ClearConjs().ClearDisjs().ProveAll(g)
	if len(bs) == 0 {
		panic("No CDP context in the given machine")
	}

	cdp = bs[0].Resolve_(vcdp).(GoValue).AsInterface().(*CDP)
	ctx = bs[0].Resolve_(vctx).(GoValue).AsInterface().(Context)
	return
}

// node_content(+NodeID, ?Text)
func NodeContent2(m Machine, args []Term) ForeignReturn {
	if !term.IsInteger(args[0]) {
		panic("node_content/2: first argument must be an integer NodeID")
	}
	//nodeId := args[0].(*term.Integer)

	c, ctx := getCDPContext(m)
	log.Println(c, ctx)
	//n, err := dom.DescribeNode().WithNodeID(nodeId).Do(ctx, h)

	return golog.ForeignFail()
}

// node_contentquad(+NodeID, ?Text)
func NodeContentQuad2(m Machine, args []Term) ForeignReturn {
	return golog.ForeignFail()
}

func NodeBoxModel2(m Machine, args []Term) ForeignReturn {
	return golog.ForeignFail()
}
