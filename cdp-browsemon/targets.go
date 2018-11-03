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

	targetsStream  <-chan Target
	handlersStream chan *TargetHandler
}

func DialBrowser(host string, port int) BrowserContext {
	ctx, ctxCancel := context.WithCancel(context.Background())
	return BrowserContext{
		client.New(client.URL(fmt.Sprintf("http://%s:%d/json", host, port))),
		ctx, &ctxCancel,
		nil, nil}
}

// :: BrowserContext -> Stream Target
func (b BrowserContext) PageTargets() <-chan Target {
	if b.targetsStream == nil {
		b.targetsStream = b.browser.WatchPageTargets(b.ctx)
	}
	return b.targetsStream
}

// :: Target -> TargetHandler
func (b BrowserContext) RunTargetHandler(t Target) *TargetHandler {
	h, err := chromedp.NewTargetHandler(t, DummyPrintf, DummyPrintf, log.Fatalf)
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

// :: BrowserContext -> Stream TargetHandler
func (b BrowserContext) PageHandlers() <-chan *TargetHandler {
	if b.handlersStream != nil {
		return b.handlersStream
	}
	b.handlersStream = make(chan *TargetHandler)

	go func(ts <-chan Target, hs chan<- *TargetHandler) {
		defer func() {
			close(b.handlersStream)
			b.handlersStream = nil
		}()
		///
		for t := range ts {
			h := b.RunTargetHandler(t)
			if h != nil {
				hs <- h
			}
		}
	}(b.PageTargets(), b.handlersStream)

	return b.handlersStream
}
