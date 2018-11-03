package browsemon

import (
	"log"
	T "testing"
)

func TestListTabs(t *T.T) {
	b := DialBrowser("localhost", 9851)
	for th := range b.PageHandlersStream() {
		log.Printf("<PageHandler>: %v", th)
	}
}
