package blockdag

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// Config is a descriptor which specifies the blockDAG instance configuration.
type Config struct {
	// DAGParams identifies which DAG parameters the DAG is associated
	// with.
	//
	// This field is required.
	DAGParams *dagconfig.Params

	// TimeSource defines the time source to use for things such as
	// block processing and determining whether or not the DAG is current.
	TimeSource TimeSource

	// SigCache defines a signature cache to use when when validating
	// signatures. This is typically most useful when individual
	// transactions are already being validated prior to their inclusion in
	// a block such as what is usually done via a transaction memory pool.
	//
	// This field can be nil if the caller is not interested in using a
	// signature cache.
	SigCache *txscript.SigCache

	// IndexManager defines an Index manager to use when initializing the
	// DAG and connecting blocks.
	//
	// This field can be nil if the caller does not wish to make use of an
	// Index manager.
	IndexManager IndexManager

	// SubnetworkID identifies which subnetwork the DAG is associated
	// with.
	//
	// This field is required.
	SubnetworkID *subnetworkid.SubnetworkID

	// DatabaseContext is the context in which all database queries related to
	// this DAG are going to run.
	DatabaseContext *dbaccess.DatabaseContext

	// MaxUTXOCacheSize is the Max size of loaded UTXO into ram from the disk in bytes
	// to support UTXO lazy-load
	MaxUTXOCacheSize uint64
}
