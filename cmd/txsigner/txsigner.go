package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
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

	pubkey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		printErrorAndExit(err, "Failed to generate a public key")
	}
	scriptPubKey, err := createScriptPubKey(pubkey)
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

func parsePrivateKey(privateKeyHex string) (*secp256k1.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, errors.Errorf("'%s' isn't a valid hex. err: '%s' ", privateKeyHex, err)
	}
	return secp256k1.DeserializePrivateKeyFromSlice(privateKeyBytes)
}

func parseTransaction(transactionHex string) (*externalapi.DomainTransaction, error) {
	serializedTx, err := hex.DecodeString(transactionHex)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode transaction hex")
	}
	var transaction appmessage.MsgTx
	err = transaction.Deserialize(bytes.NewReader(serializedTx))
	return appmessage.MsgTxToDomainTransaction(&transaction), err
}

func createScriptPubKey(publicKey *secp256k1.SchnorrPublicKey) ([]byte, error) {
	serializedKey, err := publicKey.SerializeCompressed()
	if err != nil {
		return nil, err
	}
	p2pkhAddress, err := util.NewAddressPubKeyHashFromPublicKey(serializedKey, ActiveConfig().NetParams().Prefix)
	if err != nil {
		return nil, err
	}
	scriptPubKey, err := txscript.PayToAddrScript(p2pkhAddress)
	return scriptPubKey, err
}

func signTransaction(transaction *externalapi.DomainTransaction, privateKey *secp256k1.PrivateKey, scriptPubKey []byte) error {
	for i, transactionInput := range transaction.Inputs {
		signatureScript, err := txscript.SignatureScript(transaction, i, scriptPubKey, txscript.SigHashAll, privateKey, true)
		if err != nil {
			return err
		}
		transactionInput.SignatureScript = signatureScript
	}
	return nil
}

func serializeTransaction(transaction *externalapi.DomainTransaction) (string, error) {
	msgTx := appmessage.DomainTransactionToMsgTx(transaction)
	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	err := msgTx.Serialize(buf)
	serializedTransaction := hex.EncodeToString(buf.Bytes())
	return serializedTransaction, err
}

func printErrorAndExit(err error, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s", message, err)
	os.Exit(1)
}
