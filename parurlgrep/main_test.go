package main

import (
	"log"
	T "testing"
	"time"
)
import S "strings"
import "runtime"

//import "fmt"

/// TEST DATA ///
var (
	URLs = []string{
		"https://golang.org",
		"http://ya.ru",
		"https://google.com",
		"hjhjhjhdfg",
	}
)

/// TESTS ///
func TestUrlsChannel(t *T.T) {
	input := S.NewReader(S.Join(URLs, "\n"))
	// --
	url_i := 0
	for url := range StartUrlsChannel(input) {
		if URLs[url_i] != url {
			t.Logf("Invalid URL order at %d: (out) '%s' != (in) '%s'",
				url_i, url, URLs[url_i])
			t.Fail()
		}
		url_i++
	}
	if url_i != len(URLs) {
		t.Logf("Inequal number of URLs: (out) %d != (in) %d",
			url_i, len(URLs))
		t.Fail()
	}
}

func TestConcurrentUrlMatcher(t *T.T) {
	max_workers := 3
	input := S.NewReader(S.Repeat(S.Join(URLs, "\n")+"\n", 7))

	UrlMatchCountParallel(
		ReCompile("Go"),
		StartUrlsChannel(input),
		max_workers,
	)
}

func TestMaxWorkersForUrlMatcher(t *T.T) {
	// --
	max_workers := 3
	input := S.NewReader(S.Repeat(S.Join(URLs, "\n")+"\n", 7))

	//total_urls := len(URLs) * max_workers

	// --
	goro_measuremnent_ctl := make(chan struct{})

	// --
	log.Printf("Goros[2]: %d", runtime.NumGoroutine())
	urls := StartUrlsChannel(input)
	log.Printf("Goros[3]: %d", runtime.NumGoroutine())

	// --
	log.Printf("Goros[1]: %d", runtime.NumGoroutine())
	go func(start_stop chan struct{}, interval time.Duration, precall_goros, goros_delta_threshold int) {
		log.Printf("PreCall Goros: %d", precall_goros)

		var goros0, goros int

		measure := func() (goros_delta int) {
			goros = runtime.NumGoroutine()
			goros_delta = goros - goros0
			log.Printf("Goros[...]: %d // goros-goros0 = %d", goros, goros_delta)
			if goros_delta_threshold < goros_delta {
				t.Errorf("goroutines (delta) upper bound violated: %d > %d", goros_delta, goros_delta_threshold)
			}
			return
		}

		<-start_stop
		goros0 = runtime.NumGoroutine()
		log.Printf("goros0 = %d (?== %d)", goros0, runtime.NumGoroutine())
		log.Printf("Goros[START]: %d", runtime.NumGoroutine())
		// guaranteed final one
		defer measure()

		for {
			select {
			case start_stop <- struct{}{}:
				return
			case <-time.After(interval):
				measure()
			}
		}
	}(goro_measuremnent_ctl, 50*time.Millisecond, runtime.NumGoroutine(), max_workers)

	// --
	goro_measuremnent_ctl <- struct{}{}
	total := UrlMatchCountParallel(
		ReCompile("Go"),
		urls,
		max_workers,
	)
	<-goro_measuremnent_ctl

	// --
	log.Printf("Total matches: %d", total)

}
