package main

import (
	"github.com/pkg/errors"
)

func main() {
	subCmd, config := parseCommandLine()

	var err error
	switch subCmd {
	case createSubCmd:
		err = create(config.(*createConfig))
	case balanceSubCmd:
		err = balance(config.(*balanceConfig))
	case sendSubCmd:
		err = send(config.(*sendConfig))
	default:
		err = errors.Errorf("Unknown sub-command '%s'\n", subCmd)
	}

	if err != nil {
		printErrorAndExit(err)
	}
}
