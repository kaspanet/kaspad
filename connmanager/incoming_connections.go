package connmanager

// checkIncomingConnections makes sure there's no more then maxIncoming incoming connections
// if there are - it randomly disconnects enough to go below that number
func (c *ConnectionManager) checkIncomingConnections(connSet connectionSet) {
	if len(connSet) <= c.maxIncoming {
		return
	}

	numConnectionsOverMax := len(connSet) - c.maxIncoming
	// randomly disconnect nodes until the number of incoming connections is smaller the maxIncoming
	for address, connection := range connSet {
		err := c.netAdapter.Disconnect(connection)
		if err != nil {
			log.Errorf("Error disconnecting from %s: %+v", address, err)
		}

		numConnectionsOverMax--
		if numConnectionsOverMax == 0 {
			break
		}
	}
}
