package relaytransactions

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"sync"
)

// SharedRequestedTransactions is a data structure that is shared between peers that
// holds the IDs of all the requested transactions to prevent redundant requests.
type SharedRequestedTransactions struct {
	transactions map[daghash.TxID]struct{}
	sync.Mutex
}

func (s *SharedRequestedTransactions) remove(txID *daghash.TxID) {
	s.Lock()
	defer s.Unlock()
	delete(s.transactions, *txID)
}

func (s *SharedRequestedTransactions) removeMany(txIDs []*daghash.TxID) {
	s.Lock()
	defer s.Unlock()
	for _, txID := range txIDs {
		delete(s.transactions, *txID)
	}
}

func (s *SharedRequestedTransactions) addIfNotExists(txID *daghash.TxID) (exists bool) {
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
		transactions: make(map[daghash.TxID]struct{}),
	}
}
