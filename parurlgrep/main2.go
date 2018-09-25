package main

import (
	//"flag"
	"fmt"
	"io"

	//"log"

	"regexp"
)

///////////
func ReCountMatches(re *regexp.Regexp, in io.RuneReader) (matches int) {
	for re.FindReaderIndex(in) != nil {
		matches++
	}
	return
}

///////////
func ReCountMatchesURL(re *regexp.Regexp, in URL) (int, bool) {
	urlReader := URLGet(in)
	if urlReader == nil {
		return 0, false
	}
	defer urlReader.Close()
	////
	return ReCountMatches(re, urlReader), true
}

///// MAIN /////
func main() {
	// --
	urls := [...]URL{
		"http://golang.com",
		"http://golang.com",
		"http://google.com",
	}
	match_re := regexp.MustCompile("Go")

	// --
	total := 0
	for _, url := range urls {
		if m, ok := ReCountMatchesURL(match_re, url); ok {
			total += m
			fmt.Printf("Count for %s: %d\n", url, m)
		} else {
			fmt.Printf("Count for %s: NO DATA\n", url)
		}
	}
	fmt.Printf("Total count: %d", total)
}
