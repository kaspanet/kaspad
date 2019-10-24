package main

import (
	"fmt"
	"os"
	"encoding/hex"

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
	wif, err := util.NewWIF(privateKey, activeNetParams.PrivateKeyID, true)
	if err != nil {
		panic(fmt.Sprintf("error generating wif: %s", err))
	}
	fmt.Printf("\nPrivate key wif: %s\n", wif)
	addr, err := util.NewAddressPubKeyHashFromPublicKey(privateKey.PubKey().SerializeCompressed(), activeNetParams.Prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate p2pkh address: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Address: %s\n", addr)
	hash160 := addr.Hash160()[:]
	fmt.Printf("Hash160 of address (hex): %s\n\n", hex.EncodeToString(hash160))
}
