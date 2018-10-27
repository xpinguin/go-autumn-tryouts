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
	Rect      = dom.Quad
)

////
type DOMNodeBoxModel struct {
	n   *DOMNode
	box *dom.BoxModel
}

// TargetBoxModels :: CDP -> Context -> Target/int -> Stream DOMNodeBoxModel ->(URL, error)
func TargetBoxModels(c *CDP, ctx Context, targetIndex int, boxes chan<- DOMNodeBoxModel) (pageURL string, err error) {
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
					nbox, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(ctxt, h)
					if err != nil {
						//log.Printf("{ERR} <%s>: %v", n.NodeName, err)
						continue
					}
					boxes <- DOMNodeBoxModel{n, nbox}
					// --
					//fmt.Printf("<%s>: %v\n", n.NodeName, nbox.Content)
				}
				return nil
			},
			chro.BySearch,
		),
	})
	return
}

// DumpTargetBoxModels :: CDP -> Context -> Target/int ->(URL, error)   ????: -> ()
func DumpTargetBoxModels(c *CDP, ctx Context, targetIndex int) (pageURL string, err error) {
	boxes := make(chan DOMNodeBoxModel)
	// --
	go func() {
		pageURL, err = TargetBoxModels(c, ctx, targetIndex, boxes)
		////
		if err != nil {
			log.Printf("{ERR} TargetBoxModel(..., targetIndex = %d, ...): err = %v", err)
		}
		close(boxes)
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
	url, err := DumpTargetBoxModels(c, ctx, 0)
	if err != nil {
		log.Fatalf("{ERR} Run: err = %v", err)
	}
	fmt.Printf("^^^^^ URL: %s ^^^^^\n", url)
}
