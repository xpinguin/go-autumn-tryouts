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

type DOMNodeBoxModel struct {
	n   *DOMNode
	box *dom.BoxModel
}

type Block struct {
	nid DOMNodeID
	Rect
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

// :: DOMNode -> Rect
func DOMNodeRect(nbox DOMNodeBoxModel) (Rect, bool) {
	if nbox.box == nil {
		return nil, false
	}
	return nbox.box.Content, true
}

// :: Rect -> Rect -> bool
func RectWithinRect(inner Rect, outer Rect) bool {
	if (inner[0] >= outer[0]) && (inner[1] >= outer[1]) &&
		(inner[2] <= outer[2]) && (inner[3] <= outer[3]) {
		return true
	}
	return false
}

//
type RectTreeNode struct {
	r  Rect
	cn []*RectTreeNode
}

func (rtn RectTreeNode) Empty() bool {
	return rtn.r == nil
}

func (rtn RectTreeNode) Within(outer *RectTreeNode) bool {
	return RectWithinRect(rtn.r, outer.r)
}

// BoxModelsBlocks :: Stream DOMNodeBoxModel -> Stream Block -> ()
func BoxModelsBlocks(boxes <-chan DOMNodeBoxModel, blocks chan<- Block) {
	nrects := make(map[DOMNodeID]*RectTreeNode, 2000)

	for nbox := range boxes {
		r, ok := DOMNodeRect(nbox)
		if !ok {
			continue
		}
		/// FIXME
		nid := nbox.n.NodeID
		rtn := &RectTreeNode{r, nil}
		for _, rtn0 := range nrects {
			if rtn0.Within(rtn) {
				rtn.cn = append(rtn.cn, rtn0)
			} else if rtn.Within(rtn0) {
				rtn0.cn = append(rtn0.cn, rtn)
			}
			nrects[nid] = rtn
		}

	}

	/*// FIXME: O(n*n) -> O(n*log(n))
	//nrects := make(map[DOMNodeID]Rect, 2000)
	for nid0, r0 := range nrects {
		for nid1, r1 := range nrects {
			if nid1 == nid0 {
				continue
			}
			////
		}
	}*/
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

	// box model (for the current target)
	url, err := DumpTargetBoxModels(c, ctx, 0)
	if err != nil {
		log.Fatalf("{ERR} Run: err = %v", err)
	}
	fmt.Printf("^^^^^ URL: %s ^^^^^\n", url)
}
