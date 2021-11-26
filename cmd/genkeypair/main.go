package main

import (
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		panic(err)
	}

	privateKey, publicKey, err := libkaspawallet.CreateKeyPair(false)
	if err != nil {
		panic(err)
	}

	addr, err := util.NewAddressPublicKey(publicKey, cfg.NetParams().Prefix)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Private key: %x\n", privateKey)
	fmt.Printf("Public key: %x\n", publicKey)
	fmt.Printf("Address: %s\n", addr)
}

/*
	******* EXAMPLES:
	Private key: 498d5c65d3c24135ec8cf1cd8839cd7a08ac682655da0e06433f18078279abd6
	Public key: 06631cddff32f52cbca9606360e44fa6fd49f5c9e158cf384ae252c6f7934a3d
	Address: kaspa:qqrxx8xalue02t9u49sxxc8yf7n06j04e8s43necft3993hhjd9r62w0q76wm
	h := string("PRIV_KEY_HEX")
	pk1, _ := hex.DecodeString(h)
	pk := []byte(pk1)
	deserPrivKey, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(pk)
	if err != nil {
		panic(err)
	}
	strPrivKey := deserPrivKey.SerializePrivateKey().String()
	fmt.Printf("Str Private key: %s\n", strPrivKey)
	pubKey, _ := deserPrivKey.SchnorrPublicKey()
	strPubKey, _ := pubKey.Serialize()
	///pubKey.String()
	addr2, _ := util.NewAddressPublicKey(strPubKey[:], cfg.NetParams().Prefix)
	fmt.Printf("Str Public key ( serialized + string ): %s\n", strPubKey.String())
	fmt.Printf("Str Public key (only string): %s\n", pubKey.String())
	fmt.Printf("Address2: %s\n", addr2)

	pubKeyHex := string("")
	pubKey3, _ := hex.DecodeString(pubKeyHex) // 32 bytes
	addr3, _ := util.NewAddressPublicKey(pubKey3, cfg.NetParams().Prefix)

	fmt.Printf("Address3: %s\n", addr3)
*/
