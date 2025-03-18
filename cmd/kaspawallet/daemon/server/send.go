package server

import (
	"context"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/pkg/errors"
)

func (s *server) Send(_ context.Context, request *pb.SendRequest) (*pb.SendResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	unsignedTransactions, err := s.createUnsignedTransactions(request.ToAddress, request.Amount, request.IsSendAll,
		request.From, request.UseExistingChangeAddress, request.FeePolicy)

	if err != nil {
		return nil, err
	}

	signedTransactions, err := s.signTransactions(unsignedTransactions, request.Password)
	if err != nil {
		return nil, err
	}

	txIDs, err := s.broadcast(signedTransactions, false)
	if err != nil {
		return nil, errors.Wrapf(err, "error broadcasting transactions %s", EncodeTransactionsToHex(signedTransactions))
	}

	return &pb.SendResponse{TxIDs: txIDs, SignedTransactions: signedTransactions}, nil
}
