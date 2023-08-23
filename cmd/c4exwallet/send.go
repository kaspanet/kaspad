package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/daemon/client"
	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/daemon/pb"
	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/keys"
	"github.com/c4ei/yunseokyeol/cmd/c4exwallet/libc4exwallet"
	"github.com/c4ei/yunseokyeol/domain/consensus/utils/constants"
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

	var sendAmountSompi uint64
	if !conf.IsSendAll {
		sendAmountSompi = uint64(conf.SendAmount * constants.SompiPerKaspa)
	}

	createUnsignedTransactionsResponse, err :=
		daemonClient.CreateUnsignedTransactions(ctx, &pb.CreateUnsignedTransactionsRequest{
			From:                     conf.FromAddresses,
			Address:                  conf.ToAddress,
			Amount:                   sendAmountSompi,
			IsSendAll:                conf.IsSendAll,
			UseExistingChangeAddress: conf.UseExistingChangeAddress,
		})
	if err != nil {
		return err
	}

	if len(conf.Password) == 0 {
		conf.Password = keys.GetPassword("Password:")
	}
	mnemonics, err := keysFile.DecryptMnemonics(conf.Password)
	if err != nil {
		if strings.Contains(err.Error(), "message authentication failed") {
			fmt.Fprintf(os.Stderr, "Password decryption failed. Sometimes this is a result of not "+
				"specifying the same keys file used by the wallet daemon process.\n")
		}
		return err
	}

	signedTransactions := make([][]byte, len(createUnsignedTransactionsResponse.UnsignedTransactions))
	for i, unsignedTransaction := range createUnsignedTransactionsResponse.UnsignedTransactions {
		signedTransaction, err := libc4exwallet.Sign(conf.NetParams(), mnemonics, unsignedTransaction, keysFile.ECDSA)
		if err != nil {
			return err
		}
		signedTransactions[i] = signedTransaction
	}

	if len(signedTransactions) > 1 {
		fmt.Printf("Broadcasting %d transactions\n", len(signedTransactions))
	}

	// Since we waited for user input when getting the password, which could take unbound amount of time -
	// create a new context for broadcast, to reset the timeout.
	broadcastCtx, broadcastCancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer broadcastCancel()

	response, err := daemonClient.Broadcast(broadcastCtx, &pb.BroadcastRequest{Transactions: signedTransactions})
	if err != nil {
		return err
	}
	fmt.Println("Transactions were sent successfully")
	fmt.Println("Transaction ID(s): ")
	for _, txID := range response.TxIDs {
		fmt.Printf("\t%s\n", txID)
	}

	if conf.Verbose {
		fmt.Println("Serialized Transaction(s) (can be parsed via the `parse` command or resent via `broadcast`): ")
		for _, signedTx := range signedTransactions {
			fmt.Printf("\t%x\n\n", signedTx)
		}
	}

	return nil
}
