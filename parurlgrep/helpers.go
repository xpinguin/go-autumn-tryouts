package main

import (
	"bufio"
	"io"
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
	URL
	io.ReadCloser
	io.RuneReader
}

func NewURLReader(url URL) *URLReader {
	r, err := http.Get(url)
	if r == nil {
		log.Printf("ERROR: http.Get(%s): %v", url, err)
		return nil
	}
	return &URLReader{url, r.Body, bufio.NewReader(r.Body)}
}
