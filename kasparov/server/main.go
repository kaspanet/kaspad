package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"

	"github.com/kaspanet/kaspad/kasparov/database"
	"github.com/kaspanet/kaspad/kasparov/jsonrpc"
	"github.com/kaspanet/kaspad/kasparov/server/config"
	"github.com/kaspanet/kaspad/kasparov/server/server"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/panics"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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

	err = database.Connect(&config.ActiveConfig().KasparovFlags)
	if err != nil {
		panic(errors.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := database.Close()
		if err != nil {
			panic(errors.Errorf("Error closing the database: %s", err))
		}
	}()

	err = jsonrpc.Connect(&config.ActiveConfig().KasparovFlags)
	if err != nil {
		panic(errors.Errorf("Error connecting to servers: %s", err))
	}
	defer jsonrpc.Close()

	shutdownServer := server.Start(config.ActiveConfig().HTTPListen)
	defer shutdownServer()

	interrupt := signal.InterruptListener()
	<-interrupt
}
