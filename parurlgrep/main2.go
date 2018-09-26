package main

import (
	//"flag"
	"fmt"
	"io"
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

///////////
// map-parallel??
func ReCountMatchesURLParallel(re *regexp.Regexp, in <-chan URL,
	max_jobs int /* FIXME: 'jobs' =/= 'workers' */) int {
	// --
	jobsem := make(chan struct{}, max_jobs)

	// --
	for url := range in {
		jobsem <- struct{}{}
		go func(url URL /* ??? jobsem, ??? out-chan */) {
			defer func() { <-jobsem }()
			// TODO: copy regular expr

			if m, ok := ReCountMatchesURL(re, url); ok {
				//total += m
				PrintMatchCountForURL(url, &m)
			} else {
				PrintMatchCountForURL(url, nil)
			}

			/// TODO: cancelation
			// ??? return-channel <- m
		}(url)
	}

	// --
	return 0 // !!! FIXME
}

///////////
func PrintMatchCountForURL(url URL, mcnt *int) {
	if mcnt == nil {
		fmt.Printf("Count for %s: NO DATA\n", url)
		return
	}
	fmt.Printf("Count for %s: %d\n", url, *mcnt)
}

///////////
type InputURLsChan = <-chan URL

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
			PrintMatchCountForURL(url, &m)
		} else {
			PrintMatchCountForURL(url, nil)
		}
	}
	fmt.Printf("Total count: %d", total)
}
