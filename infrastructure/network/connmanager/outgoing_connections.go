package connmanager

import "github.com/kaspanet/kaspad/app/appmessage"

// checkOutgoingConnections goes over all activeOutgoing and makes sure they are still active.
// Then it opens connections so that we have targetOutgoing active connections
func (c *ConnectionManager) checkOutgoingConnections(connSet connectionSet) {
	var connectedAddresses []*appmessage.NetAddress
	for address := range c.activeOutgoing {
		connection, ok := connSet.get(address)
		if ok { // connection is still connected
			connSet.remove(connection)
			connectedAddresses = append(connectedAddresses, connection.NetAddress())
			continue
		}

		// if connection is dead - remove from list of active ones
		delete(c.activeOutgoing, address)
	}

	liveConnections := len(c.activeOutgoing)
	if c.targetOutgoing == liveConnections {
		return
	}

	log.Debugf("Have got %d outgoing connections out of target %d, adding %d more",
		liveConnections, c.targetOutgoing, c.targetOutgoing-liveConnections)

	connectionsNeededCount := c.targetOutgoing - len(c.activeOutgoing)
	connectionAttempts := connectionsNeededCount * 2
	netAddresses := c.addressManager.RandomAddresses(connectionAttempts, connectedAddresses)

	for _, netAddress := range netAddresses {
		// Return in case we've already reached or surpassed our target
		if len(c.activeOutgoing) >= c.targetOutgoing {
			return
		}

		addressString := netAddress.TCPAddress().String()

		log.Debugf("Connecting to %s because we have %d outgoing connections and the target is "+
			"%d", addressString, len(c.activeOutgoing), c.targetOutgoing)

		err := c.initiateConnection(addressString)
		if err != nil {
			log.Infof("Couldn't connect to %s: %s", addressString, err)
			c.addressManager.RemoveAddress(netAddress)
			continue
		}

		c.activeOutgoing[addressString] = struct{}{}
	}
}
