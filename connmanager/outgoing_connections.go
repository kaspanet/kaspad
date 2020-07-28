package connmanager

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

	liveConnections := len(c.activeOutgoing)
	if c.targetOutgoing == liveConnections {
		return
	}

	log.Debugf("Have got %d outgoing connections out of target %d, adding %d more",
		liveConnections, c.targetOutgoing, c.targetOutgoing-liveConnections)

	connectionsNeededCount := c.targetOutgoing - len(c.activeOutgoing)
	connectionAttempts := connectionsNeededCount * 2
	for i := 0; i < connectionAttempts; i++ {
		// Return in case we've already reached or surpassed our target
		if len(c.activeOutgoing) >= c.targetOutgoing {
			return
		}

		address := c.addressManager.GetAddress()
		if address == nil {
			log.Warnf("No more addresses available")
			return
		}

		netAddress := address.NetAddress()
		tcpAddress := netAddress.TCPAddress()
		addressString := tcpAddress.String()
		isBanned, err := c.addressManager.IsBanned(netAddress)
		if err != nil {
			log.Infof("Couldn't resolve whether %s is banned: %s", addressString, err)
			continue
		}
		if isBanned {
			continue
		}

		c.addressManager.Attempt(netAddress)
		err = c.initiateConnection(addressString)
		if err != nil {
			log.Infof("Couldn't connect to %s: %s", addressString, err)
			continue
		}

		c.addressManager.Connected(netAddress)
		c.activeOutgoing[addressString] = struct{}{}
	}
}
