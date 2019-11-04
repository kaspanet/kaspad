package main

import (
	"fmt"
	"os"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
)

func new() {
	privateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate private key: %s", err)
		os.Exit(1)
	}

	fmt.Printf("\nPrivate key (hex): %x\n", privateKey.Serialize())

	for _, netParams := range []*dagconfig.Params{&dagconfig.MainNetParams, &dagconfig.TestNetParams, &dagconfig.DevNetParams} {
		addr, err := util.NewAddressPubKeyHashFromPublicKey(privateKey.PubKey().SerializeCompressed(), netParams.Prefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate p2pkh address: %s", err)
			os.Exit(1)
		}
		fmt.Printf("Address (%s): %s\n", netParams.Name, addr)
	}

}
