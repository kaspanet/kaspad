package server

import (
	"context"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/pkg/errors"
)

func (s *server) Broadcast(_ context.Context, request *pb.BroadcastRequest) (*pb.BroadcastResponse, error) {
	txIDs, err := s.broadcast(request.Transactions, request.IsDomain)
	if err != nil {
		return nil, err
	}

	return &pb.BroadcastResponse{TxIDs: txIDs}, nil
}

func (s *server) broadcast(transactions [][]byte, isDomain bool) ([]string, error) {

	txIDs := make([]string, len(transactions))
	var tx *externalapi.DomainTransaction
	var err error

	for i, transaction := range transactions {

		if isDomain {
			tx, err = serialization.DeserializeDomainTransaction(transaction)
			if err != nil {
				return nil, err
			}
		} else if !isDomain { //default in proto3 is false
			tx, err = libkaspawallet.ExtractTransaction(transaction, s.keysFile.ECDSA)
			if err != nil {
				return nil, err
			}
		}

		txIDs[i], err = sendTransaction(s.rpcClient, tx)
		if err != nil {
			return nil, err
		}

		s.tracker.trackTransaction(tx)

		for _, input := range tx.Inputs {
			s.tracker.trackOutpointAsSent(input.PreviousOutpoint.Clone())
			s.tracker.untrackOutpointAsReserved(input.PreviousOutpoint)
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
