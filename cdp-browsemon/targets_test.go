package browsemon

import (
	"log"
	T "testing"
	"time"
)

func dialBrowser() BrowserContext {
	return DialBrowser("localhost", 9851)
}

func sleep(seconds int) {
	time.Sleep(time.Second * time.Duration(seconds))
}

func TestListTabs(t *T.T) {
	b := dialBrowser()
	go func() {
		for th := range b.PageHandlers() {
			log.Printf("<PageHandler>: %v", th)
		}
	}()
	sleep(5)
}

func TestTabRoot(t *T.T) {
	b := dialBrowser()
	go func() {
		for th := range b.PageHandlers() {
			root, err := th.GetRoot(b.ctx)
			if err != nil {
				t.Fatal("{ERR} GetRoot(ctx): ", err)
			}
			log.Printf("root DOM Node: %v", root)
			log.Printf("root children: %v", root.Children)
		}
	}()
	sleep(5)
}
