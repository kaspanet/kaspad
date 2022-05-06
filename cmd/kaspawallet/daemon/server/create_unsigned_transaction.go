package server

import (
	"context"
	"fmt"
	"golang.org/x/exp/slices"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// TODO: Implement a better fee estimation mechanism
const feePerInput = 10000

func (s *server) CreateUnsignedTransactions(_ context.Context, request *pb.CreateUnsignedTransactionsRequest) (
	*pb.CreateUnsignedTransactionsResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	unsignedTransactions, err := s.createUnsignedTransactions(request.Address, request.Amount, request.From)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUnsignedTransactionsResponse{UnsignedTransactions: unsignedTransactions}, nil
}

func (s *server) createUnsignedTransactions(address string, amount uint64, fromAddressesString []string) ([][]byte, error) {
	if !s.isSynced() {
		return nil, errors.New("server is not synced")
	}

	err := s.refreshUTXOs()
	if err != nil {
		return nil, err
	}

	toAddress, err := util.DecodeAddress(address, s.params.Prefix)
	if err != nil {
		return nil, err
	}

	var fromAddresses []*walletAddress
	for _, from := range fromAddressesString {
		fromAddress, exists := s.addressSet[from]
		if !exists {
			return nil, fmt.Errorf("Specified from address %s does not exists", from)
		}
		fromAddresses = append(fromAddresses, fromAddress)
	}

	selectedUTXOs, changeSompi, err := s.selectUTXOs(amount, feePerInput, fromAddresses)
	if err != nil {
		return nil, err
	}

	changeAddress, changeWalletAddress, err := s.changeAddress()
	if err != nil {
		return nil, err
	}

	payments := []*libkaspawallet.Payment{{
		Address: toAddress,
		Amount:  amount,
	}}
	if changeSompi > 0 {
		payments = append(payments, &libkaspawallet.Payment{
			Address: changeAddress,
			Amount:  changeSompi,
		})
	}
	unsignedTransaction, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		payments, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	unsignedTransactions, err := s.maybeAutoCompoundTransaction(unsignedTransaction, toAddress, changeAddress, changeWalletAddress)
	if err != nil {
		return nil, err
	}
	return unsignedTransactions, nil
}

func (s *server) selectUTXOs(spendAmount uint64, feePerInput uint64, fromAddresses []*walletAddress) (
	selectedUTXOs []*libkaspawallet.UTXO, changeSompi uint64, err error) {

	selectedUTXOs = []*libkaspawallet.UTXO{}
	totalValue := uint64(0)

	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, 0, err
	}

	for _, utxo := range s.availableUtxosSortedByAmount {
		if (fromAddresses != nil && !slices.Contains(fromAddresses, utxo.address)) ||
			!isUTXOSpendable(utxo, dagInfo.VirtualDAAScore, s.params.BlockCoinbaseMaturity) {
			continue
		}
		s.tracker.trackOutpointAsReserved(*utxo.Outpoint)

		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			Outpoint:       utxo.Outpoint,
			UTXOEntry:      utxo.UTXOEntry,
			DerivationPath: s.walletAddressPath(utxo.address),
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
		for _, utxo := range selectedUTXOs {
			s.tracker.untrackOutpointAsReserved(*utxo.Outpoint)
		}
		return nil, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalSpend)/constants.SompiPerKaspa, float64(totalValue)/constants.SompiPerKaspa)
	}

	return selectedUTXOs, totalValue - totalSpend, nil
}
