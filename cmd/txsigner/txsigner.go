package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/ecc"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"os"
)

func main() {
	cfg, err := parseCommandLine()
	if err != nil {
		printErrorAndExit(err, "Failed to parse arguments")
	}

	privateKey, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		printErrorAndExit(err, "Failed to decode private key")
	}

	transaction, err := parseTransaction(cfg.Transaction)
	if err != nil {
		printErrorAndExit(err, "Failed to decode transaction")
	}

	scriptPubKey, err := createScriptPubKey(privateKey.PubKey())
	if err != nil {
		printErrorAndExit(err, "Failed to create scriptPubKey")
	}

	err = signTransaction(transaction, privateKey, scriptPubKey)
	if err != nil {
		printErrorAndExit(err, "Failed to sign transaction")
	}

	serializedTransaction, err := serializeTransaction(transaction)
	if err != nil {
		printErrorAndExit(err, "Failed to serialize transaction")
	}

	fmt.Printf("Signed Transaction (hex): %s\n\n", serializedTransaction)
}

func parsePrivateKey(privateKeyHex string) (*ecc.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	privateKey, _ := ecc.PrivKeyFromBytes(ecc.S256(), privateKeyBytes)
	return privateKey, err
}

func parseTransaction(transactionHex string) (*wire.MsgTx, error) {
	serializedTx, err := hex.DecodeString(transactionHex)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode transaction hex")
	}
	var transaction wire.MsgTx
	err = transaction.Deserialize(bytes.NewReader(serializedTx))
	return &transaction, err
}

func createScriptPubKey(publicKey *ecc.PublicKey) ([]byte, error) {
	p2pkhAddress, err := util.NewAddressPubKeyHashFromPublicKey(publicKey.SerializeCompressed(), ActiveConfig().NetParams().Prefix)
	if err != nil {
		return nil, err
	}
	scriptPubKey, err := txscript.PayToAddrScript(p2pkhAddress)
	return scriptPubKey, err
}

func signTransaction(transaction *wire.MsgTx, privateKey *ecc.PrivateKey, scriptPubKey []byte) error {
	for i, transactionInput := range transaction.TxIn {
		signatureScript, err := txscript.SignatureScript(transaction, i, scriptPubKey, txscript.SigHashAll, privateKey, true)
		if err != nil {
			return err
		}
		transactionInput.SignatureScript = signatureScript
	}
	return nil
}

func serializeTransaction(transaction *wire.MsgTx) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, transaction.SerializeSize()))
	err := transaction.Serialize(buf)
	serializedTransaction := hex.EncodeToString(buf.Bytes())
	return serializedTransaction, err
}

func printErrorAndExit(err error, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s", message, err)
	os.Exit(1)
}
