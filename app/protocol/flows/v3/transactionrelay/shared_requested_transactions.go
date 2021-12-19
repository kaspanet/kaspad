package transactionrelay

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

func (s *SharedRequestedTransactions) remove(txID *externalapi.DomainTransactionID) {
	s.Lock()
	defer s.Unlock()
	delete(s.transactions, *txID)
}

func (s *SharedRequestedTransactions) removeMany(txIDs []*externalapi.DomainTransactionID) {
	s.Lock()
	defer s.Unlock()
	for _, txID := range txIDs {
		delete(s.transactions, *txID)
	}
}

func (s *SharedRequestedTransactions) addIfNotExists(txID *externalapi.DomainTransactionID) (exists bool) {
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
