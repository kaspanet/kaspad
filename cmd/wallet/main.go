package main

func main() {
	subCommand, config := parseCommandLine()

	switch subCommand {
	case "new":
		new(config.(*newConfig))
	case "balance":
		balance(config.(*balanceConfig))
	case "send":
		send(config.(*sendConfig))
	}
}
