// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/kaspanet/kaspad/infrastructure/network/rpc/client"
)

func main() {
	// Connect to a local kaspa RPC server using HTTP POST mode.
	connCfg := &client.ConnConfig{
		Host:         "localhost:8332",
		User:         "yourrpcuser",
		Pass:         "yourrpcpass",
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := client.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)
}
