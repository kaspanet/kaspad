package flowcontext

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// SharedRequestedTransactions is a data structure that is shared between peers that
// holds the IDs of all the requested transactions to prevent redundant requests.
type SharedRequestedTransactions struct {
	transactions map[externalapi.DomainTransactionID]struct{}
	sync.Mutex
}

// Remove removes a transaction from the set.
func (s *SharedRequestedTransactions) Remove(txID *externalapi.DomainTransactionID) {
	s.Lock()
	defer s.Unlock()
	delete(s.transactions, *txID)
}

// RemoveMany removes a set of transactions from the set.
func (s *SharedRequestedTransactions) RemoveMany(txIDs []*externalapi.DomainTransactionID) {
	s.Lock()
	defer s.Unlock()
	for _, txID := range txIDs {
		delete(s.transactions, *txID)
	}
}

// AddIfNotExists adds a transaction to the set if it doesn't exist yet.
func (s *SharedRequestedTransactions) AddIfNotExists(txID *externalapi.DomainTransactionID) (exists bool) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.transactions[*txID]
	if ok {
		return true
	}
	s.transactions[*txID] = struct{}{}
	return false
}

// NewSharedRequestedTransactions returns a new instance of SharedRequestedTransactions.
func NewSharedRequestedTransactions() *SharedRequestedTransactions {
	return &SharedRequestedTransactions{
		transactions: make(map[externalapi.DomainTransactionID]struct{}),
	}
}
