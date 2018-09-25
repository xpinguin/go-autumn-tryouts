package main

import (
	"bufio"
	"log"
	"net/http"
)

///////////
func Min(x0 int, xs ...int) int {
	r := x0
	for _, x := range xs {
		if x < r {
			r = x
		}
	}
	return r
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
