package integration

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	"strings"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
)

func TestTxRelay(t *testing.T) {
	payer, mediator, payee, teardown := standardSetup(t)
	defer teardown()

	// Connect nodes in chain: payer <--> mediator <--> payee
	// So that payee doesn't directly get transactions from payer
	connect(t, payer, mediator)
	connect(t, mediator, payee)

	payeeBlockAddedChan := make(chan *appmessage.RPCBlockHeader)
	setOnBlockAddedHandler(t, payee, func(notification *appmessage.BlockAddedNotificationMessage) {
		payeeBlockAddedChan <- notification.Block.Header
	})
	// skip the first block because it's paying to genesis script
	mineNextBlock(t, payer)
	waitForPayeeToReceiveBlock(t, payeeBlockAddedChan)
	// use the second block to get money to pay with
	secondBlock := mineNextBlock(t, payer)
	waitForPayeeToReceiveBlock(t, payeeBlockAddedChan)

	// Mine BlockCoinbaseMaturity more blocks for our money to mature
	for i := uint64(0); i < payer.config.ActiveNetParams.BlockCoinbaseMaturity; i++ {
		mineNextBlock(t, payer)
		waitForPayeeToReceiveBlock(t, payeeBlockAddedChan)
	}

	// Sleep for `TransactionIDPropagationInterval` to make sure that our transaction will
	// be propagated
	time.Sleep(flowcontext.TransactionIDPropagationInterval)

	msgTx := generateTx(t, secondBlock.Transactions[transactionhelper.CoinbaseTransactionIndex], payer, payee)
	domainTransaction := appmessage.MsgTxToDomainTransaction(msgTx)
	rpcTransaction := appmessage.DomainTransactionToRPCTransaction(domainTransaction)
	response, err := payer.rpcClient.SubmitTransaction(rpcTransaction, false)
	if err != nil {
		t.Fatalf("Error submitting transaction: %+v", err)
	}
	txID := response.TransactionID

	txAddedToMempoolChan := make(chan struct{})

	spawn("TestTxRelay-WaitForTransactionPropagation", func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			_, err := payee.rpcClient.GetMempoolEntry(txID)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					continue
				}

				t.Fatalf("Error getting mempool entry: %+v", err)
			}
			close(txAddedToMempoolChan)
			return
		}
	})

	select {
	case <-txAddedToMempoolChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for transaction to be accepted into mempool")
	}
}

func waitForPayeeToReceiveBlock(t *testing.T, payeeBlockAddedChan chan *appmessage.RPCBlockHeader) {
	select {
	case <-payeeBlockAddedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added")
	}
}

func generateTx(t *testing.T, firstBlockCoinbase *externalapi.DomainTransaction, payer, payee *appHarness) *appmessage.MsgTx {
	txIns := make([]*appmessage.TxIn, 1)
	txIns[0] = appmessage.NewTxIn(appmessage.NewOutpoint(consensushashing.TransactionID(firstBlockCoinbase), 0), []byte{}, 0, 1)

	payeeAddress, err := util.DecodeAddress(payee.miningAddress, util.Bech32PrefixKaspaSim)
	if err != nil {
		t.Fatalf("Error decoding payeeAddress: %+v", err)
	}
	toScript, err := txscript.PayToAddrScript(payeeAddress)
	if err != nil {
		t.Fatalf("Error generating script: %+v", err)
	}

	txOuts := []*appmessage.TxOut{appmessage.NewTxOut(firstBlockCoinbase.Outputs[0].Value-1000, toScript)}

	msgTx := appmessage.NewNativeMsgTx(constants.MaxTransactionVersion, txIns, txOuts)

	privateKeyBytes, err := hex.DecodeString(payer.miningAddressPrivateKey)
	if err != nil {
		t.Fatalf("Error decoding private key: %+v", err)
	}
	privateKey, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		t.Fatalf("Error deserializing private key: %+v", err)
	}

	fromScript := firstBlockCoinbase.Outputs[0].ScriptPublicKey
	fromAmount := firstBlockCoinbase.Outputs[0].Value

	tx := appmessage.MsgTxToDomainTransaction(msgTx)
	tx.Inputs[0].UTXOEntry = utxo.NewUTXOEntry(fromAmount, fromScript, false, 500)
	signatureScript, err := txscript.SignatureScript(tx, 0, consensushashing.SigHashAll, privateKey,
		&consensushashing.SighashReusedValues{})
	if err != nil {
		t.Fatalf("Error signing transaction: %+v", err)
	}
	msgTx.TxIn[0].SignatureScript = signatureScript

	return msgTx
}
