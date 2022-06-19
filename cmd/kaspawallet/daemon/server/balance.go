package server

import (
	"context"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
)

type (
	balancesType    struct{ available, pending, nUtxosAvailable, nUtxosPending uint64 }
	balancesMapType map[*walletAddress]*balancesType
)

func (s *server) GetBalance(_ context.Context, _ *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, err
	}
	daaScore := dagInfo.VirtualDAAScore
	maturity := s.params.BlockCoinbaseMaturity

	balancesMap := make(balancesMapType, 0)
	for _, entry := range s.utxosSortedByAmount {
		amount := entry.UTXOEntry.Amount()
		address := entry.address
		balances, ok := balancesMap[address]
		if !ok {
			balances = new(balancesType)
			balancesMap[address] = balances
		}
		if isUTXOSpendable(entry, daaScore, maturity) {
			balances.available += amount
			balances.nUtxosAvailable++
		} else {
			balances.pending += amount
			balances.nUtxosPending++
		}
	}

	addressBalances := make([]*pb.AddressBalances, len(balancesMap))
	i := 0
	var available, pending, nUtxosAvailable, nUtxosPending uint64
	for walletAddress, balances := range balancesMap {
		address, err := libkaspawallet.Address(s.params, s.keysFile.ExtendedPublicKeys, s.keysFile.MinimumSignatures, s.walletAddressPath(walletAddress), s.keysFile.ECDSA)
		if err != nil {
			return nil, err
		}
		addressBalances[i] = &pb.AddressBalances{
			Address:         address.String(),
			Available:       balances.available,
			Pending:         balances.pending,
			NUtxosAvailable: balances.nUtxosAvailable,
			NUtxosPending:   balances.nUtxosPending,
		}
		i++
		available += balances.available
		pending += balances.pending
		nUtxosAvailable += balances.nUtxosAvailable
		nUtxosPending += balances.nUtxosPending
	}

	return &pb.GetBalanceResponse{
		Available:       available,
		Pending:         pending,
		NUtxosAvailable: nUtxosAvailable,
		NUtxosPending:   nUtxosPending,
		AddressBalances: addressBalances,
	}, nil
}

func isUTXOSpendable(entry *walletUTXO, virtualDAAScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase() {
		return true
	}
	return entry.UTXOEntry.BlockDAAScore()+coinbaseMaturity < virtualDAAScore
}
