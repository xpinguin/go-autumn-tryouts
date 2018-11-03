package browsemon

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/client"
)

type (
	Context       = context.Context
	CDPClient     = client.Client
	Target        = client.Target
	TargetHandler = chromedp.TargetHandler
)

type BrowserContext struct {
	browser *CDPClient

	ctx       Context
	ctxCancel *context.CancelFunc

	targetsStream <-chan Target
}

func DialBrowser(host string, port int) BrowserContext {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return BrowserContext{
		client.New(client.URL(fmt.Sprintf("http://%s:%d/json", host, port))),
		ctx, &ctxCancel,
		nil}
}

// :: BrowserContext -> Stream Target
func (b BrowserContext) PageTargetsStream() <-chan Target {
	if b.targetsStream == nil {
		b.targetsStream = b.browser.WatchPageTargets(b.ctx)
	}
	return b.targetsStream
}

// :: Target -> TargetHandler
func (b BrowserContext) RunTargetHandler(t Target) *TargetHandler {
	h, err := chromedp.NewTargetHandler(t, log.Printf, log.Printf, log.Fatalf)
	if err != nil {
		log.Printf("{ERR} NewTargetHandler(...): %v", err)
		return nil
	}
	if err := h.Run(b.ctx); err != nil {
		log.Printf("{ERR} TargetHandler.Run(ctx): %v", err)
		return nil
	}
	return h
}
