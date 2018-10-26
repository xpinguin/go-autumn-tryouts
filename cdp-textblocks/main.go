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
	//Target = chroclient.Target
)

type DOMNodeBoxModel struct {
	n   *DOMNode
	box *dom.BoxModel
}

// ! FIXME
// TargetBoxModel :: CDP -> Context -> Target/int -> (URL, []DOMNodeBoxModel, error)
func TargetBoxModel(c *chro.CDP, ctx Context, targetIndex int) (pageURL string, boxes []DOMNodeBoxModel, err error) {
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
					boxes = append(boxes, DOMNodeBoxModel{n, nbox})
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

//
func DumpTargetBoxModel(c *chro.CDP, ctx Context, targetIndex int) (pageURL string, err error) {
	pageURL, boxes, err := TargetBoxModel(c, ctx, targetIndex)
	if err != nil {
		return
	}
	// --
	for _, nbox := range boxes {
		fmt.Printf("[%v] <%s>: %v\n", nbox.n.NodeID, nbox.n.NodeName, nbox.box.Content)
	}
	return
}

func main() {
	//
	//ctx, ctx_cancel := context.WithCancel(context.Background())
	ctx := context.Background()

	//
	cclient := chroclient.New(chroclient.URL("http://localhost:9851/json"))
	c, err := chro.New(ctx, chro.WithClient(ctx, cclient))
	if c == nil {
		log.Fatalf("!c; err = %v", err)
	} else if err != nil {
		log.Printf("c = %v; err = %v", c, err)
	}

	// list open tabs (page-targets)
	// --
	/*ListPageTargets := func(ctx Context) []chroclient.Target {
		ts, _ := cclient.ListPageTargets(ctx)
		return ts
	}*/

	// -- revert back current tab
	/*c.Lock()
	active_handler := c.cur
	c.Unlock()
	var active_page chroclient.Target

	for i, page := range ListPageTargets(ctx) {
		if active_handler == c.GetHandlerByIndex(i) {
			active_page = page
			break
		}
	}
	defer c.SetHandlerByID(active_page.GetID())*/

	// -- iterate through all
	/*for i, page := range ListPageTargets(ctx) {
		if err := _c_client.ActivateTarget(ctx, page); err != nil {
			log.Fatalf("ActivateTarget(%v): err = %v", page, err)
		}
		// --
		var page_url string
		c.Run(ctx, chro.Tasks{
			c.SetTarget(i),
			chro.Evaluate("document.location.toString()", &page_url),
		})
		log.Printf("[%d] %s {target-id:%v}", i, page_url, page.GetID())

	}*/

	// box model (for the current target)
	url, err := DumpTargetBoxModel(c, ctx, 0)
	if err != nil {
		log.Fatalf("{ERR} Run: err = %v", err)
	}
	fmt.Printf("^^^^^ URL: %s ^^^^^\n", url)
}
