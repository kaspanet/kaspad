package txindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TxBlockData represents 
type TxBlockData struct {
	acceptingBlockHash *externalapi.DomainHash
	mergeBlockHash *externalapi.DomainHash
}

// TxAcceptingChanges is the set of changes made to the tx index after
// a successful update
type TxAcceptingChanges struct {
	toAddAccepting   map[externalapi.DomainHash][]*externalapi.DomainHash
	toRemoveAccepting   map[externalapi.DomainHash][]*externalapi.DomainHash
}

// TxAcceptingChanges is the set of changes made to the tx index after
// a successful update
type TxMergingChanges struct {
	toAddMerge  map[externalapi.DomainHash][]*externalapi.DomainHash
	toRemoveMeroge   map[externalapi.DomainHash][]*externalapi.DomainHash
}

