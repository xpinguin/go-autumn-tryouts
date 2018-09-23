package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func reCountMatches(re *regexp.Regexp, text []byte) int {
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

///// MAIN /////
func main() {
	// --
	// TODO: use "flag" package, receive from system arguments
	cnt_re, err := regexp.Compile("Go")
	if err != nil || cnt_re == nil {
		panic(err)
	}

	// --
	// TODO:
	//  - untangle ugly mess
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
				log.Printf("| %s: %d\n", url, reCountMatches(cnt_re, url_data))
				// notify

				n <- struct{}{}
			}(url, wait_ch)
			tasks_num++
		}
	}
}
