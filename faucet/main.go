package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/ecc"
	"github.com/kaspanet/kaspad/faucet/config"
	"github.com/kaspanet/kaspad/faucet/database"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/base58"
	"github.com/pkg/errors"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	faucetAddress      util.Address
	faucetPrivateKey   *ecc.PrivateKey
	faucetScriptPubKey []byte
)

func main() {
	defer panics.HandlePanic(log, nil, nil)

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
			panic(errors.Errorf("Error migrating database: %s", err))
		}
		return
	}

	err = database.Connect()
	if err != nil {
		panic(errors.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := database.Close()
		if err != nil {
			panic(errors.Errorf("Error closing the database: %s", err))
		}
	}()

	privateKeyBytes := base58.Decode(cfg.PrivateKey)
	faucetPrivateKey, _ = ecc.PrivKeyFromBytes(ecc.S256(), privateKeyBytes)

	faucetAddress, err = privateKeyToP2PKHAddress(faucetPrivateKey, config.ActiveNetParams())
	if err != nil {
		panic(errors.Errorf("Failed to get P2PKH address from private key: %s", err))
	}

	faucetScriptPubKey, err = txscript.PayToAddrScript(faucetAddress)
	if err != nil {
		panic(errors.Errorf("failed to generate faucetScriptPubKey to address: %s", err))
	}

	shutdownServer := startHTTPServer(cfg.HTTPListen)
	defer shutdownServer()

	interrupt := signal.InterruptListener()
	<-interrupt
}

// privateKeyToP2PKHAddress generates p2pkh address from private key.
func privateKeyToP2PKHAddress(key *ecc.PrivateKey, net *dagconfig.Params) (util.Address, error) {
	return util.NewAddressPubKeyHashFromPublicKey(key.PubKey().SerializeCompressed(), net.Prefix)
}
