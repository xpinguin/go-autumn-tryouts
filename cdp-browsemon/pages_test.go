package browsemon

import (
	"fmt"
	T "testing"
)

func TestPageDOMTree(t *T.T) {
	b := dialBrowser()
	func() {
		for th := range b.PageHandlers() {
			go func(th *TargetHandler) {
				p := NewPageContext(th, &b)
				for locns := range p.DOMNodes() {
					for nb := range p.BoxModels(locns.Tree) {
						fmt.Printf("<%s> %s[%d]: %v\n", locns.Loc,
							nb.n.NodeName, nb.n.NodeID, nb.box.Content)
					}
				}
			}(th)
		}
	}()
	//sleep(5)
}
