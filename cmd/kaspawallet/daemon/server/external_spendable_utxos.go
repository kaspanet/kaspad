package server

import (
	"context"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
)

func (s *server) GetExternalSpendableUTXOs(_ context.Context, request *pb.GetExternalSpendableUTXOsRequest) (*pb.GetExternalSpendableUTXOsResponse, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, err := util.DecodeAddress(request.Address, s.params.Prefix)
	if err != nil {
		return nil, err
	}
	externalUTXOs, err := s.rpcClient.GetUTXOsByAddresses([]string{request.Address})
	if err != nil {
		return nil, err
	}

	estimate, err := s.rpcClient.GetFeeEstimate()
	if err != nil {
		return nil, err
	}

	feeRate := estimate.Estimate.NormalBuckets[0].Feerate

	selectedUTXOs, err := s.selectExternalSpendableUTXOs(externalUTXOs, feeRate)
	if err != nil {
		return nil, err
	}
	return &pb.GetExternalSpendableUTXOsResponse{
		Entries: selectedUTXOs,
	}, nil
}

func (s *server) selectExternalSpendableUTXOs(externalUTXOs *appmessage.GetUTXOsByAddressesResponseMessage, feeRate float64) ([]*pb.UtxosByAddressesEntry, error) {
	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, err
	}

	daaScore := dagInfo.VirtualDAAScore
	maturity := s.params.BlockCoinbaseMaturity

	//we do not make because we do not know size, because of unspendable utxos
	var selectedExternalUtxos []*pb.UtxosByAddressesEntry

	feePerInput, err := s.estimateFeePerInput(feeRate)
	if err != nil {
		return nil, err
	}

	for _, entry := range externalUTXOs.Entries {
		if !isExternalUTXOSpendable(entry, daaScore, maturity, feePerInput) {
			continue
		}
		selectedExternalUtxos = append(selectedExternalUtxos, libkaspawallet.AppMessageUTXOToKaspawalletdUTXO(entry))
	}

	return selectedExternalUtxos, nil
}

func isExternalUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualDAAScore uint64, coinbaseMaturity uint64, feePerInput uint64) bool {
	if !entry.UTXOEntry.IsCoinbase {
		return true
	} else if entry.UTXOEntry.Amount <= feePerInput {
		return false
	}
	return entry.UTXOEntry.BlockDAAScore+coinbaseMaturity < virtualDAAScore
}
