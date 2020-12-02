package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/pkg/errors"
)

type mutableUTXODiff struct {
	*utxoDiff
}

// NewMutableUTXODiff creates an empty mutable UTXO-Diff
func NewMutableUTXODiff() model.MutableUTXODiff {
	return &mutableUTXODiff{utxoDiff: newUTXODiff()}
}

func (md *mutableUTXODiff) ToUnmutable() model.UTXODiff {
	return md.utxoDiff
}

func (md *mutableUTXODiff) WithDiffInPlace(other model.UTXODiff) error {
	o, ok := other.(*utxoDiff)
	if !ok {
		return errors.New("other is not of type *utxoDiff")
	}
	return withDiffInPlace(md, o)
}

func (md *mutableUTXODiff) AddTransaction(transaction *externalapi.DomainTransaction, blockBlueScore uint64) error {
	for _, input := range transaction.Inputs {
		err := md.removeEntry(&input.PreviousOutpoint, input.UTXOEntry)
		if err != nil {
			return err
		}
	}

	isCoinbase := transactionhelper.IsCoinBase(transaction)
	transactionID := *consensushashing.TransactionID(transaction)
	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			TransactionID: transactionID,
			Index:         uint32(i),
		}
		entry := &externalapi.UTXOEntry{
			Amount:          output.Value,
			ScriptPublicKey: output.ScriptPublicKey,
			BlockBlueScore:  blockBlueScore,
			IsCoinbase:      isCoinbase,
		}

		err := md.addEntry(outpoint, entry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (md *mutableUTXODiff) Clone() model.MutableUTXODiff {
	if md == nil {
		return nil
	}

	return &mutableUTXODiff{utxoDiff: md.utxoDiff.clone()}
}
