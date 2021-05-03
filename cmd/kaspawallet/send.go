package main

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	utxopkg "github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func send(conf *sendConfig) error {
	return nil
	//keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	//if err != nil {
	//	return err
	//}
	//
	//toAddress, err := util.DecodeAddress(conf.ToAddress, conf.ActiveNetParams.Prefix)
	//if err != nil {
	//	return err
	//}
	//
	//// TODO: Change index to real number
	//fromAddress, err := libkaspawallet.Address(conf.NetParams(), keysFile.ExtendedPublicKeys, keysFile.MinimumSignatures, 0, keysFile.ECDSA)
	//if err != nil {
	//	return err
	//}
	//
	//client, err := connectToRPC(conf.NetParams(), conf.RPCServer)
	//if err != nil {
	//	return err
	//}
	//utxos, err := fetchSpendableUTXOs(conf.NetParams(), client, fromAddress.String())
	//if err != nil {
	//	return err
	//}
	//
	//sendAmountSompi := uint64(conf.SendAmount * util.SompiPerKaspa)
	//
	//const feePerInput = 1000
	//selectedUTXOs, changeSompi, err := selectUTXOs(utxos, sendAmountSompi, feePerInput)
	//if err != nil {
	//	return err
	//}
	//
	//psTx, err := libkaspawallet.CreateUnsignedTransaction(keysFile.ExtendedPublicKeys,
	//	keysFile.MinimumSignatures,
	//	keysFile.ECDSA,
	//	[]*libkaspawallet.Payment{{
	//		Address: toAddress,
	//		Amount:  sendAmountSompi,
	//	}, {
	//		Address: fromAddress,
	//		Amount:  changeSompi,
	//	}},
	//	selectedUTXOs)
	//if err != nil {
	//	return err
	//}
	//
	//privateKeys, err := keysFile.DecryptMnemonics()
	//if err != nil {
	//	return err
	//}
	//
	//updatedPSTx, err := libkaspawallet.Sign(privateKeys, psTx, keysFile.ECDSA)
	//if err != nil {
	//	return err
	//}
	//
	//tx, err := libkaspawallet.ExtractTransaction(updatedPSTx)
	//if err != nil {
	//	return err
	//}
	//
	//transactionID, err := sendTransaction(client, tx)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Println("Transaction was sent successfully")
	//fmt.Printf("Transaction ID: \t%s\n", transactionID)
	//
	//return nil
}

func fetchSpendableUTXOs(params *dagconfig.Params, client *rpcclient.RPCClient, address string) ([]*appmessage.UTXOsByAddressesEntry, error) {
	getUTXOsByAddressesResponse, err := client.GetUTXOsByAddresses([]string{address})
	if err != nil {
		return nil, err
	}
	virtualSelectedParentBlueScoreResponse, err := client.GetVirtualSelectedParentBlueScore()
	if err != nil {
		return nil, err
	}
	virtualSelectedParentBlueScore := virtualSelectedParentBlueScoreResponse.BlueScore

	spendableUTXOs := make([]*appmessage.UTXOsByAddressesEntry, 0)
	for _, entry := range getUTXOsByAddressesResponse.Entries {
		if !isUTXOSpendable(entry, virtualSelectedParentBlueScore, params.BlockCoinbaseMaturity) {
			continue
		}
		spendableUTXOs = append(spendableUTXOs, entry)
	}
	return spendableUTXOs, nil
}

func selectUTXOs(utxos []*appmessage.UTXOsByAddressesEntry, spendAmount uint64, feePerInput uint64) (
	selectedUTXOs []*libkaspawallet.UTXO, changeSompi uint64, err error) {

	selectedUTXOs = []*libkaspawallet.UTXO{}
	totalValue := uint64(0)

	for _, utxo := range utxos {
		txID, err := transactionid.FromString(utxo.Outpoint.TransactionID)
		if err != nil {
			return nil, 0, err
		}

		rpcUTXOEntry := utxo.UTXOEntry
		scriptPublicKeyScript, err := hex.DecodeString(rpcUTXOEntry.ScriptPublicKey.Script)
		if err != nil {
			return nil, 0, err
		}

		scriptPublicKey := &externalapi.ScriptPublicKey{
			Script:  scriptPublicKeyScript,
			Version: rpcUTXOEntry.ScriptPublicKey.Version,
		}

		utxoEntry := utxopkg.NewUTXOEntry(rpcUTXOEntry.Amount, scriptPublicKey, rpcUTXOEntry.IsCoinbase, rpcUTXOEntry.BlockDAAScore)
		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			OutpointAndUTXOEntryPair: &externalapi.OutpointAndUTXOEntryPair{
				Outpoint: &externalapi.DomainOutpoint{
					TransactionID: *txID,
					Index:         utxo.Outpoint.Index,
				},
				UTXOEntry: utxoEntry,
			},
			DerivationPath: "m/0", // TODO: Change to real path
		})
		totalValue += utxo.UTXOEntry.Amount

		fee := feePerInput * uint64(len(selectedUTXOs))
		totalSpend := spendAmount + fee
		if totalValue >= totalSpend {
			break
		}
	}

	fee := feePerInput * uint64(len(selectedUTXOs))
	totalSpend := spendAmount + fee
	if totalValue < totalSpend {
		return nil, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalSpend)/util.SompiPerKaspa, float64(totalValue)/util.SompiPerKaspa)
	}

	return selectedUTXOs, totalValue - totalSpend, nil
}

func sendTransaction(client *rpcclient.RPCClient, tx *externalapi.DomainTransaction) (string, error) {
	submitTransactionResponse, err := client.SubmitTransaction(appmessage.DomainTransactionToRPCTransaction(tx))
	if err != nil {
		return "", errors.Wrapf(err, "error submitting transaction")
	}
	return submitTransactionResponse.TransactionID, nil
}
