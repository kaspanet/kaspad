package main

import (
	"context"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func (s *server) CreateUnsignedTransaction(_ context.Context, request *pb.CreateUnsignedTransactionRequest) (*pb.CreateUnsignedTransactionResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.refreshExistingUTXOs()
	if err != nil {
		return nil, err
	}

	toAddress, err := util.DecodeAddress(request.Address, s.cfg.NetParams().Prefix)
	if err != nil {
		return nil, err
	}

	sendAmountSompi := request.Amount * util.SompiPerKaspa

	// TODO: Implement a better fee estimation mechanism
	const feePerInput = 1000
	selectedUTXOs, changeSompi, err := s.selectUTXOs(sendAmountSompi, feePerInput)
	if err != nil {
		return nil, err
	}

	changeAddress, err := s.changeAddress()
	if err != nil {
		return nil, err
	}

	psTx, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		[]*libkaspawallet.Payment{{
			Address: toAddress,
			Amount:  sendAmountSompi,
		}, {
			Address: changeAddress,
			Amount:  changeSompi,
		}}, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUnsignedTransactionResponse{UnsignedTransaction: psTx}, nil
}

func (s *server) selectUTXOs(spendAmount uint64, feePerInput uint64) (
	selectedUTXOs []*libkaspawallet.UTXO, changeSompi uint64, err error) {

	selectedUTXOs = []*libkaspawallet.UTXO{}
	totalValue := uint64(0)

	for _, utxo := range s.utxos {
		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			Outpoint:       utxo.Outpoint,
			UTXOEntry:      utxo.UTXOEntry,
			DerivationPath: utxo.address.path(),
		})
		totalValue += utxo.UTXOEntry.Amount()

		fee := feePerInput * uint64(len(selectedUTXOs))
		totalSpend := spendAmount + fee
		if totalValue >= totalSpend {
			break
		}
	}

	fee := feePerInput * uint64(len(selectedUTXOs))
	totalSpend := spendAmount + fee
	if totalValue < totalSpend {
		return nil, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalSpend)/util.SompiPerKaspa, float64(totalValue)/util.SompiPerKaspa)
	}

	return selectedUTXOs, totalValue - totalSpend, nil
}
