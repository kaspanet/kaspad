package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
)

type config struct {
	Transaction string `long:"tx" description:"Unsigned transaction in HEX format" required:"true"`
	PrivateKey  string `long:"pk" description:"Private key" required:"true"`
}

func parseCommandLine() *config {
	cfg := &config{}
	parser := flags.NewParser(cfg, flags.PrintErrors|flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse arguments: %s", err)
		os.Exit(1)
	}

	return cfg
}
