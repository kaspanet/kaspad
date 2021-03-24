package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

func create(conf *createConfig) error {
	privateKey, address, err := libkaspawallet.CreateKeyPair(conf.NetParams())
	if err != nil {
		return err
	}

	fmt.Println("This is your private key, granting access to all wallet funds. Keep it safe. Use it only when sending Kaspa.")
	fmt.Printf("Private key (hex):\t%x\n\n", privateKey)

	fmt.Println("This is your public address, where money is to be sent.")
	fmt.Printf("Address (%s):\t%s\n", conf.NetParams().Name, address)

	return nil
}
