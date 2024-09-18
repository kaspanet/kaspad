package server

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/txmass"
)

// maybeAutoCompoundTransaction checks if a transaction's mass is higher that what is allowed for a standard
// transaction.
// If it is - the transaction is split into multiple transactions, each with a portion of the inputs and a single output
// into a change address.
// An additional `mergeTransaction` is generated - which merges the outputs of the above splits into a single output
// paying to the original transaction's payee.
func (s *server) maybeAutoCompoundTransaction(transaction *serialization.PartiallySignedTransaction, toAddress util.Address,
	changeAddress util.Address, changeWalletAddress *walletAddress, feeRate float64, maxFee uint64) ([][]byte, error) {

	splitTransactions, err := s.maybeSplitAndMergeTransaction(transaction, toAddress, changeAddress, changeWalletAddress, feeRate, maxFee)
	if err != nil {
		return nil, err
	}
	splitTransactionsBytes := make([][]byte, len(splitTransactions))
	for i, splitTransaction := range splitTransactions {
		splitTransactionsBytes[i], err = serialization.SerializePartiallySignedTransaction(splitTransaction)
		if err != nil {
			return nil, err
		}
	}
	return splitTransactionsBytes, nil
}

func (s *server) mergeTransaction(
	splitTransactions []*serialization.PartiallySignedTransaction,
	originalTransaction *serialization.PartiallySignedTransaction,
	toAddress util.Address,
	changeAddress util.Address,
	changeWalletAddress *walletAddress,
	feeRate float64,
	maxFee uint64,
) (*serialization.PartiallySignedTransaction, error) {
	numOutputs := len(originalTransaction.Tx.Outputs)
	if numOutputs > 2 || numOutputs == 0 {
		// This is a sanity check to make sure originalTransaction has either 1 or 2 outputs:
		// 1. For the payment itself
		// 2. (optional) for change
		return nil, errors.Errorf("original transaction has %d outputs, while 1 or 2 are expected",
			len(originalTransaction.Tx.Outputs))
	}

	totalValue := uint64(0)
	sentValue := originalTransaction.Tx.Outputs[0].Value
	utxos := make([]*libkaspawallet.UTXO, len(splitTransactions))
	for i, splitTransaction := range splitTransactions {
		output := splitTransaction.Tx.Outputs[0]
		utxos[i] = &libkaspawallet.UTXO{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(splitTransaction.Tx),
				Index:         0,
			},
			UTXOEntry:      utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, false, constants.UnacceptedDAAScore),
			DerivationPath: s.walletAddressPath(changeWalletAddress),
		}
		totalValue += output.Value
	}
	// We're overestimating a bit by assuming that any transaction will have a change output
	fee, err := s.estimateFee(utxos, feeRate, maxFee, sentValue)
	if err != nil {
		return nil, err
	}

	totalValue -= fee

	if totalValue < sentValue {
		// sometimes the fees from compound transactions make the total output higher than what's available from selected
		// utxos, in such cases - find one more UTXO and use it.
		additionalUTXOs, totalValueAdded, err := s.moreUTXOsForMergeTransaction(utxos, sentValue-totalValue, feeRate)
		if err != nil {
			return nil, err
		}
		utxos = append(utxos, additionalUTXOs...)
		totalValue += totalValueAdded
	}

	payments := []*libkaspawallet.Payment{{
		Address: toAddress,
		Amount:  sentValue,
	}}
	if totalValue > sentValue {
		payments = append(payments, &libkaspawallet.Payment{
			Address: changeAddress,
			Amount:  totalValue - sentValue,
		})
	}

	return libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures, payments, utxos)
}

func (s *server) transactionFeeRate(psTx *serialization.PartiallySignedTransaction) (float64, error) {
	totalOuts := 0
	for _, output := range psTx.Tx.Outputs {
		totalOuts += int(output.Value)
	}

	totalIns := 0
	for _, input := range psTx.PartiallySignedInputs {
		totalIns += int(input.PrevOutput.Value)
	}

	if totalIns < totalOuts {
		return 0, errors.Errorf("Transaction don't have enough funds to pay for the outputs")
	}
	fee := totalIns - totalOuts
	mass, err := s.estimateComputeMassAfterSignatures(psTx)
	if err != nil {
		return 0, err
	}
	return float64(fee) / float64(mass), nil
}

func (s *server) checkTransactionFeeRate(psTx *serialization.PartiallySignedTransaction, maxFee uint64) error {
	feeRate, err := s.transactionFeeRate(psTx)
	if err != nil {
		return err
	}

	if feeRate < 1 {
		return errors.Errorf("setting --max-fee to %d results in a fee rate of %f, which is below the minimum allowed fee rate of 1 sompi/gram", maxFee, feeRate)
	}

	return nil
}

func (s *server) maybeSplitAndMergeTransaction(transaction *serialization.PartiallySignedTransaction, toAddress util.Address,
	changeAddress util.Address, changeWalletAddress *walletAddress, feeRate float64, maxFee uint64) ([]*serialization.PartiallySignedTransaction, error) {

	err := s.checkTransactionFeeRate(transaction, maxFee)
	if err != nil {
		return nil, err
	}

	transactionMass, err := s.estimateComputeMassAfterSignatures(transaction)
	if err != nil {
		return nil, err
	}

	if transactionMass < mempool.MaximumStandardTransactionMass {
		return []*serialization.PartiallySignedTransaction{transaction}, nil
	}

	splitCount, inputCountPerSplit, err := s.splitAndInputPerSplitCounts(transaction, transactionMass, changeAddress, feeRate, maxFee)
	if err != nil {
		return nil, err
	}

	splitTransactions := make([]*serialization.PartiallySignedTransaction, splitCount)
	for i := 0; i < splitCount; i++ {
		startIndex := i * inputCountPerSplit
		endIndex := startIndex + inputCountPerSplit
		var err error
		splitTransactions[i], err = s.createSplitTransaction(transaction, changeAddress, startIndex, endIndex, feeRate, maxFee)
		if err != nil {
			return nil, err
		}

		err = s.checkTransactionFeeRate(splitTransactions[i], maxFee)
		if err != nil {
			return nil, err
		}
	}

	if len(splitTransactions) > 1 {
		mergeTransaction, err := s.mergeTransaction(splitTransactions, transaction, toAddress, changeAddress, changeWalletAddress, feeRate, maxFee)
		if err != nil {
			return nil, err
		}
		// Recursion will be 2-3 iterations deep even in the rarest` cases, so considered safe..
		splitMergeTransaction, err := s.maybeSplitAndMergeTransaction(mergeTransaction, toAddress, changeAddress, changeWalletAddress, feeRate, maxFee)
		if err != nil {
			return nil, err
		}
		splitTransactions = append(splitTransactions, splitMergeTransaction...)

	}

	return splitTransactions, nil
}

// splitAndInputPerSplitCounts calculates the number of splits to create, and the number of inputs to assign per split.
func (s *server) splitAndInputPerSplitCounts(transaction *serialization.PartiallySignedTransaction, transactionMass uint64,
	changeAddress util.Address, feeRate float64, maxFee uint64) (splitCount, inputsPerSplitCount int, err error) {

	// Create a dummy transaction which is a clone of the original transaction, but without inputs,
	// to calculate how much mass do all the inputs have
	transactionWithoutInputs := transaction.Tx.Clone()
	transactionWithoutInputs.Inputs = []*externalapi.DomainTransactionInput{}
	massWithoutInputs := s.txMassCalculator.CalculateTransactionMass(transactionWithoutInputs)

	massOfAllInputs := transactionMass - massWithoutInputs

	// Since the transaction was generated by kaspawallet, we assume all inputs have the same number of signatures, and
	// thus - the same mass.
	inputCount := len(transaction.Tx.Inputs)
	massPerInput := massOfAllInputs / uint64(inputCount)
	if massOfAllInputs%uint64(inputCount) > 0 {
		massPerInput++
	}

	// Create another dummy transaction, this time one similar to the split transactions we wish to generate,
	// but with 0 inputs, to calculate how much mass for inputs do we have available in the split transactions
	splitTransactionWithoutInputs, err := s.createSplitTransaction(transaction, changeAddress, 0, 0, feeRate, maxFee)
	if err != nil {
		return 0, 0, err
	}
	massForEverythingExceptInputsInSplitTransaction :=
		s.txMassCalculator.CalculateTransactionMass(splitTransactionWithoutInputs.Tx)
	massForInputsInSplitTransaction := mempool.MaximumStandardTransactionMass - massForEverythingExceptInputsInSplitTransaction

	inputsPerSplitCount = int(massForInputsInSplitTransaction / massPerInput)
	splitCount = inputCount / inputsPerSplitCount
	if inputCount%inputsPerSplitCount > 0 {
		splitCount++
	}

	return splitCount, inputsPerSplitCount, nil
}

func (s *server) createSplitTransaction(transaction *serialization.PartiallySignedTransaction,
	changeAddress util.Address, startIndex int, endIndex int, feeRate float64, maxFee uint64) (*serialization.PartiallySignedTransaction, error) {

	selectedUTXOs := make([]*libkaspawallet.UTXO, 0, endIndex-startIndex)
	totalSompi := uint64(0)

	for i := startIndex; i < endIndex && i < len(transaction.PartiallySignedInputs); i++ {
		partiallySignedInput := transaction.PartiallySignedInputs[i]
		selectedUTXOs = append(selectedUTXOs, &libkaspawallet.UTXO{
			Outpoint: &transaction.Tx.Inputs[i].PreviousOutpoint,
			UTXOEntry: utxo.NewUTXOEntry(
				partiallySignedInput.PrevOutput.Value, partiallySignedInput.PrevOutput.ScriptPublicKey,
				false, constants.UnacceptedDAAScore),
			DerivationPath: partiallySignedInput.DerivationPath,
		})

		totalSompi += selectedUTXOs[i-startIndex].UTXOEntry.Amount()
	}
	if len(selectedUTXOs) != 0 {
		fee, err := s.estimateFee(selectedUTXOs, feeRate, maxFee, totalSompi)
		if err != nil {
			return nil, err
		}

		totalSompi -= fee
	}

	return libkaspawallet.CreateUnsignedTransaction(s.keysFile.ExtendedPublicKeys,
		s.keysFile.MinimumSignatures,
		[]*libkaspawallet.Payment{{
			Address: changeAddress,
			Amount:  totalSompi,
		}}, selectedUTXOs)
}

func (s *server) estimateMassAfterSignatures(transaction *serialization.PartiallySignedTransaction) (uint64, error) {
	return EstimateMassAfterSignatures(transaction, s.keysFile.ECDSA, s.keysFile.MinimumSignatures, s.txMassCalculator)
}

func (s *server) estimateComputeMassAfterSignatures(transaction *serialization.PartiallySignedTransaction) (uint64, error) {
	return estimateComputeMassAfterSignatures(transaction, s.keysFile.ECDSA, s.keysFile.MinimumSignatures, s.txMassCalculator)
}

func createTransactionWithJunkFieldsForMassCalculation(transaction *serialization.PartiallySignedTransaction, ecdsa bool, minimumSignatures uint32, txMassCalculator *txmass.Calculator) (*externalapi.DomainTransaction, error) {
	transaction = transaction.Clone()
	var signatureSize uint64
	if ecdsa {
		signatureSize = secp256k1.SerializedECDSASignatureSize
	} else {
		signatureSize = secp256k1.SerializedSchnorrSignatureSize
	}

	for i, input := range transaction.PartiallySignedInputs {
		for j, pubKeyPair := range input.PubKeySignaturePairs {
			if uint32(j) >= minimumSignatures {
				break
			}
			pubKeyPair.Signature = make([]byte, signatureSize+1) // +1 for SigHashType
		}
		transaction.Tx.Inputs[i].SigOpCount = byte(len(input.PubKeySignaturePairs))
	}

	return libkaspawallet.ExtractTransactionDeserialized(transaction, ecdsa)
}

func estimateComputeMassAfterSignatures(transaction *serialization.PartiallySignedTransaction, ecdsa bool, minimumSignatures uint32, txMassCalculator *txmass.Calculator) (uint64, error) {
	transactionWithSignatures, err := createTransactionWithJunkFieldsForMassCalculation(transaction, ecdsa, minimumSignatures, txMassCalculator)
	if err != nil {
		return 0, err
	}

	return txMassCalculator.CalculateTransactionMass(transactionWithSignatures), nil
}

func EstimateMassAfterSignatures(transaction *serialization.PartiallySignedTransaction, ecdsa bool, minimumSignatures uint32, txMassCalculator *txmass.Calculator) (uint64, error) {
	transactionWithSignatures, err := createTransactionWithJunkFieldsForMassCalculation(transaction, ecdsa, minimumSignatures, txMassCalculator)
	if err != nil {
		return 0, err
	}

	return txMassCalculator.CalculateTransactionOverallMass(transactionWithSignatures), nil
}

func (s *server) moreUTXOsForMergeTransaction(alreadySelectedUTXOs []*libkaspawallet.UTXO, requiredAmount uint64, feeRate float64) (
	additionalUTXOs []*libkaspawallet.UTXO, totalValueAdded uint64, err error) {

	dagInfo, err := s.rpcClient.GetBlockDAGInfo()
	if err != nil {
		return nil, 0, err
	}
	alreadySelectedUTXOsMap := make(map[externalapi.DomainOutpoint]struct{}, len(alreadySelectedUTXOs))
	for _, alreadySelectedUTXO := range alreadySelectedUTXOs {
		alreadySelectedUTXOsMap[*alreadySelectedUTXO.Outpoint] = struct{}{}
	}

	feePerInput, err := s.estimateFeePerInput(feeRate)
	if err != nil {
		return nil, 0, err
	}

	for _, utxo := range s.utxosSortedByAmount {
		if _, ok := alreadySelectedUTXOsMap[*utxo.Outpoint]; ok {
			continue
		}
		if !s.isUTXOSpendable(utxo, dagInfo.VirtualDAAScore) {
			continue
		}
		additionalUTXOs = append(additionalUTXOs, &libkaspawallet.UTXO{
			Outpoint:       utxo.Outpoint,
			UTXOEntry:      utxo.UTXOEntry,
			DerivationPath: s.walletAddressPath(utxo.address)})
		totalValueAdded += utxo.UTXOEntry.Amount() - feePerInput
		if totalValueAdded >= requiredAmount {
			break
		}
	}
	if totalValueAdded < requiredAmount {
		return nil, 0, errors.Errorf("Insufficient funds for merge transaction")
	}

	return additionalUTXOs, totalValueAdded, nil
}
