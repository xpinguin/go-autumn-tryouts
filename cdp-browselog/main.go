package main

import (
	Ctx "context"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"

	//"github.com/chromedp/cdproto/domsnapshot"
	"github.com/chromedp/chromedp"
	cdpclient "github.com/chromedp/chromedp/client"
)

/*func ByQueryAll(s *Selector) {
	ByFunc(func(ctxt context.Context, h *TargetHandler, n *cdp.Node) ([]cdp.NodeID, error) {
		return dom.QuerySelectorAll(n.NodeID, s.selAsString()).Do(ctxt, h)
	})(s)
}*/

func main() {
	//
	//ctx, ctx_cancel := Ctx.WithCancel(Ctx.Background())
	ctx := Ctx.Background()

	//
	_cdev_client := cdpclient.New(cdpclient.URL("http://localhost:9851/json"))
	cdev, err := chromedp.New(ctx, chromedp.WithClient(ctx, _cdev_client))
	if cdev == nil {
		log.Fatalf("!cdev; err = %v", err)
	} else if err != nil {
		log.Printf("cdev = %v; err = %v", cdev, err)
	}

	// list open tabs (page-targets)
	// --
	/*ListPageTargets := func(ctx Ctx.Context) []cdpclient.Target {
		ts, _ := _cdev_client.ListPageTargets(ctx)
		return ts
	}*/

	// -- revert back current tab
	/*cdev.Lock()
	active_handler := cdev.cur
	cdev.Unlock()
	var active_page cdpclient.Target

	for i, page := range ListPageTargets(ctx) {
		if active_handler == cdev.GetHandlerByIndex(i) {
			active_page = page
			break
		}
	}
	defer cdev.SetHandlerByID(active_page.GetID())*/

	// -- iterate through all
	/*for i, page := range ListPageTargets(ctx) {
		if err := _cdev_client.ActivateTarget(ctx, page); err != nil {
			log.Fatalf("ActivateTarget(%v): err = %v", page, err)
		}
		// --
		var page_url string
		cdev.Run(ctx, chromedp.Tasks{
			cdev.SetTarget(i),
			chromedp.Evaluate("document.location.toString()", &page_url),
		})
		log.Printf("[%d] %s {target-id:%v}", i, page_url, page.GetID())

	}*/

	// box model (for the current target?)
	var page_url string
	//var boxmodel *dom.BoxModel

	err = cdev.Run(ctx, chromedp.Tasks{
		cdev.SetTarget(0),
		chromedp.Evaluate("document.location.toString()", &page_url),
		//chromedp.WaitVisible("/*", chromedp.ByQueryAll),
		//chromedp.Dimensions("div", &boxmodel, chromedp.BySearch),
		chromedp.QueryAfter(
			"*",
			func(ctxt Ctx.Context, h *chromedp.TargetHandler, nodes ...*cdp.Node) error {
				if len(nodes) < 1 {
					return fmt.Errorf("selector did not return any nodes")
				}
				for _, n := range nodes {
					nbox, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(ctxt, h)
					if err != nil {
						log.Printf("{ERR} <%s>: %v", n.NodeName, err)
						continue
					}
					// --
					log.Printf("<%s>: %v", n.NodeName, nbox.Content)
				}
				return nil
			},
			chromedp.BySearch,
		),
	})
	if err != nil {
		log.Fatalf("{ERR} Run: err = %v", err)
	}

	/*log.Printf("URL: %s ;; BOXMODEL: %p", page_url, boxmodel)
	//log.Printf("BOXMODEL OUTER SIZE: %d", len(boxmodel))

	if boxmodel == nil {
		panic("!boxmodel")
	}
	fmt.Printf("\n\n----------------------------\n")
	fmt.Println(*boxmodel)
	fmt.Println(boxmodel.ShapeOutside)*/

	// DOM snapshot
	/*cstyles := make([]string)
	domcapture_cmd := domsnapshot.CaptureSnapshot(cstyles)
	// --
	if err := cdev.Run(ctx, domcapture_cmd); err != nil {
		log.Fatalf("RUN: domsnapshot.CaptureSnapshot(): err = %v", err)
	}*/
}
