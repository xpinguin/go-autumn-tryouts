package main

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
