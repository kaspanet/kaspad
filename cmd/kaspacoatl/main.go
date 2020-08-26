package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"os"
)

func main() {
	defer panics.HandlePanic(log, "MAIN", nil)

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s\n", err)
		os.Exit(1)
	}

	client, err := connectToServer(cfg)
	if err != nil {
		panic(errors.Wrap(err, "error connecting to the RPC server"))
	}
	defer client.disconnect()
}
