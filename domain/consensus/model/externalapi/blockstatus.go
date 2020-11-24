package externalapi

// BlockStatus represents the validation state of the block.
type BlockStatus byte

// Clone returns a clone of TransactionAcceptanceData
func (bs BlockStatus) Clone() BlockStatus {
	return bs
}

const (
	// StatusInvalid indicates that the block is invalid.
	StatusInvalid BlockStatus = iota

	// StatusValid indicates that the block has been fully validated.
	StatusValid

	// StatusUTXOPendingVerification indicates that the block is pending verification against its past UTXO-Set, either
	// because it was not yet verified since the block was never in the selected parent chain, or if the
	// block violates finality.
	StatusUTXOPendingVerification

	// StatusDisqualifiedFromChain indicates that the block is not eligible to be a selected parent.
	StatusDisqualifiedFromChain

	// StatusHeaderOnly indicates that the block transactions are not held (pruned or wasn't added yet)
	StatusHeaderOnly
)

var blockStatusStrings = map[BlockStatus]string{
	StatusInvalid:                 "Invalid",
	StatusValid:                   "Valid",
	StatusUTXOPendingVerification: "UTXOPendingVerification",
	StatusDisqualifiedFromChain:   "DisqualifiedFromChain",
	StatusHeaderOnly:              "HeaderOnly",
}

func (bs BlockStatus) String() string {
	return blockStatusStrings[bs]
}
