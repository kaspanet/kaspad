package connmanager

// checkOutgoingConnections goes over all activeOutgoing and makes sure they are still active.
// Then it opens connections so that we have targetOutgoing active connections
func (c *ConnectionManager) checkOutgoingConnections(connSet connectionSet) {
	liveConnections := 0
	for address := range c.activeOutgoing {
		connection, ok := connSet.get(address)
		if ok { // connections still connected
			connSet.remove(connection)
			liveConnections++
			continue
		}

		// if connection is dead - remove from list of active ones
		delete(c.activeOutgoing, address)
	}

	connectionsToAdd := c.targetOutgoing - liveConnections
	if connectionsToAdd == 0 {
		return
	}

	log.Debugf("Have got %d outgoing connections out of target %d, adding %d more",
		liveConnections, c.targetOutgoing, connectionsToAdd)

	for i := 0; i < connectionsToAdd; i++ {
		address := c.addressManager.GetAddress()
		if address == nil {
			log.Debugf("No more addresses available")
			return
		}

		c.addressManager.Attempt(address.NetAddress())
		err := c.initiateConnection(address.NetAddress().TCPAddress().String())
		if err != nil {
			i--
		} else {
			c.addressManager.Connected(address.NetAddress())
		}
	}
}
