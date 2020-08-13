package connmanager

// checkIncomingConnections makes sure there's no more than maxIncoming incoming connections
// if there are - it randomly disconnects enough to go below that number
func (c *ConnectionManager) checkIncomingConnections(incomingConnectionSet connectionSet) {
	if len(incomingConnectionSet) <= c.maxIncoming {
		return
	}

	numConnectionsOverMax := len(incomingConnectionSet) - c.maxIncoming
	// randomly disconnect nodes until the number of incoming connections is smaller than maxIncoming
	for _, connection := range incomingConnectionSet {
		connection.Disconnect()

		numConnectionsOverMax--
		if numConnectionsOverMax == 0 {
			break
		}
	}
}
