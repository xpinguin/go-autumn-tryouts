package main

import T "testing"
import S "strings"

//import "fmt"

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
	input := S.NewReader(S.Join(URLs, "\n"))
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
	input := S.NewReader(S.Repeat(S.Join(URLs, "\n")+"\n", 7))

	UrlMatchCountParallel(
		ReCompile("Go"),
		StartUrlsChannel(input),
		5,
	)
}
