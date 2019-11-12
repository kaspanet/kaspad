package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
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

func parsePrivateKey(privateKeyHex string) (*btcec.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
	return privateKey, err
}

func parseTransaction(transactionHex string) (*wire.MsgTx, error) {
	serializedTx, err := hex.DecodeString(transactionHex)
	var transaction wire.MsgTx
	err = transaction.Deserialize(bytes.NewReader(serializedTx))
	return &transaction, err
}

func createScriptPubKey(publicKey *btcec.PublicKey) ([]byte, error) {
	p2pkhAddress, err := util.NewAddressPubKeyHashFromPublicKey(publicKey.SerializeCompressed(), ActiveConfig().NetParams().Prefix)
	scriptPubKey, err := txscript.PayToAddrScript(p2pkhAddress)
	return scriptPubKey, err
}

func signTransaction(transaction *wire.MsgTx, privateKey *btcec.PrivateKey, scriptPubKey []byte) error {
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
