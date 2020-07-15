package connmanager

import (
	"time"
)

const (
	minRetryDuration = 30 * time.Second
	maxRetryDuration = 10 * time.Minute
)

func nextRetryDuration(previousDuration time.Duration) time.Duration {
	if previousDuration == 0 {
		return minRetryDuration
	}
	if previousDuration*2 > maxRetryDuration {
		return maxRetryDuration
	}
	return previousDuration * 2
}

// checkConnectionRequests checks that all activeConnectionRequests are still active, and initiates connections
// for pendingConnectionRequests.
// While doing so, it filters out of connSet all connections that were initiated as a connectionRequest
func (c *ConnectionManager) checkConnectionRequests(connSet connectionSet) {
	c.connectionRequestsLock.Lock()
	defer c.connectionRequestsLock.Unlock()

	now := time.Now()

	for address, connReq := range c.activeConnectionRequests {
		connection, ok := connSet.get(address)
		if !ok { // a requested connection was disconnected
			delete(c.activeConnectionRequests, address)

			if connReq.isPermanent { // if is one-try - ignore. If permanent - add to pending list to retry
				connReq.nextAttempt = now
				connReq.retryDuration = time.Second
				c.pendingConnectionRequests[address] = connReq
			}
			continue
		}

		connSet.remove(connection)
	}

	for address, connReq := range c.pendingConnectionRequests {
		if connReq.nextAttempt.After(now) { // ignore connection requests which are still waiting for retry
			continue
		}

		connection, ok := connSet.get(address)
		if ok { // somehow the pendingConnectionRequest has already connected - move it to active
			delete(c.pendingConnectionRequests, address)
			c.pendingConnectionRequests[address] = connReq

			connSet.remove(connection)

			continue
		}

		// try to initiate connection
		err := c.initiateConnection(connReq.address)

		if err == nil { // if connected successfully - move from pending to active
			delete(c.pendingConnectionRequests, address)
			c.activeConnectionRequests[address] = connReq
			continue
		}
		if !connReq.isPermanent { // if connection request is one try - remove from pending and ignore failure
			delete(c.pendingConnectionRequests, address)
			continue
		}
		// if connection request is permanent - keep in pending, and increase retry time
		connReq.retryDuration = nextRetryDuration(connReq.retryDuration)
		connReq.nextAttempt = now.Add(connReq.retryDuration)
		log.Debugf("Retrying connection to %s in %s", address, connReq.retryDuration)

	}
}

// AddConnectionRequest adds the given address to list of pending connection requests
func (c *ConnectionManager) AddConnectionRequest(address string, isPermanent bool) {
	// spawn goroutine so that caller doesn't wait in case connectionManager is in the midst of handling
	// connection requests
	spawn(func() {
		c.connectionRequestsLock.Lock()
		defer c.connectionRequestsLock.Unlock()

		if _, ok := c.activeConnectionRequests[address]; ok {
			return
		}

		c.pendingConnectionRequests[address] = &connectionRequest{
			address:     address,
			isPermanent: isPermanent,
		}
	})
}
