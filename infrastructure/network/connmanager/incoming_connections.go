package connmanager

// checkIncomingConnections makes sure there's no more than maxIncoming incoming connections
// if there are - it randomly disconnects enough to go below that number
func (c *ConnectionManager) checkIncomingConnections(incomingConnectionSet connectionSet) {
	if len(incomingConnectionSet) <= c.maxIncoming {
		return
	}

	numConnectionsOverMax := len(incomingConnectionSet) - c.maxIncoming
	log.Tracef("Got %d incoming connections while only %d are allowed. Disconnecting "+
		"%d", len(incomingConnectionSet), c.maxIncoming, numConnectionsOverMax)

	// randomly disconnect nodes until the number of incoming connections is smaller than maxIncoming
	for _, connection := range incomingConnectionSet {
		log.Tracef("Disconnecting %s due to exceeding incoming connections", connection)
		connection.Disconnect()

		numConnectionsOverMax--
		if numConnectionsOverMax == 0 {
			break
		}
	}
}
