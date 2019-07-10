package main

import (
	"github.com/daglabs/btcd/signal"
	"log"
	"os"
	"runtime/debug"
)

func main() {
	defer handlePanic()

	cfg, err := parseConfig()
	if err != nil {
		log.Printf("error parsing command-line arguments: %s", err)
		os.Exit(1)
	}

	server, err := newServer(cfg)
	if err != nil {
		log.Panicf("couldn't create server: %s", err)
	}

	defer func() {
		err := server.stop()
		if err != nil {
			log.Panicf("couldn't stop server: %s", err)
		}
	}()

	go func() {
		err = server.start()
		if err != nil {
			log.Panicf("server error: %s", err)
		}
	}()

	interrupt := signal.InterruptListener()
	<-interrupt
}

func handlePanic() {
	err := recover()
	if err != nil {
		log.Printf("Fatal error: %s", err)
		log.Printf("Stack trace: %s", debug.Stack())
	}
}
