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

// checkRequestedConnections checks that all activeRequested are still active, and initiates connections
// for pendingRequested.
// While doing so, it filters out of connSet all connections that were initiated as a connectionRequest
func (c *ConnectionManager) checkRequestedConnections(connSet connectionSet) {
	c.connectionRequestsLock.Lock()
	defer c.connectionRequestsLock.Unlock()

	now := time.Now()

	for address, connReq := range c.activeRequested {
		connection, ok := connSet.get(address)
		if !ok { // a requested connection was disconnected
			delete(c.activeRequested, address)

			if connReq.isPermanent { // if is one-try - ignore. If permanent - add to pending list to retry
				connReq.nextAttempt = now
				connReq.retryDuration = time.Second
				c.pendingRequested[address] = connReq
			}
			continue
		}

		connSet.remove(connection)
	}

	for address, connReq := range c.pendingRequested {
		if connReq.nextAttempt.After(now) { // ignore connection requests which are still waiting for retry
			continue
		}

		connection, ok := connSet.get(address)
		// The pendingConnectionRequest has already connected - move it to active
		// This can happen in rare cases such as when the other side has connected to our node
		// while it has been pending on our side.
		if ok {
			delete(c.pendingRequested, address)
			c.pendingRequested[address] = connReq

			connSet.remove(connection)

			continue
		}

		// try to initiate connection
		err := c.initiateConnection(connReq.address)

		if err == nil { // if connected successfully - move from pending to active
			delete(c.pendingRequested, address)
			c.activeRequested[address] = connReq
			continue
		}
		if !connReq.isPermanent { // if connection request is one try - remove from pending and ignore failure
			delete(c.pendingRequested, address)
			continue
		}
		// if connection request is permanent - keep in pending, and increase retry time
		connReq.retryDuration = nextRetryDuration(connReq.retryDuration)
		connReq.nextAttempt = now.Add(connReq.retryDuration)
		log.Debugf("Retrying permanent connection to %s in %s", address, connReq.retryDuration)
	}
}

// AddConnectionRequest adds the given address to list of pending connection requests
func (c *ConnectionManager) AddConnectionRequest(address string, isPermanent bool) {
	// spawn goroutine so that caller doesn't wait in case connectionManager is in the midst of handling
	// connection requests
	spawn(func() {
		c.connectionRequestsLock.Lock()
		defer c.connectionRequestsLock.Unlock()

		if _, ok := c.activeRequested[address]; ok {
			return
		}

		c.pendingRequested[address] = &connectionRequest{
			address:     address,
			isPermanent: isPermanent,
		}
	})
}
