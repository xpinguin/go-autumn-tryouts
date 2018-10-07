package main

import (
	Ctx "context"
	"log"

	//"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	cdpclient "github.com/chromedp/chromedp/client"
)

func main() {
	//
	//ctx, ctx_cancel := Ctx.WithCancel(Ctx.Background())
	ctx := Ctx.Background()

	//
	_cdev_client := cdpclient.New(cdpclient.URL("http://localhost:9222/json"))
	cdev, err := chromedp.New(ctx, chromedp.WithClient(ctx, _cdev_client))
	if cdev == nil {
		log.Fatalf("!cdev; err = %v", err)
	} else if err != nil {
		log.Printf("cdev = %v; err = %v", cdev, err)
	}

	// list open tabs (page-targets)
	// --
	ListPageTargets := func(ctx Ctx.Context) []cdpclient.Target {
		ts, _ := _cdev_client.ListPageTargets(ctx)
		return ts
	}

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
	for i, page := range ListPageTargets(ctx) {
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

	}

	// further actions
}
