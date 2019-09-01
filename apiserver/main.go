package main

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/config"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/server"
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	defer panics.HandlePanic(log, logger.BackendLog)

	cfg, err := config.ParseConfig()
	if err != nil {
		panic(fmt.Errorf("Error parsing command-line arguments: %s", err))
	}

	err = database.ConnectToDB(cfg)
	if err != nil {
		panic(fmt.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := database.DB.Close()
		if err != nil {
			panic(fmt.Errorf("Error closing the database: %s", err))
		}
	}()

	client, err := connectToServer(cfg)
	if err != nil {
		panic(fmt.Errorf("Error connecting to servers: %s", err))
	}
	shutdownServer := server.Start(cfg.HTTPListen)
	defer func() {
		shutdownServer()
		disconnectFromNode(client)
	}()

	interrupt := signal.InterruptListener()
	<-interrupt
}

func disconnectFromNode(client *apiServerClient) {
	log.Infof("Disconnecting client")
	client.Disconnect()
}
