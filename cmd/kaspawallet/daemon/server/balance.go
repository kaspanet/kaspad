package server

import (
	"context"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
)

func (s *server) GetBalance(_ context.Context, _ *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	virtualSelectedParentBlueScoreResponse, err := s.rpcClient.GetVirtualSelectedParentBlueScore()
	if err != nil {
		return nil, err
	}
	virtualSelectedParentBlueScore := virtualSelectedParentBlueScoreResponse.BlueScore

	var availableBalance, pendingBalance uint64
	for _, entry := range s.utxos {
		if isUTXOSpendable(entry, virtualSelectedParentBlueScore, s.params.BlockCoinbaseMaturity) {
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

func isUTXOSpendable(entry *walletUTXO, virtualSelectedParentBlueScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase() {
		return true
	}
	blockBlueScore := entry.UTXOEntry.BlockDAAScore()
	// TODO: Check for a better alternative than virtualSelectedParentBlueScore
	return blockBlueScore+coinbaseMaturity < virtualSelectedParentBlueScore
}
