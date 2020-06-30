package p2p

import (
	"fmt"
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

// OnAddr is invoked when a peer receives an addr kaspa message and is
// used to notify the server about advertised addresses.
func (sp *Peer) OnAddr(_ *peer.Peer, msg *wire.MsgAddr) {
	// Ignore addresses when running on the simulation test network. This
	// helps prevent the network from becoming another public test network
	// since it will not be able to learn about other peers that have not
	// specifically been provided.
	if config.ActiveConfig().Simnet {
		return
	}

	if len(msg.AddrList) > addrmgr.GetAddrMax {
		sp.AddBanScoreAndPushRejectMsg(msg.Command(), wire.RejectInvalid, nil,
			peer.BanScoreSentTooManyAddresses, 0, fmt.Sprintf("address count excceeded %d", addrmgr.GetAddrMax))
		return
	}

	if msg.IncludeAllSubnetworks {
		sp.AddBanScoreAndPushRejectMsg(msg.Command(), wire.RejectInvalid, nil,
			peer.BanScoreMsgAddrWithInvalidSubnetwork, 0,
			fmt.Sprintf("got unexpected IncludeAllSubnetworks=true in [%s] command", msg.Command()))
		return
	} else if !msg.SubnetworkID.IsEqual(config.ActiveConfig().SubnetworkID) && msg.SubnetworkID != nil {
		peerLog.Errorf("Only full nodes and %s subnetwork IDs are allowed in [%s] command, but got subnetwork ID %s from %s",
			config.ActiveConfig().SubnetworkID, msg.Command(), msg.SubnetworkID, sp.Peer)
		sp.Disconnect()
		return
	}

	for _, na := range msg.AddrList {
		// Don't add more address if we're disconnecting.
		if !sp.Connected() {
			return
		}

		// Set the timestamp to 5 days ago if it's more than 24 hours
		// in the future so this address is one of the first to be
		// removed when space is needed.
		now := mstime.Now()
		if na.Timestamp.After(now.Add(time.Minute * 10)) {
			na.Timestamp = now.Add(-1 * time.Hour * 24 * 5)
		}

		// Add address to known addresses for this peer.
		sp.addKnownAddresses([]*wire.NetAddress{na})
	}

	// Add addresses to server address manager. The address manager handles
	// the details of things such as preventing duplicate addresses, max
	// addresses, and last seen updates.
	sp.server.AddrManager.AddAddresses(msg.AddrList, sp.NA(), msg.SubnetworkID)
}
