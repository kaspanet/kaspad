package transactionvalidator

import (
	"github.com/c4ei/YunSeokYeol/domain/consensus/model"
	"github.com/c4ei/YunSeokYeol/domain/consensus/model/externalapi"
	"github.com/c4ei/YunSeokYeol/domain/consensus/utils/txscript"
	"github.com/c4ei/YunSeokYeol/util/txmass"
)

const sigCacheSize = 10_000

// transactionValidator exposes a set of validation classes, after which
// it's possible to determine whether either a transaction is valid
type transactionValidator struct {
	blockCoinbaseMaturity                   uint64
	databaseContext                         model.DBReader
	pastMedianTimeManager                   model.PastMedianTimeManager
	ghostdagDataStore                       model.GHOSTDAGDataStore
	daaBlocksStore                          model.DAABlocksStore
	enableNonNativeSubnetworks              bool
	maxCoinbasePayloadLength                uint64
	ghostdagK                               externalapi.KType
	coinbasePayloadScriptPublicKeyMaxLength uint8
	sigCache                                *txscript.SigCache
	sigCacheECDSA                           *txscript.SigCacheECDSA
	txMassCalculator                        *txmass.Calculator
}

// New instantiates a new TransactionValidator
func New(blockCoinbaseMaturity uint64,
	enableNonNativeSubnetworks bool,
	maxCoinbasePayloadLength uint64,
	ghostdagK externalapi.KType,
	coinbasePayloadScriptPublicKeyMaxLength uint8,
	databaseContext model.DBReader,
	pastMedianTimeManager model.PastMedianTimeManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	daaBlocksStore model.DAABlocksStore,
	txMassCalculator *txmass.Calculator) model.TransactionValidator {

	return &transactionValidator{
		blockCoinbaseMaturity:                   blockCoinbaseMaturity,
		enableNonNativeSubnetworks:              enableNonNativeSubnetworks,
		maxCoinbasePayloadLength:                maxCoinbasePayloadLength,
		ghostdagK:                               ghostdagK,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,
		databaseContext:                         databaseContext,
		pastMedianTimeManager:                   pastMedianTimeManager,
		ghostdagDataStore:                       ghostdagDataStore,
		daaBlocksStore:                          daaBlocksStore,
		sigCache:                                txscript.NewSigCache(sigCacheSize),
		sigCacheECDSA:                           txscript.NewSigCacheECDSA(sigCacheSize),
		txMassCalculator:                        txMassCalculator,
	}
}
