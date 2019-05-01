package main

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/base58"
)

var (
	isRunning       int32
	activeNetParams *dagconfig.Params = &dagconfig.DevNetParams
	p2pkhAddress    util.Address
	privateKey      *btcec.PrivateKey
)

// privateKeyToP2pkhAddress generates p2pkh address from private key.
func privateKeyToP2pkhAddress(key *btcec.PrivateKey, net *dagconfig.Params) (util.Address, error) {
	serializedKey := key.PubKey().SerializeCompressed()
	pubKeyAddr, err := util.NewAddressPubKey(serializedKey, net.Prefix)
	if err != nil {
		return nil, err
	}
	return pubKeyAddr.AddressPubKeyHash(), nil
}

func main() {
	defer handlePanic()

	cfg, err := parseConfig()
	if err != nil {
		panic(fmt.Errorf("Error parsing command-line arguments: %s", err))
	}

	privateKeyBytes := base58.Decode(cfg.PrivateKey)
	privateKey, _ = btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

	p2pkhAddress, err = privateKeyToP2pkhAddress(privateKey, activeNetParams)
	if err != nil {
		panic(fmt.Errorf("Failed to get P2PKH address from private key: %s", err))
	}

	log.Infof("P2PKH address for private key: %s\n", p2pkhAddress)

	addressList, err := getAddressList(cfg)
	if err != nil {
		panic(fmt.Errorf("Couldn't load address list: %s", err))
	}

	clients, err := connectToServers(cfg, addressList)
	if err != nil {
		panic(fmt.Errorf("Error connecting to servers: %s", err))
	}
	defer disconnect(clients)

	atomic.StoreInt32(&isRunning, 1)

	go txLoop(clients)

	interrupt := signal.InterruptListener()
	<-interrupt

	atomic.StoreInt32(&isRunning, 0)
}

func disconnect(clients []*rpcclient.Client) {
	log.Infof("Disconnecting clients")
	for _, client := range clients {
		client.Disconnect()
	}
}

func handlePanic() {
	err := recover()
	if err != nil {
		log.Errorf("Fatal error: %s", err)
		log.Errorf("Stack trace: %s", debug.Stack())
	}
}
