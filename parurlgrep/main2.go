package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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

///////////
// StartUrlsChannel :: IO () -> Chan URL
func StartUrlsChannel(r io.Reader) <-chan URL {
	urls_chan := make(chan URL)
	///
	go func(urls_chan chan<- URL) { // _ :: Chan URL -> IO ()
		var url URL
		for {
			// read url
			n, err := fmt.Fscanln(r, &url)
			if err == io.EOF {
				close(urls_chan)
				return
			}
			if n < 1 {
				continue
			} else if n > 1 {
				log.Printf("{WARN} Scanln -> %d, %s\n", n, err)
				continue
			}
			//
			urls_chan <- url
		}
	}(urls_chan)
	///
	return urls_chan
}

///// MAIN /////
func main() {
	// --
	var match_re_src string
	default_match_re_src := flag.String("", "Go", "pattern to match (re2)")
	max_workers_num := flag.Int("k", 5, "maximum number of workers")

	flag.Parse()
	if flag.NArg() > 1 {
		flag.Usage()
		log.Fatal(flag.Args())
	} else if match_re_src = flag.Arg(0); match_re_src == "" {
		match_re_src = *default_match_re_src
	}
	/////
	log.Println(match_re_src, *max_workers_num)

	// --
	match_re := regexp.MustCompile(match_re_src)

	total := ReCountMatchesURLParallel(match_re,
		StartUrlsChannel(os.Stdin),
		*max_workers_num)
	/*
		total := 0
		for url := range StartUrlsChannel(os.Stdin) {
			if m, ok := ReCountMatchesURL(match_re, url); ok {
				total += m
				PrintMatchCountForURL(url, &m)
			} else {
				PrintMatchCountForURL(url, nil)
			}
		}*/
	fmt.Printf("Total count: %d", total)
}
