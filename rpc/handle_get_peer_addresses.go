package rpc

import "github.com/kaspanet/kaspad/rpc/model"

// handleGetPeerAddresses handles getPeerAddresses commands.
func handleGetPeerAddresses(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peersState, err := s.addressManager.PeersStateForSerialization()
	if err != nil {
		return nil, err
	}

	rpcPeersState := model.GetPeerAddressesResult{
		Version:              peersState.Version,
		Key:                  peersState.Key,
		Addresses:            make([]*model.GetPeerAddressesKnownAddressResult, len(peersState.Addresses)),
		NewBuckets:           make(map[string]*model.GetPeerAddressesNewBucketResult),
		NewBucketFullNodes:   model.GetPeerAddressesNewBucketResult{},
		TriedBuckets:         make(map[string]*model.GetPeerAddressesTriedBucketResult),
		TriedBucketFullNodes: model.GetPeerAddressesTriedBucketResult{},
	}

	for i, addr := range peersState.Addresses {
		rpcPeersState.Addresses[i] = &model.GetPeerAddressesKnownAddressResult{
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
		rpcPeersState.NewBuckets[subnetworkID] = &model.GetPeerAddressesNewBucketResult{}
		for i, addr := range bucket {
			rpcPeersState.NewBuckets[subnetworkID][i] = addr
		}
	}

	for i, addr := range peersState.NewBucketFullNodes {
		rpcPeersState.NewBucketFullNodes[i] = addr
	}

	for subnetworkID, bucket := range peersState.TriedBuckets {
		rpcPeersState.TriedBuckets[subnetworkID] = &model.GetPeerAddressesTriedBucketResult{}
		for i, addr := range bucket {
			rpcPeersState.TriedBuckets[subnetworkID][i] = addr
		}
	}

	for i, addr := range peersState.TriedBucketFullNodes {
		rpcPeersState.TriedBucketFullNodes[i] = addr
	}

	return rpcPeersState, nil
}
