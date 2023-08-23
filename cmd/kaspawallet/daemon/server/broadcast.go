package server

import (
	"context"
	"time"

	"github.com/c4ei/yunseokyeol/app/appmessage"
	"github.com/c4ei/yunseokyeol/cmd/kaspawallet/daemon/pb"
	"github.com/c4ei/yunseokyeol/cmd/kaspawallet/libkaspawallet"
	"github.com/c4ei/yunseokyeol/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"
	"github.com/c4ei/yunseokyeol/infrastructure/network/rpcclient"
	"github.com/pkg/errors"
)

func (s *server) Broadcast(_ context.Context, request *pb.BroadcastRequest) (*pb.BroadcastResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

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

		for _, input := range tx.Inputs {
			s.usedOutpoints[input.PreviousOutpoint] = time.Now()
		}
	}

	err = s.refreshUTXOs()
	if err != nil {
		return nil, err
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
