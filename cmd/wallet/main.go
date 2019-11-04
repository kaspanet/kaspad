package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	subCommandFuncs = map[string]func(){
		"new":     new,
		"balance": balance,
		"send":    send,
	}
	appName = filepath.Base(os.Args[0])
)

func printSubCommands() {
	subCommands := make([]string, 0, len(subCommandFuncs))
	for subCommand := range subCommandFuncs {
		subCommands = append(subCommands, subCommand)
	}

	fmt.Fprintf(os.Stderr, "Available sub-commands: %v\n", strings.Join(subCommands, ", "))
	fmt.Fprintf(os.Stderr, "Use `%s [sub-command] --help` to get usage instructions for a sub-command\n", appName)

	os.Exit(1)
}

func parseSubCommand() func() {
	if len(os.Args) < 2 {
		printSubCommands()
		return nil
	}

	var subCommandFunc func()
	var ok bool
	if subCommandFunc, ok = subCommandFuncs[os.Args[1]]; !ok {
		printSubCommands()
		return nil
	}

	return subCommandFunc
}

func main() {
	subCommandFunc := parseSubCommand()
	subCommandFunc()
}
