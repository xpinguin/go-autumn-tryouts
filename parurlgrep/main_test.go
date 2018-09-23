package main

import T "testing"
import S "strings"
import "fmt"

/// TEST DATA ///
var (
	URLs = []string{
		"https://golang.org",
		"http://ya.ru",
		"https://google.com",
		"hjhjhjhdfg",
	}
)

/// TESTS ///
func TestUrlsChannel(t *T.T) {
	input := NewStringReader(S.Join(URLs, "\n"))
	// --
	url_i := 0
	for url := range StartUrlsChannel(input) {
		if URLs[url_i] != url {
			t.Logf("Invalid URL order at %d: (out) '%s' != (in) '%s'",
				url_i, url, URLs[url_i])
			t.Fail()
		}
		url_i++
	}
	if url_i != len(URLs) {
		t.Logf("Inequal number of URLs: (out) %d != (in) %d",
			url_i, len(URLs))
		t.Fail()
	}
}



func TestConcurrentUrlMatcher(t *T.T) {
	// --
	match_re := ReCompile("Go")

	// --
	input := NewStringReader(S.Repeat(S.Join(URLs, "\n")+"\n", 7))
	urls_chan := StartUrlsChannel(input)
	
	// --
	type URLMatches struct {
		url string
		url_has_data bool
		matches_num int
	}
	urlmatch_chan := make(chan URLMatches, 2)
	
	// --
	urls_chan_closed := false
	tasks_scheduled := 0
	
	
	for tasks_scheduled > 0 || !urls_chan_closed {
		//fmt.Println(tasks_scheduled)
		select {
			case m, ok := <-urlmatch_chan:
				//fmt.Printf("fgfgfg")
				if !ok {
					tasks_scheduled = 0
					continue
				}
				if m.url_has_data {
					fmt.Printf("| %s: %d\n", m.url, m.matches_num)
				} else {
					fmt.Printf("| %s: NO DATA\n", m.url)
				}
				tasks_scheduled--
				
			case url, ok := <-urls_chan:
				//fmt.Printf(">> url (%t): %s\n", ok, url)
				if !ok {
					urls_chan_closed = true
					continue
				}
				
				tasks_scheduled++
				go func (url string, resch chan<- URLMatches) {
					url_data := UrlData(url)
					if url_data == nil {
						resch <- URLMatches{url, false, 0}
						return
					}
					resch <- URLMatches{url, true, ReCountMatches(match_re, url_data)}
				}(url, urlmatch_chan)
				
		}
	}
	close(urlmatch_chan)
}