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
