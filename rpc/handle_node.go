package rpc

import (
	"github.com/kaspanet/kaspad/connmanager"
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/network"
	"net"
	"strconv"
)

// handleNode handles node commands.
func handleNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.NodeCmd)

	var addr string
	var nodeID uint64
	var errN, err error
	params := s.dag.Params
	switch c.SubCmd {
	case "disconnect":
		// If we have a valid uint disconnect by node id. Otherwise,
		// attempt to disconnect by address, returning an error if a
		// valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.connectionManager.DisconnectByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr, err = network.NormalizeAddress(c.Target, params.DefaultPort)
				if err != nil {
					break
				}

				err = s.connectionManager.DisconnectByAddr(addr)
			} else {
				return nil, &model.RPCError{
					Code:    model.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.connectionManager, addr, int32(nodeID)) {

			return nil, &model.RPCError{
				Code:    model.ErrRPCMisc,
				Message: "can't disconnect a permanent peer, use remove",
			}
		}

	case "remove":
		// If we have a valid uint disconnect by node id. Otherwise,
		// attempt to disconnect by address, returning an error if a
		// valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.connectionManager.RemoveByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr, err = network.NormalizeAddress(c.Target, params.DefaultPort)
				if err != nil {
					break
				}

				err = s.connectionManager.RemoveByAddr(addr)
			} else {
				return nil, &model.RPCError{
					Code:    model.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.connectionManager, addr, int32(nodeID)) {
			return nil, &model.RPCError{
				Code:    model.ErrRPCMisc,
				Message: "can't remove a temporary peer, use disconnect",
			}
		}

	case "connect":
		addr, err = network.NormalizeAddress(c.Target, params.DefaultPort)
		if err != nil {
			break
		}

		// Default to temporary connections.
		subCmd := "temp"
		if c.ConnectSubCmd != nil {
			subCmd = *c.ConnectSubCmd
		}

		switch subCmd {
		case "perm", "temp":
			s.connectionManager.AddConnectionRequest(addr, subCmd == "perm")
		default:
			return nil, &model.RPCError{
				Code:    model.ErrRPCInvalidParameter,
				Message: "invalid subcommand for node connect",
			}
		}
	default:
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: "invalid subcommand for node",
		}
	}

	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}

// peerExists determines if a certain peer is currently connected given
// information about all currently connected peers. Peer existence is
// determined using either a target address or node id.
func peerExists(connectionManager *connmanager.ConnectionManager, addr string, nodeID int32) bool {
	for _, p := range connectionManager.ConnectedPeers() {
		if p.ToPeer().ID() == nodeID || p.ToPeer().Addr() == addr {
			return true
		}
	}
	return false
}
