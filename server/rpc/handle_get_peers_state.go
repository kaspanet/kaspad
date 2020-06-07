package rpc

import "github.com/kaspanet/kaspad/rpcmodel"

// handleGetPeersState handles getPeersState commands.
func handleGetPeersState(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peersState, err := s.cfg.addressManager.PeersStateForSerialization()
	if err != nil {
		return nil, err
	}

	rpcPeersState := rpcmodel.GetPeersStateResult{
		Version:              peersState.Version,
		Key:                  peersState.Key,
		Addresses:            make([]*rpcmodel.PeersStateKnownAddressResult, len(peersState.Addresses)),
		NewBuckets:           make(map[string]*rpcmodel.PeersStateNewBucketResult),
		NewBucketFullNodes:   rpcmodel.PeersStateNewBucketResult{},
		TriedBuckets:         make(map[string]*rpcmodel.PeersStateTriedBucketResult),
		TriedBucketFullNodes: rpcmodel.PeersStateTriedBucketResult{},
	}

	for i, addr := range peersState.Addresses {
		rpcPeersState.Addresses[i] = &rpcmodel.PeersStateKnownAddressResult{
			Addr:         addr.Addr,
			Src:          addr.Src,
			SubnetworkID: addr.SubnetworkID,
			Attempts:     addr.Attempts,
			TimeStamp:    addr.TimeStamp,
			LastAttempt:  addr.LastAttempt,
			LastSuccess:  addr.LastSuccess,
		}
	}

	for subnetworkID, bucket := range peersState.NewBuckets {
		rpcPeersState.NewBuckets[subnetworkID] = &rpcmodel.PeersStateNewBucketResult{}
		for i, addr := range bucket {
			rpcPeersState.NewBuckets[subnetworkID][i] = addr
		}
	}

	for i, addr := range peersState.NewBucketFullNodes {
		rpcPeersState.NewBucketFullNodes[i] = addr
	}

	for subnetworkID, bucket := range peersState.TriedBuckets {
		rpcPeersState.TriedBuckets[subnetworkID] = &rpcmodel.PeersStateTriedBucketResult{}
		for i, addr := range bucket {
			rpcPeersState.TriedBuckets[subnetworkID][i] = addr
		}
	}

	for i, addr := range peersState.TriedBucketFullNodes {
		rpcPeersState.TriedBucketFullNodes[i] = addr
	}

	return rpcPeersState, nil
}
