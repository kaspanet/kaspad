package connmanager

import "github.com/kaspanet/kaspad/app/appmessage"

// checkOutgoingConnections goes over all activeOutgoing and makes sure they are still active.
// Then it opens connections so that we have targetOutgoing active connections
func (c *ConnectionManager) checkOutgoingConnections(connSet connectionSet) {
	for address := range c.activeOutgoing {
		connection, ok := connSet.get(address)
		if ok { // connection is still connected
			connSet.remove(connection)
			continue
		}

		// if connection is dead - remove from list of active ones
		delete(c.activeOutgoing, address)
	}

	connections := c.netAdapter.P2PConnections()
	connectedAddresses := make([]*appmessage.NetAddress, len(connections))
	for i, connection := range connections {
		connectedAddresses[i] = connection.NetAddress()
	}

	liveConnections := len(c.activeOutgoing)
	if c.targetOutgoing == liveConnections {
		return
	}

	log.Tracef("Have got %d outgoing connections out of target %d, adding %d more",
		liveConnections, c.targetOutgoing, c.targetOutgoing-liveConnections)

	connectionsNeededCount := c.targetOutgoing - len(c.activeOutgoing)
	netAddresses := c.addressManager.RandomAddresses(connectionsNeededCount, connectedAddresses)

	for _, netAddress := range netAddresses {
		addressString := netAddress.TCPAddress().String()

		log.Tracef("Connecting to %s because we have %d outgoing connections and the target is "+
			"%d", addressString, len(c.activeOutgoing), c.targetOutgoing)

		err := c.initiateConnection(addressString)
		if err != nil {
			log.Tracef("Couldn't connect to %s: %s", addressString, err)
			c.addressManager.MarkConnectionFailure(netAddress)
			continue
		}
		c.addressManager.MarkConnectionSuccess(netAddress)

		c.activeOutgoing[addressString] = struct{}{}
	}

	if len(netAddresses) < connectionsNeededCount {
		log.Tracef("Need %d more outgoing connections - seeding addresses from DNS",
			connectionsNeededCount-len(netAddresses))

		// seedFromDNS is an asynchronous method, therefore addresses for connection
		// should be available on next iteration
		c.seedFromDNS()
	}
}
