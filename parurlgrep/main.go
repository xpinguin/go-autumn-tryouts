package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

///// MAIN /////
func main() {
	// --
	// TODO: use "flag" package, receive from system arguments
	cnt_re := ReCompile("Go")

	// --
	// TODO:
	//  - untangle the ugly mess
	//  - ensure upper bound for the number of simultaneously invoked goroutines
	wait_ch := make(chan struct{}, 2)
	tasks_num := 0

	stopped := false
_MainLoop:
	for {
		select {
		case _, ok := <-wait_ch:
			if ok {
				if tasks_num <= 0 {
					panic(tasks_num)
				}
				tasks_num--
			} else {
				log.Printf("DONE")
				break _MainLoop
			}

		default:
			if stopped {
				if tasks_num == 0 {
					close(wait_ch)
				}
				continue
			}

			var url string

			// -- read url
			n, err := fmt.Scanln(&url)
			if err == io.EOF {
				//break _MainLoop
				//???close(wait_ch)
				stopped = true
				continue
			}
			if n < 1 {
				continue
			} else if n > 1 {
				log.Printf("{WARN} Scanln -> %d, %s\n", n, err)
				continue
			}

			// -- process url
			go func(url string, n chan<- struct{}) {
				url_data := UrlData(url)
				if url_data == nil {
					log.Printf("| %s: NO DATA\n", url)
					return
				}
				log.Printf("| %s: %d\n", url, ReCountMatches(cnt_re, url_data))
				// notify

				n <- struct{}{}
			}(url, wait_ch)
			tasks_num++
		}
	}
}
