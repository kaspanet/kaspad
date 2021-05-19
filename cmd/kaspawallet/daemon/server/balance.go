package server

import (
	"context"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func (s *server) GetBalance(_ context.Context, _ *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, err
	}

	var availableBalance, pendingBalance uint64
	for _, entry := range s.utxos {
		if isUTXOSpendable(entry, dagInfo.VirtualDAAScore, s.params.BlockCoinbaseMaturity) {
			availableBalance += entry.UTXOEntry.Amount()
		} else {
			pendingBalance += entry.UTXOEntry.Amount()
		}
	}

	return &pb.GetBalanceResponse{
		Available: availableBalance,
		Pending:   pendingBalance,
	}, nil
}

func isUTXOSpendable(entry *walletUTXO, virtualDAAScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase() {
		return true
	}
	return entry.UTXOEntry.BlockDAAScore()+coinbaseMaturity < virtualDAAScore
}
