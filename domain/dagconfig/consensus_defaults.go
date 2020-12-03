package dagconfig

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"time"
)

const (
	defaultMaxCoinbasePayloadLength                = 150
	defaultMaxBlockSize                            = 1_000_000
	defaultMaxBlockParents                         = 10
	defaultMassPerTxByte                           = 1
	defaultMassPerScriptPubKeyByte                 = 10
	defaultMassPerSigOp                            = 10000
	defaultMergeSetSizeLimit                       = 1000
	defaultMaxMassAcceptedByBlock                  = 10000000
	defaultBaseSubsidy                             = 50 * constants.SompiPerKaspa
	defaultCoinbasePayloadScriptPublicKeyMaxLength = 150
	defaultGHOSTDAGK                               = 18
	defaultDifficultyAdjustmentWindowSize          = 2640
	defaultTimestampDeviationTolerance             = 132
	defaultFinalityDuration                        = 24 * time.Hour
	defaultTargetTimePerBlock                      = 1 * time.Second
)
