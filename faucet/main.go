package main

import (
	"fmt"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/faucet/config"
	"github.com/daglabs/btcd/faucet/database"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/base58"
	"github.com/pkg/errors"
	"os"

	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	faucetAddress      util.Address
	faucetPrivateKey   *btcec.PrivateKey
	faucetScriptPubKey []byte
)

// privateKeyToP2pkhAddress generates p2pkh address from private key.
func privateKeyToP2pkhAddress(key *btcec.PrivateKey, net *dagconfig.Params) (util.Address, error) {
	return util.NewAddressPubKeyHashFromPublicKey(key.PubKey().SerializeCompressed(), net.Prefix)
}

func main() {
	defer panics.HandlePanic(log, logger.BackendLog)

	err := config.Parse()
	if err != nil {
		err := errors.Wrap(err, "Error parsing command-line arguments")
		_, err = fmt.Fprintf(os.Stderr, err.Error())
		if err != nil {
			panic(err)
		}
		return
	}

	cfg, err := config.MainConfig()
	if err != nil {
		panic(err)
	}

	if cfg.Migrate {
		err := database.Migrate()
		if err != nil {
			panic(fmt.Errorf("Error migrating database: %s", err))
		}
		return
	}

	err = database.Connect()
	if err != nil {
		panic(fmt.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := database.Close()
		if err != nil {
			panic(fmt.Errorf("Error closing the database: %s", err))
		}
	}()

	privateKeyBytes := base58.Decode(cfg.PrivateKey)
	faucetPrivateKey, _ = btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

	faucetAddress, err = privateKeyToP2pkhAddress(faucetPrivateKey, config.ActiveNetParams())
	if err != nil {
		panic(fmt.Errorf("Failed to get P2PKH address from private key: %s", err))
	}

	faucetScriptPubKey, err = txscript.PayToAddrScript(faucetAddress)
	if err != nil {
		panic(fmt.Errorf("failed to generate faucetScriptPubKey to address: %s", err))
	}

	shutdownServer := startHTTPServer(cfg.HTTPListen)
	defer shutdownServer()

	interrupt := signal.InterruptListener()
	<-interrupt
}
