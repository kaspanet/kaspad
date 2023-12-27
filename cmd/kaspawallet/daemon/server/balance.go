package server

import (
	"context"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

type balancesType struct{ available, pending uint64 }
type balancesMapType map[*walletAddress]*balancesType

func (s *server) GetBalance(_ context.Context, _ *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.isSynced() {
		return nil, errors.Errorf("wallet daemon is not synced yet, %s", s.formatSyncStateReport())
	}

	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, err
	}
	daaScore := dagInfo.VirtualDAAScore
	balancesMap := make(balancesMapType, 0)
	for _, entry := range s.utxosSortedByAmount {
		amount := entry.UTXOEntry.Amount()
		address := entry.address
		balances, ok := balancesMap[address]
		if !ok {
			balances = new(balancesType)
			balancesMap[address] = balances
		}
		if s.isUTXOSpendable(entry, daaScore) {
			balances.available += amount
		} else {
			balances.pending += amount
		}
	}

	addressBalances := make([]*pb.AddressBalances, len(balancesMap))
	i := 0
	var available, pending uint64
	for walletAddress, balances := range balancesMap {
		address, err := libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, s.walletAddressPath(walletAddress), s.keysFile.ECDSA)
		if err != nil {
			return nil, err
		}
		addressBalances[i] = &pb.AddressBalances{
			Address:   address.String(),
			Available: balances.available,
			Pending:   balances.pending,
		}
		i++
		available += balances.available
		pending += balances.pending
	}

	log.Infof("GetBalance request scanned %d UTXOs overall over %d addresses", len(s.utxosSortedByAmount), len(balancesMap))

	return &pb.GetBalanceResponse{
		Available:       available,
		Pending:         pending,
		AddressBalances: addressBalances,
	}, nil
}

func (s *server) isUTXOSpendable(entry *walletUTXO, virtualDAAScore uint64) bool {
	if !entry.UTXOEntry.IsCoinbase() {
		return true
	}
	return entry.UTXOEntry.BlockDAAScore()+s.coinbaseMaturity < virtualDAAScore
}
