package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	meta "reflect"
	"regexp"
)

type Re = regexp.Regexp

///////////
type M = struct{}

////
// Re -> Stream rune -> Stream ()
func ReStreamMatchIter(re *Re, in io.RuneReader) <-chan M {
	ms := make(chan M) /* Matches */
	// --
	go func(ms chan<- M) {
		defer close(ms)
		for re.FindReaderIndex(in) != nil {
			ms <- M{} // TODO: use non-empty struct (str or index)
		}
	}(ms)
	// --
	return ms
}

////
// Re -> URL -> Stream ()
func ReURLMatchIter(re *Re, u URL) <-chan struct{} {
	s := NewURLReader(u)
	if s == nil {
		return nil
	}
	// --
	return ReStreamMatchIter(re, s)
}

///////////
type URL_chanM = struct {
	u  URL
	ms <-chan M
}

////
// Re -> Stream URL -> Stream (URL, Stream ())
func ReURLStreamMatchIter(re *Re, us <-chan URL) <-chan URL_chanM {
	ms := make(chan URL_chanM) /* MatcheRs */
	// --
	go func() {
		defer close(ms)
		for u := range us {
			ms <- URL_chanM{u, ReURLMatchIter(re, u)}
		}
	}()
	// --
	return ms
}

////
// Re -> Stream URL -> Stream (URL, Stream ()) -> ()
func ReURLStreamMatchIter_(re *Re, us <-chan URL, ms chan<- URL_chanM) {
	go func() {
		defer close(ms) // NB. auto-close
		for u := range us {
			ms <- URL_chanM{u, ReURLMatchIter(re, u)}
		}
	}()
}

///////////
// Stream rune -> Stream URL
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

///////////
type SelectCase = meta.SelectCase

///
func NewSelectCaseRecv(c interface{}) SelectCase {
	return SelectCase{
		meta.SelectRecv,
		meta.ValueOf(c),
		meta.ValueOf(nil), //meta.Zero(meta.TypeOf(c).Elem()),
	}
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
	// matchers :: Stream (URL, Stream ())
	url_ms := make(chan URL_chanM) //, *max_workers_num)
	ReURLStreamMatchIter_(
		regexp.MustCompile(match_re_src),
		StartUrlsChannel(os.Stdin),
		url_ms)

	// --
	nilcase := NewSelectCaseRecv(nil)
	dcase := NewSelectCaseRecv(url_ms)

	chans := make([]SelectCase, *max_workers_num)
	for i, _ := range chans {
		chans[i] = dcase
	}
	workers_num := len(chans)

	// --
	type URL_ctrM = struct {
		URL_chanM
		ms_ctr int
	}
	urls_ctr := make([]URL_ctrM, len(chans))

	total := 0

	for workers_num > 0 {
		ci, v, copen := meta.Select(chans)
		// --
		switch rv := v.Interface().(type) {
		///////
		case M:
			if !copen {
				total += urls_ctr[ci].ms_ctr
				fmt.Printf("Count for %s: %d\n", urls_ctr[ci].u, urls_ctr[ci].ms_ctr)
				///
				if url_ms != nil {
					chans[ci] = dcase
				} else {
					chans[ci] = nilcase
					workers_num--
				}
				break
			}
			urls_ctr[ci].ms_ctr++
		///////
		case URL_chanM:
			if !copen {
				for i, c := range chans {
					if c == dcase {
						chans[i] = nilcase
						workers_num--
					}
				}
				///
				log.Println("Closed URL matchers channel `URL_chanM`")
				url_ms = nil
				break
			}
			///
			if rv.ms == nil {
				fmt.Printf("Count for %s: NO DATA\n", rv.u)
				break
			}
			///
			chans[ci] = NewSelectCaseRecv(rv.ms)
			urls_ctr[ci] = URL_ctrM{rv, 0}
		}
	}

	// --
	fmt.Printf("Total count: %d", total)
}
