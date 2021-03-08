package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/profiling"
)

const timeout = 5 * time.Second

func main() {
	err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %+v", err)
		os.Exit(1)
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())
	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	messagesChan := common.ScanHexFile(cfg.MessagesFilePath)

	err = sendMessages(cfg.NodeP2PAddress, messagesChan)
	if err != nil {
		log.Errorf("Error sending messages: %+v", err)
		backendLog.Close()
		os.Exit(1)
	}
}
