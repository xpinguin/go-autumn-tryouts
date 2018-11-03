package browsemon

import (
	"fmt"
	//"log"
	T "testing"
)

func TestPageDOMTree(t *T.T) {
	b := dialBrowser()
	func() {
		for th := range b.PageHandlers() {
			p := NewPageContext(th, &b)
			for locns := range p.DOMNodes() {
				fmt.Printf("LOCATION: %s\n\n", locns.Loc)
				for nb := range p.BoxModels(locns.Tree) {
					fmt.Printf("'%v'\n", nb.box.Content)
				}
			}
			fmt.Printf("\n---------\n")
			break
		}
	}()
	//sleep(5)
}
