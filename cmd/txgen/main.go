package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync/atomic"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/base58"
)

var (
	isRunning       int32
	activeNetParams *dagconfig.Params = &dagconfig.DevNetParams
	pkHash          util.Address
	privateKey      *btcec.PrivateKey
)

// keyToAddr maps the passed private to corresponding p2pkh address.
func keyToAddr(key *btcec.PrivateKey, net *dagconfig.Params) (util.Address, error) {
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
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s", err)
		os.Exit(1)
	}

	if cfg.GenerateAddress {
		privateKey, err = btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate private key: %s", err)
			os.Exit(1)
		}
		fmt.Printf("\nPrivate key (base-58): %s\n", base58.Encode(privateKey.Serialize()))
		pkHash, err = keyToAddr(privateKey, activeNetParams)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get pkhash from private key: %s", err)
			os.Exit(1)
		}
		fmt.Printf("Public key hash: %s\n\n", pkHash)
		os.Exit(0)
	}

	privateKeyBytes := base58.Decode(cfg.PrivateKey)
	privateKey, _ = btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

	pkHash, err = keyToAddr(privateKey, activeNetParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get pkhash from private key: %s", err)
		os.Exit(1)
	}

	fmt.Printf("pkhash for private key: %s\n", pkHash)

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

	err = txLoop(clients)
	if err != nil {
		panic(fmt.Errorf("Error in main loop: %s", err))
	}
}

func disconnect(clients []*rpcclient.Client) {
	for _, client := range clients {
		client.Disconnect()
	}
}

func handlePanic() {
	err := recover()
	if err != nil {
		log.Printf("Fatal error: %s", err)
		log.Printf("Stack trace: %s", debug.Stack())
	}
}
