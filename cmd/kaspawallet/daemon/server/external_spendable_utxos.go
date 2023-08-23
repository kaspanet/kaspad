package server

import (
	"context"

	"github.com/c4ei/YunSeokYeol/app/appmessage"
	"github.com/c4ei/YunSeokYeol/cmd/kaspawallet/daemon/pb"
	"github.com/c4ei/YunSeokYeol/cmd/kaspawallet/libkaspawallet"
	"github.com/c4ei/YunSeokYeol/util"
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
	selectedUTXOs, err := s.selectExternalSpendableUTXOs(externalUTXOs, request.Address)
	if err != nil {
		return nil, err
	}
	return &pb.GetExternalSpendableUTXOsResponse{
		Entries: selectedUTXOs,
	}, nil
}

func (s *server) selectExternalSpendableUTXOs(externalUTXOs *appmessage.GetUTXOsByAddressesResponseMessage, address string) ([]*pb.UtxosByAddressesEntry, error) {
	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, err
	}

	daaScore := dagInfo.VirtualDAAScore
	maturity := s.params.BlockCoinbaseMaturity

	//we do not make because we do not know size, because of unspendable utxos
	var selectedExternalUtxos []*pb.UtxosByAddressesEntry

	for _, entry := range externalUTXOs.Entries {
		if !isExternalUTXOSpendable(entry, daaScore, maturity) {
			continue
		}
		selectedExternalUtxos = append(selectedExternalUtxos, libkaspawallet.AppMessageUTXOToKaspawalletdUTXO(entry))
	}

	return selectedExternalUtxos, nil
}

func isExternalUTXOSpendable(entry *appmessage.UTXOsByAddressesEntry, virtualDAAScore uint64, coinbaseMaturity uint64) bool {
	if !entry.UTXOEntry.IsCoinbase {
		return true
	} else if entry.UTXOEntry.Amount <= feePerInput {
		return false
	}
	return entry.UTXOEntry.BlockDAAScore+coinbaseMaturity < virtualDAAScore
}
