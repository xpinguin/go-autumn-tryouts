package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	meta "reflect"
	"regexp"
	"runtime"
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
		///
		for re.FindReaderIndex(in) != nil {
			ms <- M{} // TODO: use non-empty struct (str or index)
		}
	}(ms)
	// --
	return ms
}

// Re -> URL -> Maybe Stream ()
func ReURLMatchIter(re *Re, u URL) <-chan M {
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

// Re -> Stream URL -> Stream (URL, Maybe Stream ()) -> ()
func ReURLStreamMatchIter_(re *Re, us <-chan URL, ms chan<- URL_chanM) {
	go func() {
		defer close(ms)
		///
		for u := range us {
			ms <- URL_chanM{u, ReURLMatchIter(re, u)}
		}
	}()
}

// Re -> Stream URL -> Stream (URL, Maybe Stream ())
func ReURLStreamMatchIter(re *Re, us <-chan URL) <-chan URL_chanM {
	ms := make(chan URL_chanM) /* MatcheRs */
	ReURLStreamMatchIter_(re, us, ms)
	return ms
}

///////////
type URL_ctrM = struct {
	u    URL
	ctrs <-chan int
}

// Re -> Stream URL -> Stream (URL, Maybe Stream Int) -> ()
func ReURLStreamMatchCounter_(re *Re, us <-chan URL, ms chan<- URL_ctrM) {
	go func() {
		defer close(ms)
		///
		for u := range us {
			ms <- URL_ctrM{u, MatchesCounter(ReURLMatchIter(re, u))}
		}
	}()
}

// Re -> Stream URL -> Stream (URL, Maybe Stream Int)
func ReURLStreamMatchCounter(re *Re, us <-chan URL) chan<- URL_ctrM {
	ms := make(chan URL_ctrM) /* MatcheRs */
	ReURLStreamMatchCounter_(re, us, ms)
	return ms
}

///////////
// Stream rune -> Stream URL
func UrlsIter_(r io.Reader, urls chan<- URL) {
	go func() {
		defer close(urls)
		///
		var u URL
		for {
			// -- read url
			n, err := fmt.Fscanln(r, &u)
			if err == io.EOF {
				return
			}
			if n < 1 {
				continue
			} else if n > 1 {
				log.Printf("{WARN} Scanln -> %d, %s\n", n, err)
				continue
			}
			// --
			urls <- u
		}
	}()
}

func UrlsIter(r io.Reader) <-chan URL {
	urls_chan := make(chan URL)
	UrlsIter_(r, urls_chan)
	return urls_chan
}

///////////
// Stream () -> Int
func MatchesCount(ms <-chan M) int {
	ctr := 0
	for range ms {
		ctr++
	}
	return ctr
}

// Maybe Stream () -> Stream Int -> Bool
func MatchesCounter_(ms <-chan M, ctr_out chan<- int) bool {
	if ms == nil {
		return false
	}
	// --
	go func() {
		defer close(ctr_out)
		///
		ctr_out <- MatchesCount(ms)
	}()
	return true
}

// Maybe Stream () -> Maybe Stream Int
func MatchesCounter(ms <-chan M) <-chan int {
	if ms == nil {
		return nil
	}
	// --
	if ctr_out := make(chan int); MatchesCounter_(ms, ctr_out) {
		return ctr_out
	}
	return nil
}

// Maybe Stream () -> Stream Maybe Int -> ()
func MatchesCounterIter_(ms <-chan M, ctr_out chan<- *int) {
	go func() {
		defer close(ctr_out)
		///
		if ms == nil {
			ctr_out <- nil
			return
		}
		ctr := MatchesCount(ms)
		ctr_out <- &ctr
	}()
}

// Maybe Stream () -> Stream Maybe Int
func MatchesCounterIter(ms <-chan M) <-chan *int {
	ctr_out := make(chan *int)
	MatchesCounterIter_(ms, ctr_out)
	return ctr_out
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
	gomaxprocs := runtime.GOMAXPROCS(runtime.NumCPU())
	log.Printf("[set] GOMAXPROCS = %d", gomaxprocs)
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
	url_ms := make(chan URL_chanM, *max_workers_num) // :: Stream (URL, Maybe Stream ())
	ReURLStreamMatchIter_(
		regexp.MustCompile(match_re_src),
		UrlsIter(os.Stdin),
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
			ms := MatchesCounter(rv.ms)
			//ms := MatchesCounterIter(rv.ms)
			//ms := rv.ms
			if ms == nil {
				fmt.Printf("[URL_chanM] Count for %s: NO DATA\n", rv.u)
				break
			}
			///
			chans[ci] = NewSelectCaseRecv(ms)
			urls_ctr[ci] = URL_ctrM{rv, 0}
		///////
		case *int:
			if !copen {
				if url_ms != nil {
					chans[ci] = dcase
				} else {
					chans[ci] = nilcase
					workers_num--
				}
				break
			}
			///
			if rv == nil {
				fmt.Printf("[*int] Count for %s: NO DATA\n", urls_ctr[ci].u)
				break
			}
			fmt.Printf("[*int] Count for %s: %d\n", urls_ctr[ci].u, *rv)
			total += *rv
		///////
		case int:
			if !copen {
				if url_ms != nil {
					chans[ci] = dcase
				} else {
					chans[ci] = nilcase
					workers_num--
				}
				break
			}
			///
			fmt.Printf("[int] Count for %s: %d\n", urls_ctr[ci].u, rv)
			total += rv
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
		}

	}

	// --
	fmt.Printf("Total count: %d", total)
}
