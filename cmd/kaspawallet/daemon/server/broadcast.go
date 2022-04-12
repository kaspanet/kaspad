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
	var domainTransaction *externalapi.DomainTransaction
	var err error

	if !request.IsDomain { 
		domainTransaction, err = serialization.DeserializeDomainTransaction(request.Transaction)
		if err != nil {
			return nil, err
		}
	} else { //default in proto3 is false
		domainTransaction, err = libkaspawallet.DeserializedTransactionFromSerializedPartiallySigned(request.Transaction, s.keysFile.ECDSA)
		if err != nil {
			return nil, err
		}
	}

	txID, err := sendTransaction(s.rpcClient, domainTransaction)
	if err != nil {
		return nil, err
	}

	return &pb.BroadcastResponse{TxID: txID}, nil
}

func sendTransaction(client *rpcclient.RPCClient, tx *externalapi.DomainTransaction) (string, error) {
	submitTransactionResponse, err := client.SubmitTransaction(appmessage.DomainTransactionToRPCTransaction(tx), false)
	if err != nil {
		return "", errors.Wrapf(err, "error submitting transaction")
	}
	return submitTransactionResponse.TransactionID, nil
}
