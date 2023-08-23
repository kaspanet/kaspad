package main

import (
	"fmt"

	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/libc4exwallet"
	"github.com/c4ei/yunseokyeol/util"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		panic(err)
	}

	privateKey, publicKey, err := libc4exwallet.CreateKeyPair(false)
	if err != nil {
		panic(err)
	}

	addr, err := util.NewAddressPublicKey(publicKey, cfg.NetParams().Prefix)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Private key: %x\n", privateKey)
	fmt.Printf("Address: %s\n", addr)
}
