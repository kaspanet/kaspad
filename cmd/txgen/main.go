package main

import (
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/base58"
	"github.com/daglabs/btcd/util/panics"
	"github.com/pkg/errors"
)

var (
	activeNetParams  = &dagconfig.DevNetParams
	p2pkhAddress     util.Address
	secondaryAddress util.Address
	privateKey       *btcec.PrivateKey
)

// privateKeyToP2pkhAddress generates p2pkh address from private key.
func privateKeyToP2pkhAddress(key *btcec.PrivateKey, net *dagconfig.Params) (util.Address, error) {
	return util.NewAddressPubKeyHashFromPublicKey(key.PubKey().SerializeCompressed(), net.Prefix)
}

func main() {
	defer panics.HandlePanic(log, nil, nil)

	cfg, err := parseConfig()
	if err != nil {
		panic(errors.Errorf("Error parsing command-line arguments: %s", err))
	}

	privateKeyBytes := base58.Decode(cfg.PrivateKey)
	privateKey, _ = btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

	p2pkhAddress, err = privateKeyToP2pkhAddress(privateKey, activeNetParams)
	if err != nil {
		panic(errors.Errorf("Failed to get P2PKH address from private key: %s", err))
	}

	log.Infof("P2PKH address for private key: %s\n", p2pkhAddress)

	if cfg.SecondaryAddress != "" {
		secondaryAddress, err = util.DecodeAddress(cfg.SecondaryAddress, activeNetParams.Prefix)
		if err != nil {
			panic(errors.Errorf("Failed to decode secondary address %s: %s", cfg.SecondaryAddress, err))
		}
	}

	client, err := connectToServer(cfg)
	if err != nil {
		panic(errors.Errorf("Error connecting to servers: %s", err))
	}
	defer disconnect(client)

	spawn(func() {
		err := txLoop(client, cfg)
		if err != nil {
			panic(err)
		}
	})

	interrupt := signal.InterruptListener()
	<-interrupt
}

func disconnect(client *txgenClient) {
	log.Infof("Disconnecting client")
	client.Disconnect()
}
