package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
	"os"
)

func parsePrivateKey(privateKeyHex string) (*btcec.PrivateKey, *btcec.PublicKey) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	exitIfError(err, "Failed to decode private key")
	privateKey, publicKey := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
	return privateKey, publicKey
}

func parseTransaction(hexString string) *wire.MsgTx {
	serializedTx, err := hex.DecodeString(hexString)
	exitIfError(err, "Failed to decode transaction")

	var transaction wire.MsgTx
	err = transaction.Deserialize(bytes.NewReader(serializedTx))
	exitIfError(err, "Failed to decode transaction")

	return &transaction
}

func createPayToAddressScript(publicKey *btcec.PublicKey) []byte {
	activeNetParams := &dagconfig.DevNetParams
	p2pkhAddress, err := util.NewAddressPubKeyHashFromPublicKey(publicKey.SerializeCompressed(), activeNetParams.Prefix)
	payToAddrScript, err := txscript.PayToAddrScript(p2pkhAddress)
	exitIfError(err, "Failed to create pay-to-address-script")
	return payToAddrScript
}

func signTransaction(transaction *wire.MsgTx, privateKey *btcec.PrivateKey, payToAddressScript []byte) {
	for i, transactionInput := range transaction.TxIn {
		signatureScript, err := txscript.SignatureScript(transaction, i, payToAddressScript, txscript.SigHashAll, privateKey, true)
		exitIfError(err, "Failed to sign transaction")
		transactionInput.SignatureScript = signatureScript
	}
}

func serializeTransaction(transaction *wire.MsgTx) string {
	buf := bytes.NewBuffer(make([]byte, 0, transaction.SerializeSize()))
	err := transaction.Serialize(buf)
	serializedTransaction := hex.EncodeToString(buf.Bytes())
	exitIfError(err, "Failed to serialize transaction")

	return serializedTransaction
}

func exitIfError(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", message, err)
		os.Exit(1)
	}
}

func main() {
	cfg := parseCommandLine()
	privateKey, publicKey := parsePrivateKey(cfg.PrivateKey)
	transaction := parseTransaction(cfg.Transaction)
	payToAddressScript := createPayToAddressScript(publicKey)
	signTransaction(transaction, privateKey, payToAddressScript)
	fmt.Printf("Signed Transaction (hex): %s\n\n", serializeTransaction(transaction))
}
