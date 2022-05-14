package server

import (
	"context"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"golang.org/x/exp/slices"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)


func (s *server) CreateUnsignedTransactionVerbose(_ context.Context, request *pb.CreateUnsignedTransactionVerboseRequest) (
	*pb.CreateUnsignedTransactionsResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	inputs, err := protoOutputsToDomainOutputs(request.Inputs)
	outputs, err := protoPaymentToLibPayment(request.Outputs, s.params.Prefix)

	unsignedTransactions, err := s.createUnsignedTransactionVerbose(inputs, outputs)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUnsignedTransactionsResponse{UnsignedTransactions: unsignedTransactions}, nil
}

func (s *server) createUnsignedTransactionVerbose(inputs []externalapi.DomainOutpoint, payments []*libkaspawallet.Payment) ([][]byte, error) {
	if !s.isSynced() {
		return nil, errors.New("server is not synced")
	}

	err := s.refreshUTXOs()
	if err != nil {
		return nil, err
	}

	selectedUTXOs, totalValue, err := s.selectUTXOsByOutpoints(inputs)
	if err != nil {
		return nil, err
	}
	if len(selectedUTXOs) < len(inputs) {
		return nil, errors.New("Some UTXOs are unavailable")
	}

	totalSpend := uint64(0)
	for _, payment := range payments {
		totalSpend += payment.Amount
	}

	changeSompi := totalValue - totalSpend - feePerInput * uint64(len(selectedUTXOs))
	if changeSompi < 0 {
		return nil, errors.New("Total input is not enough to cover total output and fees")
	}

	changeAddress, _, err := s.changeAddress()
	if err != nil {
		return nil, err
	}

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

	return [][]byte{unsignedTransaction}, nil
}

func (s *server) selectUTXOsByOutpoints(inputs []externalapi.DomainOutpoint) (selectedUTXOs []*libkaspawallet.UTXO, totalValue uint64, err error) {
	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, 0, err
	}
	for _, utxo := range s.utxosSortedByAmount {
		if !slices.Contains(inputs, *utxo.Outpoint) ||
			!isUTXOSpendable(utxo, dagInfo.VirtualDAAScore, s.params.BlockCoinbaseMaturity) {
			continue
		}

		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			Outpoint:       utxo.Outpoint,
			UTXOEntry:      utxo.UTXOEntry,
			DerivationPath: s.walletAddressPath(utxo.address),
		})
		totalValue += utxo.UTXOEntry.Amount()
	}
	return selectedUTXOs, totalValue, nil
}

func protoOutputsToDomainOutputs(request_inputs []*pb.Outpoint) (inputs []externalapi.DomainOutpoint, err error) {
	for _, input := range request_inputs {
		txId, err := externalapi.NewDomainTransactionIDFromString(input.TransactionId)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, externalapi.DomainOutpoint{
			TransactionID: *txId,
			Index:         input.Index,
		})
	}
	return inputs, nil
}

func protoPaymentToLibPayment(request_outputs []*pb.PaymentOutput, prefix util.Bech32Prefix) (outputs []*libkaspawallet.Payment, err error) {
	for _, output := range request_outputs {
		address, err := util.DecodeAddress(output.Address, prefix)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, &libkaspawallet.Payment{
			Address: address,
			Amount:  output.Amount,
		})
	}
	return outputs, nil
}