package main

import (
	"bufio"
	//"flag"
	"fmt"
	"io"

	//"io/ioutil"
	"log"
	"net/http"

	//"os"
	"regexp"
)

///////////
func ReCompile(re_src string) *regexp.Regexp {
	re, err := regexp.Compile(re_src)
	if err != nil || re == nil {
		panic(err)
	}
	return re
}

func ReCountMatches(re *regexp.Regexp, in io.RuneReader) (matches int) {
	for re.FindReaderIndex(in) != nil {
		matches++
	}
	return
}

///////////
type URL = string

type URLReader struct {
	url        URL
	bodyReader *bufio.Reader
	_resp      *http.Response
}

func (ur URLReader) Close() { ur._resp.Body.Close() }

// TODO: subclassing: employ interface somehow
func (ur URLReader) Read(p []byte) (int, error) {
	return ur.bodyReader.Read(p)
}
func (ur URLReader) ReadRune() (rune, int, error) {
	return ur.bodyReader.ReadRune()
}

func URLGet(url URL) *URLReader {
	r, err := http.Get(url)
	if r == nil {
		log.Printf("ERROR: http.Get(%s): %v", url, err)
		return nil
	}
	return &URLReader{
		url:        url,
		_resp:      r,
		bodyReader: bufio.NewReader(r.Body),
	}
}

///////////
//func URLDataMatch

///// MAIN /////
func main() {
	urls := [...]URL{
		"http://golang.com",
		"http://golang.com",
		"http://google.com",
	}
	match_re := regexp.MustCompile("Go")

	total := 0
	for _, url := range urls {
		urlReader := URLGet(url)
		if urlReader == nil {
			fmt.Printf("Count for %s: NO DATA", url)
			continue
		}
		defer urlReader.Close()
		///
		m := ReCountMatches(match_re, urlReader)
		total += m
		fmt.Printf("Count for %s: %d", url, m)
	}
}
