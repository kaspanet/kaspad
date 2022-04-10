package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"

	"github.com/kaspanet/kaspad/util"
)

const feePerInput = 100000

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
	publicKey := hex.EncodeToString(publicKeybytes)

	//NewAddress might seem confusing, but the function is entirely deterministic based on the public key
	//hence, it should provide the same public key as provided in genKeyPair
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

	newAddressResponse, err := daemonClient.NewAddress(ctx, &pb.NewAddressRequest{})
	if err != nil {
		return err
	}

	toAddress, err := util.DecodeAddress(newAddressResponse.Address, conf.ActiveNetParams.Prefix)
	if err != nil {
		return err
	}

	paymentAmount := uint64(0)

	for _, UTXO := range UTXOs {
		paymentAmount = paymentAmount + UTXO.UTXOEntry.Amount()
	}
	fmt.Println("Found ", formatKas(paymentAmount), " Extractable KAS in address")

	payments := make([]*libkaspawallet.Payment, 1)
	payments[0] = &libkaspawallet.Payment{
		Address: toAddress,
		Amount:  paymentAmount,
	}

	partialySignedTransaction, err := libkaspawallet.CreateUnsignedTransactionWithSchnorrPublicKey(UTXOs, publicKey, payments, toAddress, feePerInput)
	if err != nil {
		return err
	}

	splitPartiallySignedTransactions, err := libkaspawallet.CompoundUnsignedTransactionByMaxMassForScnorrPrivateKey(conf.NetParams(), partialySignedTransaction, payments, toAddress, feePerInput)
	if err != nil {
		return err
	}

	TotalSent := uint64(0)
	fmt.Println("Sending to wallet change address:	", toAddress)
	for _, splitPartiallySignedTransaction := range splitPartiallySignedTransactions {
		err := func() error {
			ctx2, cancel2 := context.WithTimeout(context.Background(), daemonTimeout)
			defer cancel2()
			libkaspawallet.SignWithSchnorrPrivteKey(conf.NetParams(), privateKeyBytes, splitPartiallySignedTransaction)

			serializedSplitPartiallySignedTransaction, err := serialization.SerializeDomainTransaction(splitPartiallySignedTransaction.Tx)
			if err != nil {
				fmt.Println(err)
				return err
			}

			broadcastResponse, err := daemonClient.Broadcast(ctx2, &pb.BroadcastRequest{
				Transaction: serializedSplitPartiallySignedTransaction,
			})
			if err != nil {
				fmt.Println(err)
				return err
			}

			fmt.Println("Transaction was sent successfully")
			fmt.Printf("Transaction ID: \t%s\n", broadcastResponse.TxID)
			for _, output := range splitPartiallySignedTransaction.Tx.Outputs {
				TotalSent = TotalSent + output.Value
				fmt.Println("Extracted ", formatKas(TotalSent), " out of ", formatKas(paymentAmount), " KAS")

			}
			return nil

		}()
		if err != nil {
			return err
		}
	}
	fmt.Println("Done")
	return nil

}
