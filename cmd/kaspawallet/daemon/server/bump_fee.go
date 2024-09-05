package server

import (
	"context"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

func (s *server) BumpFee(_ context.Context, request *pb.BumpFeeRequest) (*pb.BumpFeeResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	entry, err := s.rpcClient.GetMempoolEntry(request.TxID, false, false)
	if err != nil {
		return nil, err
	}

	domainTx, err := appmessage.RPCTransactionToDomainTransaction(entry.Entry.Transaction)
	if err != nil {
		return nil, err
	}

	mass := s.txMassCalculator.CalculateTransactionMass(domainTx) // TODO: Does GetMempoolEntry already provide the mass?
	feeRate := float64(entry.Entry.Fee) / float64(mass)
	newFeeRate, err := s.calculateFeeRate(request.FeeRate)
	if err != nil {
		return nil, err
	}

	if feeRate >= newFeeRate {
		return nil, errors.Errorf("new fee rate (%f) is not higher than the current fee rate (%f)", newFeeRate, feeRate)
	}

	outpointsSet := make(map[externalapi.DomainOutpoint]struct{})
	for _, input := range domainTx.Inputs {
		outpointsSet[input.PreviousOutpoint] = struct{}{}
	}

	var maxUTXO *walletUTXO
	for _, utxo := range s.utxosSortedByAmount {
		if _, ok := outpointsSet[*utxo.Outpoint]; !ok {
			continue
		}

		if maxUTXO == nil || utxo.UTXOEntry.Amount() > maxUTXO.UTXOEntry.Amount() {
			maxUTXO = utxo
		}
	}

	if maxUTXO == nil {
		return nil, errors.Errorf("no UTXOs were found for transaction %s. This probably means the transaction is already accepted", request.TxID)
	}

	if len(domainTx.Outputs) == 0 || len(domainTx.Outputs) > 2 {
		return nil, errors.Errorf("kaspawallet supports only transactions with 1 or 2 outputs in transaction %s, but this transaction got %d", request.TxID, len(domainTx.Outputs))
	}

	var fromAddresses []*walletAddress
	for _, from := range request.From {
		fromAddress, exists := s.addressSet[from]
		if !exists {
			return nil, errors.Errorf("specified from address %s does not exists", from)
		}
		fromAddresses = append(fromAddresses, fromAddress)
	}

	selectedUTXOs, spendValue, changeSompi, err := s.selectUTXOsWithPreselected([]*walletUTXO{maxUTXO}, outpointsSet, domainTx.Outputs[0].Value, false, newFeeRate, fromAddresses)
	if err != nil {
		return nil, err
	}

	_, toAddress, err := txscript.ExtractScriptPubKeyAddress(domainTx.Outputs[0].ScriptPublicKey, s.params)
	if err != nil {
		return nil, err
	}

	changeAddress, changeWalletAddress, err := s.changeAddress(request.UseExistingChangeAddress, fromAddresses)
	if err != nil {
		return nil, err
	}

	if len(selectedUTXOs) == 0 {
		return nil, errors.Errorf("couldn't find funds to spend")
	}

	payments := []*libkaspawallet.Payment{{
		Address: toAddress,
		Amount:  spendValue,
	}}
	if changeSompi > 0 {
		_, changeAddress, err := txscript.ExtractScriptPubKeyAddress(domainTx.Outputs[0].ScriptPublicKey, s.params)
		if err != nil {
			return nil, err
		}

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

	unsignedTransactions, err := s.maybeAutoCompoundTransaction(unsignedTransaction, toAddress, changeAddress, changeWalletAddress, feeRate)
	if err != nil {
		return nil, err
	}

	if request.Password == "" {
		return &pb.BumpFeeResponse{
			Transactions: unsignedTransactions,
		}, nil
	}

	signedTransactions, err := s.signTransactions(unsignedTransactions, request.Password)
	if err != nil {
		return nil, err
	}

	txIDs, err := s.broadcastRBF(signedTransactions, false)
	if err != nil {
		return nil, err
	}

	return &pb.BumpFeeResponse{
		TxIDs:        txIDs,
		Transactions: signedTransactions,
	}, nil
}
