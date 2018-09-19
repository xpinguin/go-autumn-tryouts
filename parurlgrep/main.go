package main

import (
	//"os"
	"fmt"
	"io"
	"io/ioutil"
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
		_start = _start + _next[1]
		//
		matches_num = matches_num + 1
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
		fmt.Printf("{ERR} ReadAll(R) -> err(%s), R = %s\n", err, r.Body)
		return nil
	}

	return data
}

///// MAIN /////
func main() {
	// --
	E_NODATA = errors.New("NO DATA")
	_reUrlCountMatches := func (re, url) (int, error) {
		url_data := UrlData(url)
		if url_data == nil {
			return nil, E_NODATA
		}
		return reCountMatches(cnt_re, url_data), nil
	}
	
	// --
	// TODO: use "flag" package, receive from system arguments
	cnt_re, _ = regexp.Compile("Go")
	
	// --
	// TODO:
	//	- goroutine + buffered channel
	//	- revert back the nonsensical `_reUrlCountMatches` func,
	//	  to an anonymous lambda binding
	for {
		url := ""
		// --
		n, err := fmt.Scanln(&url)
		if err == io.EOF {
			break
		}
		if n < 1 {
			continue
		} else if n > 1 {
			fmt.Printf("{WARN} Scanln -> %d, %s\n", n, err)
			continue
		}
		// --
		fmt.Printf(">> %s: NO DATA\n", url)
		
		fmt.Printf(">> %s: %d\n", url, )
	}
}
