package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
)

func main() {
	subCmd, config := parseCommandLine()

	var err error
	switch subCmd {
	case createSubCmd:
		err = create()
	case balanceSubCmd:
		err = balance(config.(*balanceConfig))
	case sendSubCmd:
		err = send(config.(*sendConfig))
	default:
		err = errors.Errorf("Unknown sub-command '%s'\n", subCmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
