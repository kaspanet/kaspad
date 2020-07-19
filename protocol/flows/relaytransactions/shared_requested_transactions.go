package relaytransactions

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"sync"
)

type SharedRequestedTransactions struct {
	transactions map[daghash.TxID]struct{}
	sync.Mutex
}

func (s *SharedRequestedTransactions) remove(txID *daghash.TxID) {
	s.Lock()
	defer s.Unlock()
	delete(s.transactions, *txID)
}

func (s *SharedRequestedTransactions) removeSet(txIDs map[daghash.TxID]struct{}) {
	s.Lock()
	defer s.Unlock()
	for txID := range txIDs {
		delete(s.transactions, txID)
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

// TODO(libp2p) move to manager scope
var requestedTransactions = &SharedRequestedTransactions{
	transactions: make(map[daghash.TxID]struct{}),
}
