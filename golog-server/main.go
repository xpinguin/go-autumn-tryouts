// golog-cli-server
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

func RunPrologCLI(in interface{ io.Reader }, out interface{ io.Writer }) {
	in_ := bufio.NewReader(in)
	//out := bufio.NewWriter(out)

	m := golog.NewInteractiveMachine()

	for {
		fmt.Fprintf(out, "?- ")
		// --
		var query string
		for {
			s, err := in_.ReadString('\n')
			if err != nil {
				log.Printf("{ERR} ReadString: %v", err)
				break
			}
			// --
			query += strings.TrimSpace(s)
			if query[len(query)-1] == '.' {
				fprintf(out, "<<< [0] '%s'", query)
				break
			}
		}
		// --
		goal, err := gologrdr.Term(query)
		if err != nil {
			fprintf(out, "Problem parsing the query: %s", err)
			continue
		}
		fprintf(out, "<<< [1] '%v'", goal)
		// --
		rs := m.ProveAll(goal)
		fprintf(out, "<<< [RESULT] '%v'", rs)
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
