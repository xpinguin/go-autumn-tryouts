package main

import (
	"log"
	//"time"
)

// PageTargets :: CDPClient -> Context -> Stream Target -> error
func PageTargets(cl *CDPClient, ctx Context, pages chan<- Target) (err error) {
	ts, err := cl.ListPageTargets(ctx)
	if err != nil {
		return
	}
	// --
	for _, t := range ts {
		pages <- t
	}
	return
}

// IterPageTargets :: CDPClient -> Context -> Stream Target
func IterPageTargets(cl *CDPClient, ctx Context) <-chan Target {
	pages := make(chan Target)
	// --
	go func() {
		err := PageTargets(cl, ctx, pages)
		////
		if err != nil {
			log.Printf("{ERR} PageTargets(...): err = %v", err)
		}
		close(pages)
	}()
	// --
	return pages
}

/* TODO
func TestListTabs() {
	// list open tabs (page-targets)
	ts, err := cclient.ListPageTargets(ctx)
	if err != nil {
		log.Printf("{ERR} ListPageTargets(...): err = %v", err)
	}
	//for pt := range IterPageTargets(cclient, ctx) {
	for _, pt := range ts {
		//cclient.ActivateTarget(ctx, pt)
		c.AddTarget(ctx, pt)
	}
	<-time.After(10 * time.Second) // FIXME
	for i, ptID := range c.ListTargets() {
		// --
		var url string
		err := c.Run(ctx, chro.Tasks{
			//c.SetTargetByID(ptID),
			c.SetTarget(i),
			chro.Evaluate("document.location.toString()", &url),
		})
		if err != nil {
			log.Printf("[%d]/%v CDP.Run(...): err = %v", i, ptID, err)
		} else {
			log.Printf("[%d]/%v %s", i, ptID, url)
		}
	}
}*/
