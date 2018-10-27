package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	chro "github.com/chromedp/chromedp"
	chroclient "github.com/chromedp/chromedp/client"
)

type (
	Context       = context.Context
	DOMNode       = cdp.Node
	TargetHandler = chro.TargetHandler
	Target        = chroclient.Target

	CDP       = chro.CDP
	CDPClient = chroclient.Client

	DOMNodeID = cdp.NodeID

	// FIXME: dom.Quad <=> set of corner vertices ~[4]{x, y}; !!!!  IT IS NOT A Rect!
	Rect = dom.Quad
)

const (
	// NB. from `chromedp/js.go`
	// textJS is a javascript snippet that returns the concatenated textContent
	// of all visible (ie, offsetParent !== null) children.
	textJS = `(function(a) {
		var s = '';
		for (var i = 0; i < a.length; i++) {
			if (a[i].offsetParent !== null) {
				s += a[i].textContent;
			}
		}
		return s;
	})($x('%s/node()'))`
)

////
type DOMNodeBoxModel struct {
	n    *DOMNode
	box  *dom.BoxModel
	text string
}

// TargetBoxModels_ :: CDP -> Context -> Target/int -> Stream DOMNodeBoxModel ->(URL, error)
func TargetBoxModels_(c *CDP, ctx Context, targetIndex int, boxes chan<- DOMNodeBoxModel) (pageURL string, err error) {
	err = c.Run(ctx, chro.Tasks{
		c.SetTarget(targetIndex),
		chro.Evaluate("document.location.toString()", &pageURL),
		chro.QueryAfter(
			"*",
			func(ctxt Context, h *TargetHandler, nodes ...*DOMNode) error {
				if len(nodes) < 1 {
					return fmt.Errorf("selector did not return any nodes")
				}
				for _, n := range nodes {
					// -- box model
					nbox, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(ctxt, h)
					if err != nil {
						//log.Printf("{ERR} <%s>: %v", n.NodeName, err)
						continue
					}
					// -- text content
					var ntext string
					chro.EvaluateAsDevTools(fmt.Sprintf(textJS, n.FullXPath()),
						&ntext).Do(ctxt, h)
					// --
					boxes <- DOMNodeBoxModel{n, nbox, ntext}
				}
				return nil
			},
			chro.BySearch,
		),
	})
	return
}

// TargetBoxModels :: CDP -> Context -> Target/int -> (Stream DOMNodeBoxModel, URL, error)
func TargetBoxModels(c *CDP, ctx Context, targetIndex int) (boxes <-chan DOMNodeBoxModel, pageURL string, err error) {
	boxes_ := make(chan DOMNodeBoxModel)
	boxes = boxes_
	// --
	go func() {
		defer close(boxes_)
		////
		pageURL, err = TargetBoxModels_(c, ctx, targetIndex, boxes_)
		if err != nil {
			log.Printf("{ERR} TargetBoxModel(..., targetIndex = %d, ...): err = %v", targetIndex, err)
		}
	}()
	// --
	return // FIXME: `pageURL` and `err` won't work properly
}

// DumpTargetBoxModels :: CDP -> Context -> Target/int ->(URL, error)   ????: -> ()
func DumpTargetBoxModels(c *CDP, ctx Context, targetIndex int) (pageURL string, err error) {
	boxes := make(chan DOMNodeBoxModel)
	// --
	go func() {
		defer close(boxes)
		////
		pageURL, err = TargetBoxModels_(c, ctx, targetIndex, boxes)
		if err != nil {
			log.Printf("{ERR} TargetBoxModel(..., targetIndex = %d, ...): err = %v", targetIndex, err)
		}
	}()
	// --
	for nbox := range boxes {
		fmt.Printf("[%v] <%s>: %v\n", nbox.n.NodeID, nbox.n.NodeName, nbox.box.Content)
	}
	return
}

////
func main() {
	//ctx, ctx_cancel := context.WithCancel(context.Background())
	ctx := context.Background()

	cclient := chroclient.New(chroclient.URL("http://localhost:9851/json"))
	c, err := chro.New(ctx, chro.WithClient(ctx, cclient))
	if c == nil {
		log.Fatalf("!c; err = %v", err)
	} else if err != nil {
		log.Printf("c = %v; err = %v", c, err)
	}

	// box model (for the current target)
	TestTextBlock(ctx, c, cclient)
	/*url, err := DumpTargetBoxModels(c, ctx, 0)
	if err != nil {
		log.Fatalf("{ERR} Run: err = %v", err)
	}
	fmt.Printf("^^^^^ URL: %s ^^^^^\n", url)*/
}
