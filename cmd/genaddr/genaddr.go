package main

import (
	"fmt"
	"os"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/base58"
)

func main() {
	activeNetParams := &dagconfig.DevNetParams
	privateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate private key: %s", err)
		os.Exit(1)
	}
	fmt.Printf("\nPrivate key (base-58): %s\n", base58.Encode(privateKey.Serialize()))
	serializedKey := privateKey.PubKey().SerializeCompressed()
	pubKeyAddr, err := util.NewAddressPubKey(serializedKey, activeNetParams.Prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate public key address: %s", err)
		os.Exit(1)
	}
	addr := pubKeyAddr.AddressPubKeyHash()
	fmt.Printf("Public key: %s\n\n", addr)
}
