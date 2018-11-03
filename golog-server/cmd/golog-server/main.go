package main

import (
	"log"
	"net"
	"os"

	"github.com/mndrix/golog"
	"github.com/xpinguin/go-autumn-tryouts/golog-server"
)

func newPrologEngine() golog.Machine {
	return golog.NewInteractiveMachine()
}

func main() {
	// TODO: use `flags` to select the running mode
	// --
	go server.RunPrologCLI(newPrologEngine(), os.Stdin, os.Stdout)

	// --
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
		go server.RunPrologCLI(newPrologEngine(), c, c)
	}
}
