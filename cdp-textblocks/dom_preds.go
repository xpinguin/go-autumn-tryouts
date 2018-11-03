// Prolog predicates for querying the DOM tree
package main

import (
	"fmt"
	"log"

	"github.com/mndrix/golog"
	"github.com/mndrix/golog/term"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	chro "github.com/chromedp/chromedp"
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
	/*if !term.IsString(args[0]) {
		panic("node_content/2: first argument must be an XPath")
	}
	nodeXPath := args[0].String()
	log.Printf("nodeXPath: %s", nodeXPath)*/

	if !term.IsInteger(args[0]) {
		panic("node_content/2: first argument must be an integer NodeID")
	}
	nodeId := DOMNodeID(args[0].(*term.Integer).Value().Int64())

	c, ctx := getCDPContext(m)
	var nodeText string

	err := c.Run(ctx, chro.Tasks{
		//c.SetTarget(0),
		chro.ActionFunc(func(ctx Context, h cdp.Executor) (err error) {
			n, err := dom.DescribeNode().WithNodeID(nodeId).Do(ctx, h)
			if err != nil {
				log.Println("{ERR}", err)
				return
			}
			err = chro.EvaluateAsDevTools(fmt.Sprintf(textJS, n.FullXPath()), &nodeText).Do(ctx, h)
			log.Printf("XPath = %v ;; nodeText = %v ;; err = %v", n.FullXPath(), nodeText, err)
			return
		}),
	})
	if err != nil {
		panic(fmt.Sprintf("{ERR} CDP.Run(...): %v", err))
	}

	/*err := c.Run(ctx, chro.EvaluateAsDevTools(fmt.Sprintf(textJS, nodeXPath), &nodeText))
	if err != nil {
		panic(fmt.Sprintf("{ERR} CDP.Run(...): %v", err))
	}*/

	return golog.ForeignUnify(args[1], term.NewCodeList(nodeText))
}

// node_contentquad(+NodeID, ?dom.Quad)
func NodeContentQuad2(m Machine, args []Term) ForeignReturn {
	if !term.IsInteger(args[0]) {
		panic("node_content/2: first argument must be an integer NodeID")
	}
	nodeId := DOMNodeID(args[0].(*term.Integer).Value().Int64())

	c, ctx := getCDPContext(m)
	var q dom.Quad

	err := c.Run(ctx, chro.Tasks{
		c.SetTarget(0),
		chro.ActionFunc(func(ctx Context, h cdp.Executor) (err error) {
			nbox, err := dom.GetBoxModel().WithNodeID(nodeId).Do(ctx, h)
			if err == nil {
				q = nbox.Content
			}
			return
		}),
	})
	if err != nil {
		panic(fmt.Sprintf("{ERR} CDP.Run(...): %v", err))
	}

	var qterm term.Term
	if len(q) == 0 {
		qterm = term.NewAtom("NoQuad")
	} else {
		qterm = term.NewCallable("Quad", term.NewFloat64(q[0]), term.NewFloat64(q[1]), term.NewFloat64(q[2]), term.NewFloat64(q[3]))
	}
	return golog.ForeignUnify(args[1], qterm)
}

// node_contentquad(+NodeID, ?dom.BoxModel)
func NodeBoxModel2(m Machine, args []Term) ForeignReturn {
	return golog.ForeignFail()
}
