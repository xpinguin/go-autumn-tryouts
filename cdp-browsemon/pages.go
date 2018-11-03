package browsemon

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	//"github.com/chromedp/chromedp"
)

type (
	DOMNode = cdp.Node
)

type LocationDOMTree struct {
	Loc  string
	Tree <-chan *DOMNode
}

type PageContext struct {
	page *TargetHandler

	ctx       Context
	ctxCancel *context.CancelFunc

	domNodesStream chan LocationDOMTree // TODO: locking
}

// :: TargetHandler -> Maybe BrowserContext -> PageContext
func NewPageContext(page *TargetHandler, browser *BrowserContext) PageContext {
	if page == nil {
		panic("page TargetHandler must be specified")
	}

	var ctx Context
	var ctxCancel context.CancelFunc
	if browser != nil {
		ctx = browser.ctx
		ctxCancel = *browser.ctxCancel
	} else {
		ctx, ctxCancel = context.WithCancel(context.Background())
	}

	return PageContext{page, ctx, &ctxCancel, nil}
}

func (p *PageContext) Location() (r string) {
	err := chromedp.Location(&r).Do(p.ctx, p.page)
	if err != nil {
		panic(err)
	}
	return
}

// :: PageContext -> Stream (Location, Stream DOMNode)
func (p *PageContext) DOMNodes() <-chan LocationDOMTree {
	if p.domNodesStream != nil {
		return p.domNodesStream
	}
	p.domNodesStream = make(chan LocationDOMTree)

	// FIXME: memory leak
	go func(loctree chan<- LocationDOMTree) {
		defer func() {
			close(p.domNodesStream)
			p.domNodesStream = nil
		}()
		/// --
		for {
			done := make(chan struct{})
			nodes := make(chan *DOMNode)
			lns := LocationDOMTree{p.Location(), nodes}

			/// --
			go func(ns chan<- *DOMNode, done <-chan struct{}) {
				defer func() {
					close(ns)
					<-done
				}()
				///
				root, err := p.page.GetRoot(p.ctx)
				if err != nil {
					log.Println("{ERR} TargetHandler.GetRoot(ctx): ", err)
					return
				}
				///
				err = dom.RequestChildNodes(root.NodeID).WithDepth(-1).Do(p.ctx, p.page)
				if err != nil {
					log.Println("{ERR} dom.RequestChildNodes(...): ", err)
					return
				}
				/// traverse DOM-tree breadth-first
				front := []*DOMNode{root}
				for len(front) > 0 {
					n := front[0]
					select {
					case ns <- n:
					case _, ok := <-done:
						if ok {
							return
						}
					}
					front = append(front[1:], n.Children...)
				}
			}(nodes, done)
			/// --
			loctree <- lns
			/// --
			for sameLoc := true; sameLoc; {
				select {
				case <-time.After(time.Second):
					if p.Location() != lns.Loc {
						log.Printf("%v != %v", p.Location(), lns.Loc)
						sameLoc = false
					}
				}
			}
			done <- struct{}{}
			close(done)
		}
	}(p.domNodesStream)

	return p.domNodesStream
}

type DOMNodeBoxModel struct {
	n   *DOMNode
	box *dom.BoxModel
}

// :: Stream DOMNode -> Stream (DOMNode, dom.BoxModel)
func (p *PageContext) BoxModels(ns <-chan *DOMNode) <-chan DOMNodeBoxModel {
	bs := make(chan DOMNodeBoxModel)
	go func(bs chan<- DOMNodeBoxModel) {
		defer func() {
			close(bs)
		}()
		///
		for n := range ns {
			nbox, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(p.ctx, p.page)
			if err != nil {
				//log.Printf("{ERR} <%s>: %v", n.NodeName, err)
				continue
			}
			bs <- DOMNodeBoxModel{n, nbox}
		}
	}(bs)
	return bs
}
