package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

func ReCompile(re_src string) *regexp.Regexp {
	re, err := regexp.Compile(re_src)
	if err != nil || re == nil {
		panic(err)
	}
	return re
}

func ReCountMatches(re *regexp.Regexp, text []byte) int {
	matches_num := 0
	_start := 0
	for {
		_next := re.FindIndex(text[_start:])
		if _next == nil {
			break
		}
		_start += _next[1]
		matches_num++
	}
	return matches_num
}

func UrlData(url string) []byte {
	r, err := http.Get(url)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return nil
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("{ERR} ReadAll(R) -> err(%s), R = %s\n", err, r.Body)
		return nil
	}

	return data
}

func StartUrlsChannel(r io.Reader) <-chan string {
	urls_chan := make(chan string)
	go func(urls_chan chan<- string) {
		var url string
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
	return urls_chan
}

func UrlMatchCountParallel(re *regexp.Regexp,
	urls <-chan string, max_workers int) {
	// --
	type URLMatches struct {
		url          string
		url_has_data bool
		matches_num  int
	}

	urlmatch_chan := make(chan URLMatches, max_workers)

	// --
	urls_chan_closed := false
	tasks_scheduled := 0

	for tasks_scheduled > 0 || !urls_chan_closed {
		//fmt.Println(tasks_scheduled)
		select {
		case m, ok := <-urlmatch_chan:
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

		case url, ok := <-urls:
			//fmt.Printf(">> url (%t): %s\n", ok, url)
			if !ok {
				urls_chan_closed = true
				continue
			}

			tasks_scheduled++
			go func(url string, resch chan<- URLMatches) {
				url_data := UrlData(url)
				if url_data == nil {
					resch <- URLMatches{url, false, 0}
					return
				}
				resch <- URLMatches{url, true, ReCountMatches(re, url_data)}
			}(url, urlmatch_chan)
		}
	}
	close(urlmatch_chan)
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

	//log.Println(match_re_src, *max_workers_num)

	// --
	UrlMatchCountParallel(
		ReCompile(match_re_src),
		StartUrlsChannel(os.Stdin),
		*max_workers_num,
	)

}
