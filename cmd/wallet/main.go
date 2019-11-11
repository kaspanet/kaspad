package main

import (
	"fmt"
	"os"

	"github.com/daglabs/btcd/dagconfig"
)

var activeNetParams = dagconfig.MainNetParams

func main() {
	subCommand, config := parseCommandLine()

	switch subCommand {
	case "new":
		new(config.(*newConfig))
	case "balance":
		err := balance(config.(*balanceConfig))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
	case "send":
		err := send(config.(*sendConfig))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}

	}
}
