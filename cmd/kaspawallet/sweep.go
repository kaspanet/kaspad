package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/utils"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/txmass"
	"github.com/pkg/errors"
)

const feePerInput = 10000

func sweep(conf *sweepConfig) error {

	privateKeyBytes, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		return err
	}

	publicKeybytes, err := libkaspawallet.PublicKeyFromPrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}

	addressPubKey, err := util.NewAddressPublicKey(publicKeybytes, conf.NetParams().Prefix)
	if err != nil {
		return err
	}

	address, err := util.DecodeAddress(addressPubKey.String(), conf.NetParams().Prefix)
	if err != nil {
		return err
	}

	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	getExternalSpendableUTXOsResponse, err := daemonClient.GetExternalSpendableUTXOs(ctx, &pb.GetExternalSpendableUTXOsRequest{
		Address: address.String(),
	})
	if err != nil {
		return err
	}

	UTXOs, err := libkaspawallet.KaspawalletdUTXOsTolibkaspawalletUTXOs(getExternalSpendableUTXOsResponse.Entries)
	if err != nil {
		return err
	}

	paymentAmount := uint64(0)

	if len(UTXOs) == 0 {
		return errors.Errorf("Could not find any spendable UTXOs in %s", addressPubKey)
	}

	for _, UTXO := range UTXOs {
		paymentAmount = paymentAmount + UTXO.UTXOEntry.Amount()
	}

	newAddressResponse, err := daemonClient.NewAddress(ctx, &pb.NewAddressRequest{})
	if err != nil {
		return err
	}

	toAddress, err := util.DecodeAddress(newAddressResponse.Address, conf.ActiveNetParams.Prefix)
	if err != nil {
		return err
	}

	splitTransactions, err := createSplitTransactionsWithSchnorrPrivteKey(conf.NetParams(), UTXOs, toAddress, feePerInput)
	if err != nil {
		return err
	}

	serializedSplitTransactions, err := signWithSchnorrPrivateKey(conf.NetParams(), privateKeyBytes, splitTransactions)
	if err != nil {
		return err
	}

	fmt.Println("\nSweeping...")
	fmt.Println("\tFrom:\t", addressPubKey)
	fmt.Println("\tTo:\t", toAddress)

	response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{
		IsDomain:     true,
		Transactions: serializedSplitTransactions,
	})
	if err != nil {
		return err
	}

	totalExtracted := uint64(0)

	fmt.Println("\nTransaction ID(s):")
	for i, txID := range response.TxIDs {
		fmt.Printf("\t%s\n", txID)
		fmt.Println("\tSwept:\t", utils.FormatKas(splitTransactions[i].Outputs[0].Value), " KAS")
		totalExtracted = totalExtracted + splitTransactions[i].Outputs[0].Value
	}

	fmt.Println("\nTotal Funds swept (including transaction fees):")
	fmt.Println("\t", utils.FormatKas(totalExtracted), " KAS")

	return nil
}

func newDummyTransaction() *externalapi.DomainTransaction {
	return &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       make([]*externalapi.DomainTransactionInput, 0), //we create empty inputs
		LockTime:     0,
		Outputs:      make([]*externalapi.DomainTransactionOutput, 1), // we should always have 1 output to the toAdress
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Payload:      nil,
	}
}

func createSplitTransactionsWithSchnorrPrivteKey(
	params *dagconfig.Params,
	selectedUTXOs []*libkaspawallet.UTXO,
	toAddress util.Address,
	feePerInput int) ([]*externalapi.DomainTransaction, error) {

	var splitTransactions []*externalapi.DomainTransaction

	// Add extra mass to transaction, to account for future signatures.
	extraMass := uint64(7000)
	massCalculater := txmass.NewCalculator(params.MassPerTxByte, params.MassPerScriptPubKeyByte, params.MassPerSigOp)

	currentIdxOfSplit := 0

	totalAmount := uint64(0)

	totalSplitAmount := uint64(0)

	scriptPublicKey, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, err
	}

	lastValidTx := newDummyTransaction()
	currentTx := newDummyTransaction() //i.e. the tested tx

	//loop through utxos commit segments that don't violate max mass
	for i, currentUTXO := range selectedUTXOs {

		if currentIdxOfSplit == 0 {
			lastValidTx = newDummyTransaction()
			currentTx = newDummyTransaction()
		}

		currentAmount := currentUTXO.UTXOEntry.Amount()
		totalSplitAmount = totalSplitAmount + currentAmount
		totalAmount = totalAmount + currentAmount

		currentTx.Inputs = append(
			currentTx.Inputs,
			&externalapi.DomainTransactionInput{
				PreviousOutpoint: *currentUTXO.Outpoint,
				UTXOEntry: utxo.NewUTXOEntry(
					currentUTXO.UTXOEntry.Amount(),
					currentUTXO.UTXOEntry.ScriptPublicKey(),
					false,
					constants.UnacceptedDAAScore,
				),
				SigOpCount: 1,
			},
		)

		currentTx.Outputs[0] = &externalapi.DomainTransactionOutput{
			Value:           totalSplitAmount - uint64(len(currentTx.Inputs)*feePerInput),
			ScriptPublicKey: scriptPublicKey,
		}

		if massCalculater.CalculateTransactionMass(currentTx)+extraMass >= mempool.MaximumStandardTransactionMass {

			//in this loop we assume a transaction with one input and one output cannot violate max transaction mass, hence a sanity check.
			if len(currentTx.Inputs) == 1 {
				return nil, errors.Errorf("transaction with one input and one output violates transaction mass")
			}

			splitTransactions = append(splitTransactions, lastValidTx)
			currentIdxOfSplit = 0
			totalSplitAmount = 0
			totalAmount = totalAmount + currentAmount
			lastValidTx = newDummyTransaction()
			currentTx = newDummyTransaction()
			continue
		}

		//Special case, end of inputs, with no violation, where we can assign currentTX to split and break
		if i == len(selectedUTXOs)-1 {
			splitTransactions = append(splitTransactions, currentTx)
			break

		}
		totalAmount = totalAmount + currentAmount
		currentIdxOfSplit++

		lastValidTx = currentTx.Clone()
		currentTx.Outputs = make([]*externalapi.DomainTransactionOutput, 1)

	}
	return splitTransactions, nil
}

func signWithSchnorrPrivateKey(params *dagconfig.Params, privateKeyBytes []byte, domainTransactions []*externalapi.DomainTransaction) ([][]byte, error) {

	schnorrkeyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	serializedDomainTransactions := make([][]byte, len(domainTransactions))

	for i1, domainTransaction := range domainTransactions {

		sighashReusedValues := &consensushashing.SighashReusedValues{}

		for i2, input := range domainTransaction.Inputs {
			signature, err := txscript.SignatureScript(domainTransaction, i2, consensushashing.SigHashAll, schnorrkeyPair, sighashReusedValues)
			if err != nil {
				return nil, err
			}
			input.SignatureScript = signature
		}
		serializedDomainTransactions[i1], err = serialization.SerializeDomainTransaction(domainTransaction)
		if err != nil {
			return nil, err
		}
	}

	return serializedDomainTransactions, nil
}
