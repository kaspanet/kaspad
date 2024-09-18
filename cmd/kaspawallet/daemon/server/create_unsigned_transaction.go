package server

import (
	"context"
	"fmt"
	"math"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/pb"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

// The minimal change amount to target in order to avoid large storage mass (see KIP9 for more details).
// By having at least 10KAS in the change output we make sure that the storage mass charged for change is
// at most 1000 gram. Generally, if the payment is above 10KAS as well, the resulting storage mass will be
// in the order of magnitude of compute mass and wil not incur additional charges.
// Additionally, every transaction with send value > ~0.1 KAS should succeed (at most ~99K storage mass for payment
// output, thus overall lower than standard mass upper bound which is 100K gram)
const minChangeTarget = constants.SompiPerKaspa * 10

// The current minimal fee rate according to mempool standards
const minFeeRate = 1.0

func (s *server) CreateUnsignedTransactions(_ context.Context, request *pb.CreateUnsignedTransactionsRequest) (
	*pb.CreateUnsignedTransactionsResponse, error,
) {
	s.lock.Lock()
	defer s.lock.Unlock()

	unsignedTransactions, err := s.createUnsignedTransactions(request.Address, request.Amount, request.IsSendAll,
		request.From, request.UseExistingChangeAddress, request.FeePolicy)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUnsignedTransactionsResponse{UnsignedTransactions: unsignedTransactions}, nil
}

func (s *server) calculateFeeLimits(requestFeePolicy *pb.FeePolicy) (feeRate float64, maxFee uint64, err error) {
	feeRate = minFeeRate
	maxFee = math.MaxUint64

	if requestFeePolicy == nil {
		requestFeePolicy = &pb.FeePolicy{}
	}

	switch requestFeePolicy := requestFeePolicy.FeePolicy.(type) {
	case *pb.FeePolicy_ExactFeeRate:
		feeRate = requestFeePolicy.ExactFeeRate
		if feeRate < minFeeRate {
			return 0, 0, errors.Errorf("requested fee rate %f is too low, minimum fee rate is %f", feeRate, minFeeRate)
		}
	case *pb.FeePolicy_MaxFeeRate:
		estimate, err := s.rpcClient.GetFeeEstimate()
		if err != nil {
			return 0, 0, err
		}
		if requestFeePolicy.MaxFeeRate < minFeeRate {
			return 0, 0, errors.Errorf("requested max fee rate %f is too low, minimum fee rate is %f", requestFeePolicy.MaxFeeRate, minFeeRate)
		}
		feeRate = math.Min(estimate.Estimate.NormalBuckets[0].Feerate, requestFeePolicy.MaxFeeRate)
	case *pb.FeePolicy_MaxFee:
		estimate, err := s.rpcClient.GetFeeEstimate()
		if err != nil {
			return 0, 0, err
		}
		feeRate = estimate.Estimate.NormalBuckets[0].Feerate
		maxFee = requestFeePolicy.MaxFee
	case nil:
		estimate, err := s.rpcClient.GetFeeEstimate()
		if err != nil {
			return 0, 0, err
		}
		feeRate = estimate.Estimate.NormalBuckets[0].Feerate
		// Default to a bound of max 1 KAS as fee
		maxFee = constants.SompiPerKaspa
	}

	return feeRate, maxFee, nil
}

func (s *server) createUnsignedTransactions(address string, amount uint64, isSendAll bool, fromAddressesString []string, useExistingChangeAddress bool, requestFeePolicy *pb.FeePolicy) ([][]byte, error) {
	if !s.isSynced() {
		return nil, errors.Errorf("wallet daemon is not synced yet, %s", s.formatSyncStateReport())
	}

	feeRate, maxFee, err := s.calculateFeeLimits(requestFeePolicy)
	if err != nil {
		return nil, err
	}

	// make sure address string is correct before proceeding to a
	// potentially long UTXO refreshment operation
	toAddress, err := util.DecodeAddress(address, s.params.Prefix)
	if err != nil {
		return nil, err
	}

	var fromAddresses []*walletAddress
	for _, from := range fromAddressesString {
		fromAddress, exists := s.addressSet[from]
		if !exists {
			return nil, fmt.Errorf("specified from address %s does not exists", from)
		}
		fromAddresses = append(fromAddresses, fromAddress)
	}

	changeAddress, changeWalletAddress, err := s.changeAddress(useExistingChangeAddress, fromAddresses)
	if err != nil {
		return nil, err
	}

	selectedUTXOs, spendValue, changeSompi, err := s.selectUTXOs(amount, isSendAll, feeRate, maxFee, fromAddresses)
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

	unsignedTransactions, err := s.maybeAutoCompoundTransaction(unsignedTransaction, toAddress, changeAddress, changeWalletAddress, feeRate, maxFee)
	if err != nil {
		return nil, err
	}
	return unsignedTransactions, nil
}

func (s *server) selectUTXOs(spendAmount uint64, isSendAll bool, feeRate float64, maxFee uint64, fromAddresses []*walletAddress) (
	selectedUTXOs []*libkaspawallet.UTXO, totalReceived uint64, changeSompi uint64, err error) {
	return s.selectUTXOsWithPreselected(nil, map[externalapi.DomainOutpoint]struct{}{}, spendAmount, isSendAll, feeRate, maxFee, fromAddresses)
}

func (s *server) selectUTXOsWithPreselected(preSelectedUTXOs []*walletUTXO, allowUsed map[externalapi.DomainOutpoint]struct{}, spendAmount uint64, isSendAll bool, feeRate float64, maxFee uint64, fromAddresses []*walletAddress) (
	selectedUTXOs []*libkaspawallet.UTXO, totalReceived uint64, changeSompi uint64, err error) {

	preSelectedSet := make(map[externalapi.DomainOutpoint]struct{})
	for _, utxo := range preSelectedUTXOs {
		preSelectedSet[*utxo.Outpoint] = struct{}{}
	}
	totalValue := uint64(0)

	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, 0, 0, err
	}

	var fee uint64
	iteration := func(utxo *walletUTXO, avoidPreselected bool) (bool, error) {
		if (fromAddresses != nil && !walletAddressesContain(fromAddresses, utxo.address)) ||
			!s.isUTXOSpendable(utxo, dagInfo.VirtualDAAScore) {
			return true, nil
		}

		if broadcastTime, ok := s.usedOutpoints[*utxo.Outpoint]; ok {
			if _, ok := allowUsed[*utxo.Outpoint]; !ok {
				if s.usedOutpointHasExpired(broadcastTime) {
					delete(s.usedOutpoints, *utxo.Outpoint)
				} else {
					return true, nil
				}
			}
		}

		if avoidPreselected {
			if _, ok := preSelectedSet[*utxo.Outpoint]; ok {
				return true, nil
			}
		}

		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			Outpoint:       utxo.Outpoint,
			UTXOEntry:      utxo.UTXOEntry,
			DerivationPath: s.walletAddressPath(utxo.address),
		})

		totalValue += utxo.UTXOEntry.Amount()

		// We're overestimating a bit by assuming that any transaction will have a change output
		fee, err = s.estimateFee(selectedUTXOs, feeRate, maxFee, spendAmount)
		if err != nil {
			return false, err
		}

		totalSpend := spendAmount + fee
		// Two break cases (if not send all):
		// 		1. totalValue == totalSpend, so there's no change needed -> number of outputs = 1, so a single input is sufficient
		// 		2. totalValue > totalSpend, so there will be change and 2 outputs, therefor in order to not struggle with --
		//		   2.1 go-nodes dust patch we try and find at least 2 inputs (even though the next one is not necessary in terms of spend value)
		// 		   2.2 KIP9 we try and make sure that the change amount is not too small
		if !isSendAll && (totalValue == totalSpend || (totalValue >= totalSpend+minChangeTarget && len(selectedUTXOs) > 1)) {
			return false, nil
		}

		return true, nil
	}

	shouldContinue := true
	for _, utxo := range preSelectedUTXOs {
		shouldContinue, err = iteration(utxo, false)
		if err != nil {
			return nil, 0, 0, err
		}

		if !shouldContinue {
			break
		}
	}

	if shouldContinue {
		for _, utxo := range s.utxosSortedByAmount {
			shouldContinue, err := iteration(utxo, true)
			if err != nil {
				return nil, 0, 0, err
			}

			if !shouldContinue {
				break
			}
		}
	}

	var totalSpend uint64
	if isSendAll {
		totalSpend = totalValue
		totalReceived = totalValue - fee
	} else {
		totalSpend = spendAmount + fee
		totalReceived = spendAmount
	}
	if totalValue < totalSpend {
		return nil, 0, 0, errors.Errorf("Insufficient funds for send: %f required, while only %f available",
			float64(totalSpend)/constants.SompiPerKaspa, float64(totalValue)/constants.SompiPerKaspa)
	}

	return selectedUTXOs, totalReceived, totalValue - totalSpend, nil
}

func (s *server) estimateFee(selectedUTXOs []*libkaspawallet.UTXO, feeRate float64, maxFee uint64, recipientValue uint64) (uint64, error) {
	fakePubKey := [util.PublicKeySize]byte{}
	fakeAddr, err := util.NewAddressPublicKey(fakePubKey[:], s.params.Prefix)
	if err != nil {
		return 0, err
	}

	totalValue := uint64(0)
	for _, utxo := range selectedUTXOs {
		totalValue += utxo.UTXOEntry.Amount()
	}

	// This is an approximation for the distribution of value between the recipient output and the change output.
	var mockPayments []*libkaspawallet.Payment
	if totalValue > recipientValue {
		mockPayments = []*libkaspawallet.Payment{
			{
				Address: fakeAddr,
				Amount:  recipientValue,
			},
			{
				Address: fakeAddr,
				Amount:  totalValue - recipientValue, // We ignore the fee since we expect it to be insignificant in mass calculation.
			},
		}
	} else {
		mockPayments = []*libkaspawallet.Payment{
			{
				Address: fakeAddr,
				Amount:  totalValue,
			},
		}
	}

	mockTx, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		mockPayments, selectedUTXOs)
	if err != nil {
		return 0, err
	}

	mass, err := s.estimateMassAfterSignatures(mockTx)
	if err != nil {
		return 0, err
	}

	return min(uint64(math.Ceil(float64(mass)*feeRate)), maxFee), nil
}

func (s *server) estimateFeePerInput(feeRate float64) (uint64, error) {
	mockUTXO := &libkaspawallet.UTXO{
		Outpoint: &externalapi.DomainOutpoint{
			TransactionID: externalapi.DomainTransactionID{},
			Index:         0,
		},
		UTXOEntry: utxo.NewUTXOEntry(1, &externalapi.ScriptPublicKey{
			Script:  nil,
			Version: 0,
		}, false, 0),
		DerivationPath: "m",
	}

	mockTx, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		nil, []*libkaspawallet.UTXO{mockUTXO})
	if err != nil {
		return 0, err
	}

	// Here we use compute mass to avoid dividing by zero. This is ok since `s.estimateFeePerInput` is only used
	// in the case of compound transactions that have a compute mass higher than its storage mass.
	mass, err := s.estimateComputeMassAfterSignatures(mockTx)
	if err != nil {
		return 0, err
	}

	mockTxWithoutUTXO, err := libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		nil, nil)
	if err != nil {
		return 0, err
	}

	massWithoutUTXO, err := s.estimateComputeMassAfterSignatures(mockTxWithoutUTXO)
	if err != nil {
		return 0, err
	}

	inputMass := mass - massWithoutUTXO

	return uint64(float64(inputMass) * feeRate), nil
}

func walletAddressesContain(addresses []*walletAddress, contain *walletAddress) bool {
	for _, address := range addresses {
		if *address == *contain {
			return true
		}
	}

	return false
}
