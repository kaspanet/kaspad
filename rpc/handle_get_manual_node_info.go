package rpc

import (
	"net"

	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/pointers"
)

// handleGetManualNodeInfo handles getManualNodeInfo commands.
func handleGetManualNodeInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.GetManualNodeInfoCmd)
	results, err := getManualNodesInfo(s, c.Details, c.Node)
	if err != nil {
		return nil, err
	}
	if resultsNonDetailed, ok := results.([]string); ok {
		return resultsNonDetailed[0], nil
	}
	resultsDetailed := results.([]*rpcmodel.GetManualNodeInfoResult)
	return resultsDetailed[0], nil
}

// getManualNodesInfo handles getManualNodeInfo and getAllManualNodesInfo commands.
func getManualNodesInfo(s *Server, detailsArg *bool, node string) (interface{}, error) {

	details := detailsArg == nil || *detailsArg

	// Retrieve a list of persistent (manual) peers from the server and
	// filter the list of peers per the specified address (if any).
	peers := s.connectionManager.PersistentPeers()
	if node != "" {
		found := false
		for i, peer := range peers {
			if peer.ToPeer().Addr() == node {
				peers = peers[i : i+1]
				found = true
			}
		}
		if !found {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCClientNodeNotAdded,
				Message: "Node has not been added",
			}
		}
	}

	// Without the details flag, the result is just a slice of the addresses as
	// strings.
	if !details {
		results := make([]string, 0, len(peers))
		for _, peer := range peers {
			results = append(results, peer.ToPeer().Addr())
		}
		return results, nil
	}

	// With the details flag, the result is an array of JSON objects which
	// include the result of DNS lookups for each peer.
	results := make([]*rpcmodel.GetManualNodeInfoResult, 0, len(peers))
	for _, rpcPeer := range peers {
		// Set the "address" of the peer which could be an ip address
		// or a domain name.
		peer := rpcPeer.ToPeer()
		var result rpcmodel.GetManualNodeInfoResult
		result.ManualNode = peer.Addr()
		result.Connected = pointers.Bool(peer.Connected())

		// Split the address into host and port portions so we can do
		// a DNS lookup against the host. When no port is specified in
		// the address, just use the address as the host.
		host, _, err := net.SplitHostPort(peer.Addr())
		if err != nil {
			host = peer.Addr()
		}

		var ipList []string
		switch {
		case net.ParseIP(host) != nil:
			ipList = make([]string, 1)
			ipList[0] = host
		default:
			// Do a DNS lookup for the address. If the lookup fails, just
			// use the host.
			ips, err := s.cfg.Lookup(host)
			if err != nil {
				ipList = make([]string, 1)
				ipList[0] = host
				break
			}
			ipList = make([]string, 0, len(ips))
			for _, ip := range ips {
				ipList = append(ipList, ip.String())
			}
		}

		// Add the addresses and connection info to the result.
		addrs := make([]rpcmodel.GetManualNodeInfoResultAddr, 0, len(ipList))
		for _, ip := range ipList {
			var addr rpcmodel.GetManualNodeInfoResultAddr
			addr.Address = ip
			addr.Connected = "false"
			if ip == host && peer.Connected() {
				addr.Connected = logger.DirectionString(peer.Inbound())
			}
			addrs = append(addrs, addr)
		}
		result.Addresses = &addrs
		results = append(results, &result)
	}
	return results, nil
}
