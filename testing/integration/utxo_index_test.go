package integration

import (
	"encoding/hex"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
)

func TestUTXOIndex(t *testing.T) {
	// Setup a single kaspad instance
	harnessParams := &harnessParams{
		p2pAddress:              p2pAddress1,
		rpcAddress:              rpcAddress1,
		miningAddress:           miningAddress1,
		miningAddressPrivateKey: miningAddress1PrivateKey,
		utxoIndex:               true,
	}
	kaspad, teardown := setupHarness(t, harnessParams)
	defer teardown()

	// skip the first block because it's paying to genesis script,
	// which contains no outputs
	mineNextBlock(t, kaspad)

	// Register for UTXO changes
	const blockAmountToMine = 100
	onUTXOsChangedChan := make(chan *appmessage.UTXOsChangedNotificationMessage, blockAmountToMine)
	err := kaspad.rpcClient.RegisterForUTXOsChangedNotifications([]string{miningAddress1}, func(
		notification *appmessage.UTXOsChangedNotificationMessage) {

		onUTXOsChangedChan <- notification
	})
	if err != nil {
		t.Fatalf("Failed to register for UTXO change notifications: %s", err)
	}

	// Mine some blocks
	for i := 0; i < blockAmountToMine; i++ {
		mineNextBlock(t, kaspad)
	}

	// Collect the UTXO and make sure there's nothing in Removed
	// Note that we expect blockAmountToMine-1 messages because
	// the last block won't be accepted until the next block is
	// mined
	var notificationEntries []*appmessage.UTXOsByAddressesEntry
	for i := 0; i < blockAmountToMine; i++ {
		notification := <-onUTXOsChangedChan
		if len(notification.Removed) > 0 {
			t.Fatalf("Unexpectedly received that a UTXO has been removed")
		}
		for _, added := range notification.Added {
			notificationEntries = append(notificationEntries, added)
		}
	}

	// Submit a few transactions that spends some UTXOs
	const transactionAmountToSpend = 5
	for i := 0; i < transactionAmountToSpend; i++ {
		rpcTransaction := buildTransactionForUTXOIndexTest(t, notificationEntries[i])
		_, err = kaspad.rpcClient.SubmitTransaction(rpcTransaction)
		if err != nil {
			t.Fatalf("Error submitting transaction: %s", err)
		}
	}

	// Mine a block to include the above transactions
	mineNextBlock(t, kaspad)

	// Make sure this block removed the UTXOs we spent
	notification := <-onUTXOsChangedChan
	if len(notification.Removed) != transactionAmountToSpend {
		t.Fatalf("Unexpected amount of removed UTXOs. Want: %d, got: %d",
			transactionAmountToSpend, len(notification.Removed))
	}
	for i := 0; i < transactionAmountToSpend; i++ {
		entry := notificationEntries[i]

		found := false
		for _, removed := range notification.Removed {
			if *removed.Outpoint == *entry.Outpoint {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Missing entry amongst removed UTXOs: %s:%d",
				entry.Outpoint.TransactionID, entry.Outpoint.Index)
		}
	}
	for _, added := range notification.Added {
		notificationEntries = append(notificationEntries, added)
	}

	// Remove the UTXOs we spent from `notificationEntries`
	notificationEntries = notificationEntries[transactionAmountToSpend:]

	// Get all the UTXOs and make sure the response is equivalent
	// to the data collected via notifications
	utxosByAddressesResponse, err := kaspad.rpcClient.GetUTXOsByAddresses([]string{miningAddress1})
	if err != nil {
		t.Fatalf("Failed to get UTXOs: %s", err)
	}
	if len(notificationEntries) != len(utxosByAddressesResponse.Entries) {
		t.Fatalf("Unexpected amount of UTXOs. Want: %d, got: %d",
			len(notificationEntries), len(utxosByAddressesResponse.Entries))
	}
	for _, notificationEntry := range notificationEntries {
		var foundResponseEntry *appmessage.UTXOsByAddressesEntry
		for _, responseEntry := range utxosByAddressesResponse.Entries {
			if *notificationEntry.Outpoint == *responseEntry.Outpoint {
				foundResponseEntry = responseEntry
				break
			}
		}
		if foundResponseEntry == nil {
			t.Fatalf("Missing entry in UTXOs response: %s:%d",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index)
		}
		if notificationEntry.UTXOEntry.Amount != foundResponseEntry.UTXOEntry.Amount {
			t.Fatalf("Unexpected UTXOEntry for outpoint %s:%d. Want: %+v, got: %+v",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index,
				notificationEntry.UTXOEntry, foundResponseEntry.UTXOEntry)
		}
		if notificationEntry.UTXOEntry.BlockBlueScore != foundResponseEntry.UTXOEntry.BlockBlueScore {
			t.Fatalf("Unexpected UTXOEntry for outpoint %s:%d. Want: %+v, got: %+v",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index,
				notificationEntry.UTXOEntry, foundResponseEntry.UTXOEntry)
		}
		if notificationEntry.UTXOEntry.IsCoinbase != foundResponseEntry.UTXOEntry.IsCoinbase {
			t.Fatalf("Unexpected UTXOEntry for outpoint %s:%d. Want: %+v, got: %+v",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index,
				notificationEntry.UTXOEntry, foundResponseEntry.UTXOEntry)
		}
		if *notificationEntry.UTXOEntry.ScriptPublicKey != *foundResponseEntry.UTXOEntry.ScriptPublicKey {
			t.Fatalf("Unexpected UTXOEntry for outpoint %s:%d. Want: %+v, got: %+v",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index,
				notificationEntry.UTXOEntry, foundResponseEntry.UTXOEntry)
		}
	}
}

func buildTransactionForUTXOIndexTest(t *testing.T, entry *appmessage.UTXOsByAddressesEntry) *appmessage.RPCTransaction {
	transactionIDBytes, err := hex.DecodeString(entry.Outpoint.TransactionID)
	if err != nil {
		t.Fatalf("Error decoding transaction ID: %s", err)
	}
	transactionID, err := transactionid.FromBytes(transactionIDBytes)
	if err != nil {
		t.Fatalf("Error decoding transaction ID: %s", err)
	}

	txIns := make([]*appmessage.TxIn, 1)
	txIns[0] = appmessage.NewTxIn(appmessage.NewOutpoint(transactionID, entry.Outpoint.Index), []byte{}, 0)

	payeeAddress, err := util.DecodeAddress(miningAddress1, util.Bech32PrefixKaspaSim)
	if err != nil {
		t.Fatalf("Error decoding payeeAddress: %+v", err)
	}
	toScript, err := txscript.PayToAddrScript(payeeAddress)
	if err != nil {
		t.Fatalf("Error generating script: %+v", err)
	}

	txOuts := []*appmessage.TxOut{appmessage.NewTxOut(entry.UTXOEntry.Amount-1000, toScript)}

	fromScriptCode, err := hex.DecodeString(entry.UTXOEntry.ScriptPublicKey.Script)
	if err != nil {
		t.Fatalf("Error decoding script public key: %s", err)
	}
	fromScript := &externalapi.ScriptPublicKey{Script: fromScriptCode, Version: 0}
	fromAmount := entry.UTXOEntry.Amount

	msgTx := appmessage.NewNativeMsgTx(constants.MaxTransactionVersion, txIns, txOuts)

	privateKeyBytes, err := hex.DecodeString(miningAddress1PrivateKey)
	if err != nil {
		t.Fatalf("Error decoding private key: %+v", err)
	}
	privateKey, err := secp256k1.DeserializePrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		t.Fatalf("Error deserializing private key: %+v", err)
	}

	tx := appmessage.MsgTxToDomainTransaction(msgTx)
	tx.Inputs[0].UTXOEntry = utxo.NewUTXOEntry(fromAmount, fromScript, false, 500)

	signatureScript, err := txscript.SignatureScript(tx, 0, consensushashing.SigHashAll, privateKey,
		&consensushashing.SighashReusedValues{})
	if err != nil {
		t.Fatalf("Error signing transaction: %+v", err)
	}
	msgTx.TxIn[0].SignatureScript = signatureScript

	domainTransaction := appmessage.MsgTxToDomainTransaction(msgTx)
	return appmessage.DomainTransactionToRPCTransaction(domainTransaction)
}
