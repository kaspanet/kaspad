package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"

	"github.com/daglabs/btcd/kasparov/database"
	"github.com/daglabs/btcd/kasparov/jsonrpc"
	"github.com/daglabs/btcd/kasparov/server/config"
	"github.com/daglabs/btcd/kasparov/server/server"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
)

func main() {
	defer panics.HandlePanic(log, nil, nil)

	err := config.Parse()
	if err != nil {
		errString := fmt.Sprintf("Error parsing command-line arguments: %s", err)
		_, fErr := fmt.Fprintf(os.Stderr, errString)
		if fErr != nil {
			panic(errString)
		}
		return
	}

	err = database.Connect(&config.ActiveConfig().ApiServerFlags)
	if err != nil {
		panic(errors.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := database.Close()
		if err != nil {
			panic(errors.Errorf("Error closing the database: %s", err))
		}
	}()

	err = jsonrpc.Connect(&config.ActiveConfig().ApiServerFlags)
	if err != nil {
		panic(errors.Errorf("Error connecting to servers: %s", err))
	}
	defer jsonrpc.Close()

	shutdownServer := server.Start(config.ActiveConfig().HTTPListen)
	defer shutdownServer()

	interrupt := signal.InterruptListener()
	<-interrupt
}
