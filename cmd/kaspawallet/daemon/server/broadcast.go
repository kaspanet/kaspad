package server

import (
	"context"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/pkg/errors"
)

func (s *server) Broadcast(_ context.Context, request *pb.BroadcastRequest) (*pb.BroadcastResponse, error) {
	txIDs, err := s.broadcast([][]byte{request.Transaction})
	if err != nil {
		return nil, err
	}

	return &pb.BroadcastResponse{TxID: txIDs[0]}, nil
}

func (s *server) broadcast(transactions [][]byte) ([]string, error) {
	txIDs := make([]string, len(transactions))

	for i, transaction := range transactions {
		tx, err := libkaspawallet.ExtractTransaction(transaction, s.keysFile.ECDSA)
		if err != nil {
			return nil, err
		}

		txIDs[i], err = sendTransaction(s.rpcClient, tx)
		if err != nil {
			return nil, err
		}
	}

	return txIDs, nil
}

func sendTransaction(client *rpcclient.RPCClient, tx *externalapi.DomainTransaction) (string, error) {
	submitTransactionResponse, err := client.SubmitTransaction(appmessage.DomainTransactionToRPCTransaction(tx), false)
	if err != nil {
		return "", errors.Wrapf(err, "error submitting transaction")
	}
	return submitTransactionResponse.TransactionID, nil
}
