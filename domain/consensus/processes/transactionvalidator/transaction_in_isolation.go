package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/pkg/errors"
)

func (v *transactionValidator) ValidateTransactionInIsolation(tx *externalapi.DomainTransaction) error {
	err := v.checkTransactionInputCount(tx)
	if err != nil {
		return err
	}
	err = v.checkTransactionAmountRanges(tx)
	if err != nil {
		return err
	}
	err = v.checkDuplicateTransactionInputs(tx)
	if err != nil {
		return err
	}
	err = v.checkCoinbaseLength(tx)
	if err != nil {
		return err
	}
	err = v.checkTransactionPayloadHash(tx)
	if err != nil {
		return err
	}
	err = v.checkGasInBuiltInOrNativeTransactions(tx)
	if err != nil {
		return err
	}
	err = v.checkSubnetworkRegistryTransaction(tx)
	if err != nil {
		return err
	}

	err = v.checkNativeTransactionPayload(tx)
	if err != nil {
		return err
	}

	// TODO: fill it with the right subnetwork id.
	err = v.checkTransactionSubnetwork(tx, nil)
	if err != nil {
		return err
	}
	return nil
}

func (v *transactionValidator) checkTransactionInputCount(tx *externalapi.DomainTransaction) error {
	// A non-coinbase transaction must have at least one input.
	if !transactionhelper.IsCoinBase(tx) && len(tx.Inputs) == 0 {
		return errors.Wrapf(ruleerrors.ErrNoTxInputs, "transaction has no inputs")
	}
	return nil
}

func (v *transactionValidator) checkTransactionAmountRanges(tx *externalapi.DomainTransaction) error {
	// Ensure the transaction amounts are in range. Each transaction
	// output must not be negative or more than the max allowed per
	// transaction. Also, the total of all outputs must abide by the same
	// restrictions. All amounts in a transaction are in a unit value known
	// as a sompi. One kaspa is a quantity of sompi as defined by the
	// sompiPerKaspa constant.
	var totalSompi uint64
	for _, txOut := range tx.Outputs {
		sompi := txOut.Value
		if sompi > constants.MaxSompi {
			return errors.Wrapf(ruleerrors.ErrBadTxOutValue, "transaction output value of %d is "+
				"higher than max allowed value of %d", sompi, constants.MaxSompi)
		}

		// Binary arithmetic guarantees that any overflow is detected and reported.
		// This is impossible for Kaspa, but perhaps possible if an alt increases
		// the total money supply.
		newTotalSompi := totalSompi + sompi
		if newTotalSompi < totalSompi {
			return errors.Wrapf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
				"outputs exceeds max allowed value of %d",
				constants.MaxSompi)
		}
		totalSompi = newTotalSompi
		if totalSompi > constants.MaxSompi {
			return errors.Wrapf(ruleerrors.ErrBadTxOutValue, "total value of all transaction "+
				"outputs is %d which is higher than max "+
				"allowed value of %d", totalSompi,
				constants.MaxSompi)
		}
	}

	return nil
}

func (v *transactionValidator) checkDuplicateTransactionInputs(tx *externalapi.DomainTransaction) error {
	existingTxOut := make(map[externalapi.DomainOutpoint]struct{})
	for _, txIn := range tx.Inputs {
		if _, exists := existingTxOut[txIn.PreviousOutpoint]; exists {
			return errors.Wrapf(ruleerrors.ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOutpoint] = struct{}{}
	}
	return nil
}

func (v *transactionValidator) checkCoinbaseLength(tx *externalapi.DomainTransaction) error {
	if !transactionhelper.IsCoinBase(tx) {
		return nil
	}

	// Coinbase payload length must not exceed the max length.
	payloadLen := len(tx.Payload)
	if payloadLen > constants.MaxCoinbasePayloadLength {
		return errors.Wrapf(ruleerrors.ErrBadCoinbasePayloadLen, "coinbase transaction payload length "+
			"of %d is out of range (max: %d)",
			payloadLen, constants.MaxCoinbasePayloadLength)
	}

	return nil
}

func (v *transactionValidator) checkTransactionPayloadHash(tx *externalapi.DomainTransaction) error {
	if tx.SubnetworkID != subnetworks.SubnetworkIDNative {
		payloadHash := hashes.HashData(tx.Payload)
		if tx.PayloadHash != *payloadHash {
			return errors.Wrapf(ruleerrors.ErrInvalidPayloadHash, "invalid payload hash")
		}
	} else if tx.PayloadHash != (externalapi.DomainHash{}) {
		return errors.Wrapf(ruleerrors.ErrInvalidPayloadHash, "unexpected non-empty payload hash in native subnetwork")
	}
	return nil
}

func (v *transactionValidator) checkGasInBuiltInOrNativeTransactions(tx *externalapi.DomainTransaction) error {
	// Transactions in native, registry and coinbase subnetworks must have Gas = 0
	if subnetworks.IsBuiltInOrNative(tx.SubnetworkID) && tx.Gas > 0 {
		return errors.Wrapf(ruleerrors.ErrInvalidGas, "transaction in the native or "+
			"registry subnetworks has gas > 0 ")
	}
	return nil
}

func (v *transactionValidator) checkSubnetworkRegistryTransaction(tx *externalapi.DomainTransaction) error {
	if tx.SubnetworkID != subnetworks.SubnetworkIDRegistry {
		return nil
	}

	if len(tx.Payload) != 8 {
		return errors.Wrapf(ruleerrors.ErrSubnetworkRegistry, "validation failed: subnetwork registry "+
			"tx has an invalid payload")
	}
	return nil
}

func (v *transactionValidator) checkNativeTransactionPayload(tx *externalapi.DomainTransaction) error {
	if tx.SubnetworkID == subnetworks.SubnetworkIDNative && len(tx.Payload) > 0 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload, "transaction in the native subnetwork "+
			"includes a payload")
	}
	return nil
}

func (v *transactionValidator) checkTransactionSubnetwork(tx *externalapi.DomainTransaction,
	subnetworkID *externalapi.DomainSubnetworkID) error {
	if !v.enableNonNativeSubnetworks && tx.SubnetworkID != subnetworks.SubnetworkIDNative &&
		tx.SubnetworkID != subnetworks.SubnetworkIDCoinbase {
		return errors.Wrapf(ruleerrors.ErrSubnetworksDisabled, "transaction has non native or coinbase "+
			"subnetwork ID")
	}

	// If we are a partial node, only transactions on built in subnetworks
	// or our own subnetwork may have a payload
	isLocalNodeFull := subnetworkID == nil
	shouldTxBeFull := subnetworks.IsBuiltIn(tx.SubnetworkID) || subnetworks.IsEqual(&tx.SubnetworkID, subnetworkID)
	if !isLocalNodeFull && !shouldTxBeFull && len(tx.Payload) > 0 {
		return errors.Wrapf(ruleerrors.ErrInvalidPayload,
			"transaction that was expected to be partial has a payload "+
				"with length > 0")
	}
	return nil
}
