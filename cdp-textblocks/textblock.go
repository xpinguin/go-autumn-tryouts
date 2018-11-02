package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	//"github.com/mndrix/golog"
	"github.com/mndrix/golog/term"
)

////
//type Rect = []float64

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

////
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

////
type Block struct {
	nid DOMNodeID
	Rect
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

////
type DOMNodeRectReader struct {
	nboxes <-chan DOMNodeBoxModel
}

func (rr DOMNodeRectReader) Read(s []byte) (int, error) {
	//defer fmt.Println(s)
	//defer fmt.Printf("%s", s)

	if rr.nboxes == nil {
		log.Fatalln("{FATAL} !rr.nboxes")
	}

	var nbox DOMNodeBoxModel
	var nrect Rect
	var ok bool

	for !ok || nrect == nil {
		nbox, ok = <-rr.nboxes
		//log.Println(nbox, nrect, ok)
		if !ok {
			return 0, io.EOF
		}
		// --
		nrect, ok = DOMNodeRect(nbox)
		//log.Println(nbox, nrect, ok)
	}

	// --
	var nvterm string
	if nv := nbox.text; len(nv) > 0 && len(nv) < 100 {
		/// ! FIXME: Golog's TermReader requires `s` to be less than 1000 symbols
		/// TODO: implement as TermReader, rather than io.Reader
		///
		/// TODO: use ForeignPredicate for the node value: `node_value`/2
		///
		nvcodes := term.NewCodeListFromDoubleQuotedString(
			strings.Replace(`"`+nv+`"`, "\n", "\\n", -1))
		nvterm = fmt.Sprintf(`node(%d, value(%s)).`+"\n", nbox.n.NodeID, nvcodes.String())
	}

	// --
	nboxterm := fmt.Sprintf("node(%d, rect(%.1f, %.1f, %.1f, %.1f)).\n", nbox.n.NodeID,
		nrect[0], nrect[1], nrect[2], nrect[3])

	// --
	nterms := nboxterm + nvterm
	//fmt.Print(nterms) /// temporary!
	copy(s, nterms)
	return len(nterms), nil
}

//// TEST
func TestTextBlock(ctx Context, c *CDP, cclient *CDPClient) {
	pl := NewDOMMachine(c, ctx)

	var pageURL string
	var err error
	boxes := make(chan DOMNodeBoxModel)

	go func() {
		defer close(boxes)
		////
		pageURL, err = TargetBoxModels_(c, ctx, 0, boxes)
		if err != nil {
			log.Printf("{ERR} TargetBoxModel(...): err = %v", err)
		}
	}()

	//pl = pl.Consult(DOMNodeRectReader{boxes})
	//log.Println(pl.ProveAll("node(Id, rect(0.0, T, W, H))."))

	log.Printf("^^^^^ URL: %s ^^^^^\n", pageURL)
	//fmt.Println(pl.ProveAll("listing."))

	RunPrologCLI(pl, os.Stdin, os.Stdout)
}
