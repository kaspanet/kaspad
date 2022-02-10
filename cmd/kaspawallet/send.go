package main

import (
	"context"
	"fmt"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/client"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/keys"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/pkg/errors"
)

func send(conf *sendConfig) error {
	keysFile, err := keys.ReadKeysFile(conf.NetParams(), conf.KeysFile)
	if err != nil {
		return err
	}

	if len(keysFile.ExtendedPublicKeys) > len(keysFile.EncryptedMnemonics) {
		return errors.Errorf("Cannot use 'send' command for multisig wallet without all of the keys")
	}

	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	sendAmountSompi := uint64(conf.SendAmount * constants.SompiPerKaspa)
	createUnsignedTransactionsResponse, err :=
		daemonClient.CreateUnsignedTransactions(ctx, &pb.CreateUnsignedTransactionsRequest{
			Address: conf.ToAddress,
			Amount:  sendAmountSompi,
		})
	if err != nil {
		return err
	}

	mnemonics, err := keysFile.DecryptMnemonics(conf.Password)
	if err != nil {
		return err
	}

	signedTransactions := make([][]byte, len(createUnsignedTransactionsResponse.UnsignedTransactions))
	for i, unsignedTransaction := range createUnsignedTransactionsResponse.UnsignedTransactions {
		signedTransaction, err := libkaspawallet.Sign(conf.NetParams(), mnemonics, unsignedTransaction, keysFile.ECDSA)
		if err != nil {
			return err
		}
		signedTransactions[i] = signedTransaction
	}

	if len(signedTransactions) > 1 {
		fmt.Printf("Broadcasting %d transactions", len(signedTransactions))
	}
	for _, signedTransaction := range signedTransactions {
		return func() error { // surround with func so that defer runs separately per transaction
			ctx2, cancel2 := context.WithTimeout(context.Background(), daemonTimeout)
			defer cancel2()
			broadcastResponse, err := daemonClient.Broadcast(ctx2, &pb.BroadcastRequest{
				Transaction: signedTransaction,
			})
			if err != nil {
				return err
			}

			fmt.Println("Transaction was sent successfully")
			fmt.Printf("Transaction ID: \t%s\n", broadcastResponse.TxID)
			return nil
		}()
	}

	return nil
}
