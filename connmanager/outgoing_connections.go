package connmanager

// checkOutgoingConnections goes over all activeOutgoing and makes sure they are still active.
// Then it opens connections so that we have targetOutgoing active connections
func (c *ConnectionManager) checkOutgoingConnections(connSet connectionSet) {
	for address := range c.activeOutgoing {
		connection, ok := connSet.get(address)
		if ok { // connections still connected
			connSet.remove(connection)
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

	for len(c.activeOutgoing) < c.targetOutgoing {
		address := c.addressManager.GetAddress()
		if address == nil {
			log.Warnf("No more addresses available")
			return
		}

		netAddress := address.NetAddress()
		c.addressManager.Attempt(netAddress)
		addressString := netAddress.TCPAddress().String()
		err := c.initiateConnection(addressString)
		if err != nil {
			log.Infof("Couldn't connect to %s: %s", addressString, err)
			continue
		}

		c.addressManager.Connected(address.NetAddress())
		c.activeOutgoing[address.NetAddress().TCPAddress().String()] = struct{}{}
	}
}
