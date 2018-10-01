package main

import (
	"bufio"
	//"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

func NewURLReader(uraw URL) (urdr *URLReader) {
	path := uraw
	scheme := ""
	// --
	u, err := url.Parse(uraw)
	if u != nil {
		// HACK: for instance, Win32 drive letter s parsed as a scheme
		if len(u.Scheme) > 1 /* && len(u.Opaque) == 0*/ {
			path = u.Path
			scheme = u.Scheme
		}
	} else {
		log.Printf("{ERR} failed to parse '%s' (net/url): %v", uraw, err, u)
	}
	// --
	switch scheme {
	/// as network URI
	case "http", "https":
		ustr := u.String()
		r, err := http.Get(ustr)
		if r == nil {
			log.Printf("ERROR: http.Get(%s): %v", ustr, err)
			return
		}
		urdr = &URLReader{ustr, r.Body, bufio.NewReader(r.Body)}
		log.Printf(">> Parsed as HTTP URI: %v", u)
	/// as Path
	case "", "file":
		p, err := filepath.Abs(path)
		if err != nil {
			log.Printf("{ERR} failed to parse '%v' as path (%p)", err, p)
			return
		}
		f, err := os.Open(p)
		if f == nil {
			log.Printf("{ERR} unable to open file: %v", err)
			return
		}
		urdr = &URLReader{p, f, bufio.NewReader(f)}
		log.Printf(">> Parsed as FilePath: %v", u)
	/// unknown scheme
	default:
		log.Printf("{ERR} unknown URL(%v) scheme: '%v' ", u, scheme)
	}
	// --
	return
}
