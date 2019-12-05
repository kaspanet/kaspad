package main

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/jsonrpc"
	"github.com/daglabs/btcd/apiserver/syncd/config"
	"github.com/daglabs/btcd/apiserver/syncd/mqtt"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
	"github.com/pkg/errors"
	"os"
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

	if config.ActiveConfig().Migrate {
		err := database.Migrate(&config.ActiveConfig().ApiServerFlags)
		if err != nil {
			panic(errors.Errorf("Error migrating database: %s", err))
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

	err = mqtt.Connect()
	if err != nil {
		panic(errors.Errorf("Error connecting to MQTT: %s", err))
	}
	defer mqtt.Close()

	err = jsonrpc.Connect(&config.ActiveConfig().ApiServerFlags)
	if err != nil {
		panic(errors.Errorf("Error connecting to servers: %s", err))
	}
	defer jsonrpc.Close()

	doneChan := make(chan struct{}, 1)
	spawn(func() {
		err := startSync(doneChan)
		if err != nil {
			panic(err)
		}
	})

	interrupt := signal.InterruptListener()
	<-interrupt

	// Gracefully stop syncing
	doneChan <- struct{}{}
}
