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
	gologterm "github.com/mndrix/golog/term"
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
		ans, err := func() (r interface{}, err error) {
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

			// TODO: return triplet (goal, goal-variables, answer-bindings)
			//		 for the further processing outside of this routine
			bindings := m.ProveAll(goal)
			if len(bindings) == 0 {
				r = "no."
				return
			}
			////
			goalvars := gologterm.Variables(goal)
			if goalvars.Size() == 0 {
				r = "yes."
				return
			}
			////
			for i, b := range bindings {
				lines := make([]string, 0)
				goalvars.ForEach(func(name string, variable interface{}) {
					v := variable.(*gologterm.Variable)
					val := b.Resolve_(v)
					line := fmt.Sprintf("%s = %s", name, val)
					lines = append(lines, line)
				})

				if i == len(bindings)-1 {
					lines[len(lines)-1] += "\t."
				} else {
					lines[len(lines)-1] += "\t;"
				}
				r = strings.Join(lines, "\r\n")
			}
			return
		}()
		// --
		if err != nil {
			continue // TODO: indicate somewhere
		}
		fprintf(out, "%v", ans)
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
