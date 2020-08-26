package main

import (
	"fmt"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error parsing command-line arguments: %s", err))
	}

	client, err := connectToServer(cfg)
	if err != nil {
		printErrorAndExit(fmt.Sprintf("error connecting to the RPC server: %s", err))
	}
	defer client.disconnect()

	requestString := "{\"getCurrentNetworkRequest\": {}}"
	responseString := client.post(requestString)

	fmt.Print(responseString)
}
