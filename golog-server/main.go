// golog-server
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/mndrix/golog"
	gologrdr "github.com/mndrix/golog/read"
)

func fprintf(f interface{ io.Writer }, s string, opts ...interface{}) {
	fmt.Fprintf(f, s+"\r\n", opts...)
}

func RunPrologCLI(in_ interface{ io.Reader }, out interface{ io.Writer }) {
	in := bufio.NewReader(in_)
	m := golog.NewInteractiveMachine()

	for inClosed := false; !inClosed; {
		fmt.Fprintf(out, "?- ")
		// --
		rs, err := func() (r interface{}, err error) {
			// read logical line
			var query, s string
			for {
				s, err = in.ReadString('\n')
				if err == io.EOF {
					log.Printf("Input stream closed. Stopping CLI")
					inClosed = true
					return
				} else if err != nil {
					log.Printf("{ERR} ReadString: %v", err)
					return
				}

				query += strings.TrimSpace(s)
				if len(query) == 0 || (len(query) > 0 && query[len(query)-1] == '.') {
					fprintf(out, "<<< [0] '%s'", query)
					break
				}
			}

			// parse line & feed into the Prolog engine
			defer func() {
				if exc := recover(); exc != nil {
					r = exc
				}
			}()

			goal, err := gologrdr.Term(query)
			if err != nil {
				fprintf(out, "Problem parsing the query: %s", err)
				return
			}

			fprintf(out, "<<< [1] '%v'", goal)
			r = m.ProveAll(goal)
			return
		}()
		// --
		if err == nil {
			fprintf(out, "[RESULT] '%v'", rs)
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4567")
	if err != nil {
		log.Fatalln("{ERR} net.Listen(...):", err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("{ERR} Accept():", err)
			continue
		}
		// --
		log.Printf("Starting CLI for client: %v", c.RemoteAddr())
		go RunPrologCLI(c, c)
	}
}
