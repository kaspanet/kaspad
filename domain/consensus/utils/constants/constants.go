package constants

import "math"

const (
	// MaxBlockVersion represents the current version of blocks mined and the maximum block version
	// this node is able to validate
	MaxBlockVersion uint16 = 0

	// MaxTransactionVersion is the current latest supported transaction version.
	MaxTransactionVersion uint16 = 0

	// MaxScriptPublicKeyVersion is the current latest supported public key script version.
	MaxScriptPublicKeyVersion uint16 = 0

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
	SequenceLockTimeDisabled uint64 = 1 << 63

	// SequenceLockTimeIsSeconds is a flag that if set on a transaction
	// input's sequence number, the relative locktime has units of 524,2788
	// milliseconds
	SequenceLockTimeIsSeconds uint64 = 1 << 62

	// SequenceLockTimeMask is a mask that extracts the relative locktime
	// when masked against the transaction input sequence number.
	SequenceLockTimeMask uint64 = 0x00000000ffffffff

	// SequenceLockTimeGranularity is the defined time based granularity
	// for milliseconds-based relative time locks. When converting from milliseconds
	// to a sequence number, the value is multiplied by this amount,
	// therefore the granularity of relative time locks 1000 milliseconds.
	// Enforced relative lock times are multiples of 524288 milliseconds.
	SequenceLockTimeGranularity = 1000

	// LockTimeThreshold is the number below which a lock time is
	// interpreted to be a block number. Since an average of one block
	// is generated per 10 minutes, this allows blocks for about 9,512
	// years.
	LockTimeThreshold = 5e8 // Tue Nov 5 00:53:20 1985 UTC
)
