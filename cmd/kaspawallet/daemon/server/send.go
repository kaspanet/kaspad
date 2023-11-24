package server

import (
	"context"
	"errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func (s *server) Send(_ context.Context, request *pb.SendRequest) (*pb.SendResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Infof("===wallet server get request: %+v", request)

	unsignedTransactions, err := s.createUnsignedTransactions(request.ToAddress, request.Amount, request.IsSendAll,
		request.From, request.UseExistingChangeAddress, "")

	if err != nil {
		return nil, err
	}

	signedTransactions, err := s.signTransactions(unsignedTransactions, request.Password)
	if err != nil {
		return nil, err
	}

	log.Infof("===wallet server signedTransactions: %+v", signedTransactions)

	return nil, errors.New("test")

	// txIDs, err := s.broadcast(signedTransactions, false)
	// if err != nil {
	// 	return nil, err
	// }

	// return &pb.SendResponse{TxIDs: txIDs, SignedTransactions: signedTransactions}, nil
}
