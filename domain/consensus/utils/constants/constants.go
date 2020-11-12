package constants

import "math"

const (
	// BlockVersion represents the current version of blocks mined and the maximum block version
	// this node is able to validate
	BlockVersion = 1

	// TransactionVersion is the current latest supported transaction version.
	TransactionVersion = 1

	// SompiPerKaspa is the number of sompi in one kaspa (1 KAS).
	SompiPerKaspa = 100_000_000

	// MaxSompi is the maximum transaction amount allowed in sompi.
	MaxSompi = 21_000_000 * SompiPerKaspa

	// MaxCoinbasePayloadLength is the maximum length in bites allowed for a block's coinbase's payload
	MaxCoinbasePayloadLength = 150

	// MaxBlockSize is the maximum size in bytes a block is allowed
	MaxBlockSize = 1_000_000

	// MaxBlockParents is the maximum number of blocks a block is allowed to point to
	MaxBlockParents = 10

	// MassPerTxByte is the number of grams that any byte
	// adds to a transaction.
	MassPerTxByte = 1

	// MassPerScriptPubKeyByte is the number of grams that any
	// scriptPubKey byte adds to a transaction.
	MassPerScriptPubKeyByte = 10

	// MassPerSigOp is the number of grams that any
	// signature operation adds to a transaction.
	MassPerSigOp = 10000

	// MergeSetSizeLimit is the maximum number of blocks in a block's merge set
	MergeSetSizeLimit = 1000

	// MaxMassAcceptedByBlock is the maximum total transaction mass a block may accept.
	MaxMassAcceptedByBlock = 10000000

	// BaseSubsidy is the starting subsidy amount for mined blocks.
	BaseSubsidy = 50 * SompiPerKaspa

	// CoinbasePayloadScriptPublicKeyMaxLength is the maximum allowed script public key in the coinbase's payload
	CoinbasePayloadScriptPublicKeyMaxLength = 150

	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint64 = math.MaxUint64

	// SequenceLockTimeDisabled is a flag that if set on a transaction
	// input's sequence number, the sequence number will not be interpreted
	// as a relative locktime.
	SequenceLockTimeDisabled = 1 << 31

	// SequenceLockTimeIsSeconds is a flag that if set on a transaction
	// input's sequence number, the relative locktime has units of 512
	// seconds.
	SequenceLockTimeIsSeconds = 1 << 22

	// SequenceLockTimeMask is a mask that extracts the relative locktime
	// when masked against the transaction input sequence number.
	SequenceLockTimeMask = 0x0000ffff

	// SequenceLockTimeGranularity is the defined time based granularity
	// for milliseconds-based relative time locks. When converting from milliseconds
	// to a sequence number, the value is right shifted by this amount,
	// therefore the granularity of relative time locks in 524288 or 2^19
	// seconds. Enforced relative lock times are multiples of 524288 milliseconds.
	SequenceLockTimeGranularity = 19
)
