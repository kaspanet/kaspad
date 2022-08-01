package utxoindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ScriptPublicKeyString is a script public key represented as a string
// We use this type rather than just a byte slice because Go maps don't
// support slices as keys. See: UTXOChanges
type ScriptPublicKeyString string

// UTXOOutpointEntryPairs is a map between UTXO outpoints to UTXO entries
type UTXOOutpointEntryPairs map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// UTXOOutpoints is a set of UTXO outpoints
type UTXOOutpoints map[externalapi.DomainOutpoint]interface{}

// UTXOChanges is the set of changes made to the UTXO index after
// a successful update
type UTXOChanges struct {
	Added   map[ScriptPublicKeyString]UTXOOutpointEntryPairs
	Removed map[ScriptPublicKeyString]UTXOOutpointEntryPairs
}
