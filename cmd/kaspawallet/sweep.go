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

	//Sweep assumes the passed private key is a schnorr private key: I see no function to explictly test the type of key
	//It will almost certainly fail somewhere if this is not the case
	privateKeyBytes, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	publicKeybytes, err := libkaspawallet.PublicKeyFromPrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}

	//NewAddress might seem confusing, but the function is entirely deterministic based on the public key
	//hence, it should provide the same public key, as provided in genKeyPair
	addressPubKey, err := util.NewAddressPublicKey(publicKeybytes, conf.NetParams().Prefix)
	if err != nil {
		return err
	}
	fmt.Println("Found associated public address:	", addressPubKey)

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
	fmt.Println("Found ", utils.FormatKas(paymentAmount), " Extractable KAS in address")

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

	serializedSplitTransactions, err := signWithSchnorrPrivteKey(conf.NetParams(), privateKeyBytes, splitTransactions)
	if err != nil {
		return err
	}

	fmt.Println("Sweeping...")
	fmt.Println("	From:	", addressPubKey)
	fmt.Println("	To:	", toAddress)

	response, err := daemonClient.Broadcast(ctx, &pb.BroadcastRequest{
		IsDomain:     true,
		Transactions: serializedSplitTransactions,
	})
	if err != nil {
		return err
	}

	fmt.Println("Transaction ID(s): ")
	for i, txID := range response.TxIDs {
		fmt.Printf("\t%s\n", txID)
		fmt.Println("\tExtracted: ", utils.FormatKas(splitTransactions[i].Outputs[0].Value), " KAS \n")
	}

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

	// I need to add extra mass to transaction, txmasscalculater seems to undercalculate, by around this amount.
	extraMass := uint64(7000)
	massCalculater := txmass.NewCalculator(params.MassPerTxByte, params.MassPerScriptPubKeyByte, params.MassPerSigOp)

	//for refernece this keeps track in respect to dummyWindow[0], not dummyWindow[1]
	currentIdxOfSplit := 0

	totalAmount := uint64(0)

	totalSplitAmount := uint64(0)

	ScriptPublicKey, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, err
	}

	dummyTransactionWindow := make([]*externalapi.DomainTransaction, 2)
	//[0] is the last build that didn't violate mass, that can be added as a split
	dummyTransactionWindow[0] = newDummyTransaction()
	//[1] represents the tested unsigned transaction
	dummyTransactionWindow[1] = newDummyTransaction()

	//loop through utxos commit segments that don't violate max mass
	for i, currentUTXO := range selectedUTXOs {

		if currentIdxOfSplit == 0 {
			dummyTransactionWindow[0] = newDummyTransaction()
			dummyTransactionWindow[1] = newDummyTransaction()
			//we assume a transaction without inputs, and 1 undefined output cannot violate transaction mass
		}

		currentAmount := currentUTXO.UTXOEntry.Amount()
		totalSplitAmount = totalSplitAmount + currentAmount
		totalAmount = totalAmount + currentAmount

		dummyTransactionWindow[1].Inputs = append(
			dummyTransactionWindow[1].Inputs,
			&externalapi.DomainTransactionInput{
				PreviousOutpoint: *currentUTXO.Outpoint,
				UTXOEntry:        utxo.NewUTXOEntry(
					currentUTXO.UTXOEntry.Amount(),
					currentUTXO.UTXOEntry.ScriptPublicKey(),
					false,
					constants.UnacceptedDAAScore,
				),
				SigOpCount:       1,
			},
		)

		dummyTransactionWindow[1].Outputs[0] = &externalapi.DomainTransactionOutput{
			Value:           totalSplitAmount - uint64(len(dummyTransactionWindow[1].Inputs)*feePerInput),
			ScriptPublicKey: ScriptPublicKey,
		}

		if massCalculater.CalculateTransactionMass(dummyTransactionWindow[1])+extraMass >= mempool.MaximumStandardTransactionMass {
			splitTransactions = append(splitTransactions, dummyTransactionWindow[0])
			currentIdxOfSplit = 0
			totalSplitAmount = 0
			totalAmount = totalAmount + currentAmount
			dummyTransactionWindow[0] = newDummyTransaction()
			dummyTransactionWindow[1] = newDummyTransaction()
			if i == len(selectedUTXOs)-1 {
				splitTransactions = append(splitTransactions, dummyTransactionWindow[1])
				break
			}
			continue
		}

		//Special case, end of inputs, with no violation, where we can assign dummyWindow[1] to split and break
		if i == len(selectedUTXOs)-1 {
			splitTransactions = append(splitTransactions, dummyTransactionWindow[1])
			break

		}
		totalAmount = totalAmount + currentAmount
		currentIdxOfSplit++

		dummyTransactionWindow[0] = dummyTransactionWindow[1].Clone()
		dummyTransactionWindow[1].Outputs = make([]*externalapi.DomainTransactionOutput, 1)

	}
	return splitTransactions, nil
}

func signWithSchnorrPrivteKey(params *dagconfig.Params, privateKeyBytes []byte, domainTransactions []*externalapi.DomainTransaction) ([][]byte, error) {

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
