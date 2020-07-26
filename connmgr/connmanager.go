// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	nativeerrors "errors"
	"fmt"
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/wire"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// maxFailedAttempts is the maximum number of successive failed connection
// attempts after which network failure is assumed and new connections will
// be delayed by the configured retry duration.
const maxFailedAttempts = 25

var (
	// maxRetryDuration is the max duration of time retrying of a persistent
	// connection is allowed to grow to. This is necessary since the retry
	// logic uses a backoff mechanism which increases the interval base times
	// the number of retries that have been done.
	maxRetryDuration = time.Minute * 5

	// defaultRetryDuration is the default duration of time for retrying
	// persistent connections.
	defaultRetryDuration = time.Second * 5
)

var (
	//ErrDialNil is used to indicate that Dial cannot be nil in the configuration.
	ErrDialNil = errors.New("Config: Dial cannot be nil")

	// ErrMaxOutboundPeers is an error that is thrown when the max amount of peers had
	// been reached.
	ErrMaxOutboundPeers = errors.New("max outbound peers reached")

	// ErrAlreadyConnected is an error that is thrown if the peer is already
	// connected.
	ErrAlreadyConnected = errors.New("peer already connected")

	// ErrAlreadyPermanent is an error that is thrown if the peer is already
	// connected as a permanent peer.
	ErrAlreadyPermanent = errors.New("peer exists as a permanent peer")

	// ErrPeerNotFound is an error that is thrown if the peer was not found.
	ErrPeerNotFound = errors.New("peer not found")

	//ErrAddressManagerNil is used to indicate that Address Manager cannot be nil in the configuration.
	ErrAddressManagerNil = errors.New("Config: Address manager cannot be nil")
)

// ConnState represents the state of the requested connection.
type ConnState uint8

// ConnState can be either pending, established, disconnected or failed. When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result. An established
// connection which was disconnected is categorized as disconnected.
const (
	ConnPending ConnState = iota
	ConnFailing
	ConnCanceled
	ConnEstablished
	ConnDisconnected
)

// ConnReq is the connection request to a network address. If permanent, the
// connection will be retried on disconnection.
type ConnReq struct {
	// The following variables must only be used atomically.
	id uint64

	Addr      *net.TCPAddr
	Permanent bool

	conn       net.Conn
	state      ConnState
	stateMtx   sync.RWMutex
	retryCount uint32
}

// updateState updates the state of the connection request.
func (c *ConnReq) updateState(state ConnState) {
	c.stateMtx.Lock()
	defer c.stateMtx.Unlock()
	c.state = state
}

// ID returns a unique identifier for the connection request.
func (c *ConnReq) ID() uint64 {
	return atomic.LoadUint64(&c.id)
}

// State is the connection state of the requested connection.
func (c *ConnReq) State() ConnState {
	c.stateMtx.RLock()
	defer c.stateMtx.RUnlock()
	state := c.state
	return state
}

// String returns a human-readable string for the connection request.
func (c *ConnReq) String() string {
	if c.Addr == nil || c.Addr.String() == "" {
		return fmt.Sprintf("reqid %d", atomic.LoadUint64(&c.id))
	}
	return fmt.Sprintf("%s (reqid %d)", c.Addr, atomic.LoadUint64(&c.id))
}

// Config holds the configuration options related to the connection manager.
type Config struct {
	// Listeners defines a slice of listeners for which the connection
	// manager will take ownership of and accept connections. When a
	// connection is accepted, the OnAccept handler will be invoked with the
	// connection. Since the connection manager takes ownership of these
	// listeners, they will be closed when the connection manager is
	// stopped.
	//
	// This field will not have any effect if the OnAccept field is not
	// also specified. It may be nil if the caller does not wish to listen
	// for incoming connections.
	Listeners []net.Listener

	// OnAccept is a callback that is fired when an inbound connection is
	// accepted. It is the caller's responsibility to close the connection.
	// Failure to close the connection will result in the connection manager
	// believing the connection is still active and thus have undesirable
	// side effects such as still counting toward maximum connection limits.
	//
	// This field will not have any effect if the Listeners field is not
	// also specified since there couldn't possibly be any accepted
	// connections in that case.
	OnAccept func(net.Conn)

	// TargetOutbound is the number of outbound network connections to
	// maintain. Defaults to 8.
	TargetOutbound uint32

	// RetryDuration is the duration to wait before retrying connection
	// requests. Defaults to 5s.
	RetryDuration time.Duration

	// OnConnection is a callback that is fired when a new outbound
	// connection is established.
	OnConnection func(*ConnReq, net.Conn)

	// OnConnectionFailed is a callback that is fired when a new outbound
	// connection has failed to be established.
	OnConnectionFailed func(*ConnReq)

	// OnDisconnection is a callback that is fired when an outbound
	// connection is disconnected.
	OnDisconnection func(*ConnReq)

	AddrManager *addrmgr.AddrManager

	// Dial connects to the address on the named network. It cannot be nil.
	Dial func(net.Addr) (net.Conn, error)
}

// registerPending is used to register a pending connection attempt. By
// registering pending connection attempts we allow callers to cancel pending
// connection attempts before their successful or in the case they're not
// longer wanted.
type registerPending struct {
	c    *ConnReq
	done chan struct{}
}

// handleConnected is used to queue a successful connection.
type handleConnected struct {
	c    *ConnReq
	conn net.Conn
}

// handleDisconnected is used to remove a connection.
type handleDisconnected struct {
	id    uint64
	retry bool
}

// handleFailed is used to remove a pending connection.
type handleFailed struct {
	c   *ConnReq
	err error
}

// ConnManager provides a manager to handle network connections.
type ConnManager struct {
	// The following variables must only be used atomically.
	connReqCount uint64
	start        int32
	stop         int32

	addressMtx         sync.Mutex
	usedOutboundGroups map[string]int64
	usedAddresses      map[string]struct{}

	cfg            Config
	wg             sync.WaitGroup
	failedAttempts uint64
	requests       chan interface{}
	quit           chan struct{}
}

// handleFailedConn handles a connection failed due to a disconnect or any
// other failure. If permanent, it retries the connection after the configured
// retry duration. Otherwise, if required, it makes a new connection request.
// After maxFailedConnectionAttempts new connections will be retried after the
// configured retry duration.
func (cm *ConnManager) handleFailedConn(c *ConnReq, err error) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	// Don't write throttled logs more than once every throttledConnFailedLogInterval
	shouldWriteLog := shouldWriteConnFailedLog(err)
	if shouldWriteLog {
		// If we are to write a log, set its lastLogTime to now
		setConnFailedLastLogTime(err, time.Now())
	}

	if c.Permanent {
		c.retryCount++
		d := time.Duration(c.retryCount) * cm.cfg.RetryDuration
		if d > maxRetryDuration {
			d = maxRetryDuration
		}
		if shouldWriteLog {
			log.Debugf("Retrying further connections to %s every %s", c, d)
		}
		spawnAfter(d, func() {
			cm.connect(c)
		})
	} else {
		if c.Addr != nil {
			cm.releaseAddress(c.Addr)
		}
		cm.failedAttempts++
		if cm.failedAttempts >= maxFailedAttempts {
			if shouldWriteLog {
				log.Debugf("Max failed connection attempts reached: [%d] "+
					"-- retrying further connections every %s", maxFailedAttempts,
					cm.cfg.RetryDuration)
			}
			spawnAfter(cm.cfg.RetryDuration, cm.NewConnReq)
		} else {
			spawn(cm.NewConnReq)
		}
	}
}

func (cm *ConnManager) releaseAddress(addr *net.TCPAddr) {
	cm.addressMtx.Lock()
	defer cm.addressMtx.Unlock()

	groupKey := usedOutboundGroupsKey(addr)
	cm.usedOutboundGroups[groupKey]--
	if cm.usedOutboundGroups[groupKey] < 0 {
		panic(fmt.Errorf("cm.usedOutboundGroups[%s] has a negative value of %d. This should never happen", groupKey, cm.usedOutboundGroups[groupKey]))
	}
	delete(cm.usedAddresses, usedAddressesKey(addr))
}

func (cm *ConnManager) markAddressAsUsed(addr *net.TCPAddr) {
	cm.usedOutboundGroups[usedOutboundGroupsKey(addr)]++
	cm.usedAddresses[usedAddressesKey(addr)] = struct{}{}
}

func (cm *ConnManager) isOutboundGroupUsed(addr *net.TCPAddr) bool {
	_, ok := cm.usedOutboundGroups[usedOutboundGroupsKey(addr)]
	return ok
}

func (cm *ConnManager) isAddressUsed(addr *net.TCPAddr) bool {
	_, ok := cm.usedAddresses[usedAddressesKey(addr)]
	return ok
}

func usedOutboundGroupsKey(addr *net.TCPAddr) string {
	// A fake service flag is used since it doesn't affect the group key.
	na := wire.NewNetAddress(addr, wire.SFNodeNetwork)
	return addrmgr.GroupKey(na)
}

func usedAddressesKey(addr *net.TCPAddr) string {
	return addr.String()
}

// throttledError defines an error type whose logs get throttled. This is to
// prevent flooding the logs with identical errors.
type throttledError error

var (
	// throttledConnFailedLogInterval is the minimum duration of time between
	// the logs defined in throttledConnFailedLogs.
	throttledConnFailedLogInterval = time.Minute * 10

	// throttledConnFailedLogs are logs that get written at most every
	// throttledConnFailedLogInterval. Each entry in this map defines a type
	// of error that we want to throttle. The value of each entry is the last
	// time that type of log had been written.
	throttledConnFailedLogs = map[throttledError]time.Time{
		ErrNoAddress: {},
	}

	// ErrNoAddress is an error that is thrown when there aren't any
	// valid connection addresses.
	ErrNoAddress throttledError = errors.New("no valid connect address")
)

// shouldWriteConnFailedLog resolves whether to write logs related to connection
// failures. Errors that had not been previously registered in throttledConnFailedLogs
// and non-error (nil values) must always be logged.
func shouldWriteConnFailedLog(err error) bool {
	if err == nil {
		return true
	}
	lastLogTime, ok := throttledConnFailedLogs[err]
	return !ok || lastLogTime.Add(throttledConnFailedLogInterval).Before(time.Now())
}

// setConnFailedLastLogTime sets the last log time of the specified error
func setConnFailedLastLogTime(err error, lastLogTime time.Time) {
	var throttledErr throttledError
	nativeerrors.As(err, &throttledErr)
	throttledConnFailedLogs[err] = lastLogTime
}

// connHandler handles all connection related requests. It must be run as a
// goroutine.
//
// The connection handler makes sure that we maintain a pool of active outbound
// connections so that we remain connected to the network. Connection requests
// are processed and mapped by their assigned ids.
func (cm *ConnManager) connHandler() {

	var (
		// pending holds all registered conn requests that have yet to
		// succeed.
		pending = make(map[uint64]*ConnReq)

		// conns represents the set of all actively connected peers.
		conns = make(map[uint64]*ConnReq, cm.cfg.TargetOutbound)
	)

out:
	for {
		select {
		case req := <-cm.requests:
			switch msg := req.(type) {

			case registerPending:
				connReq := msg.c
				connReq.updateState(ConnPending)
				pending[msg.c.id] = connReq
				close(msg.done)

			case handleConnected:
				connReq := msg.c

				if _, ok := pending[connReq.id]; !ok {
					if msg.conn != nil {
						msg.conn.Close()
					}
					log.Debugf("Ignoring connection for "+
						"canceled connreq=%s", connReq)
					continue
				}

				connReq.updateState(ConnEstablished)
				connReq.conn = msg.conn
				conns[connReq.id] = connReq
				log.Debugf("Connected to %s", connReq)
				connReq.retryCount = 0

				delete(pending, connReq.id)

				if cm.cfg.OnConnection != nil {
					cm.cfg.OnConnection(connReq, msg.conn)
				}

			case handleDisconnected:
				connReq, ok := conns[msg.id]
				if !ok {
					connReq, ok = pending[msg.id]
					if !ok {
						log.Errorf("Unknown connid=%d",
							msg.id)
						continue
					}

					// Pending connection was found, remove
					// it from pending map if we should
					// ignore a later, successful
					// connection.
					connReq.updateState(ConnCanceled)
					log.Debugf("Canceling: %s", connReq)
					delete(pending, msg.id)
					continue

				}

				// An existing connection was located, mark as
				// disconnected and execute disconnection
				// callback.
				log.Debugf("Disconnected from %s", connReq)
				delete(conns, msg.id)

				if connReq.conn != nil {
					connReq.conn.Close()
				}

				if cm.cfg.OnDisconnection != nil {
					spawn(func() {
						cm.cfg.OnDisconnection(connReq)
					})
				}

				// All internal state has been cleaned up, if
				// this connection is being removed, we will
				// make no further attempts with this request.
				if !msg.retry {
					connReq.updateState(ConnDisconnected)
					continue
				}

				// Otherwise, we will attempt a reconnection.
				// The connection request is re added to the
				// pending map, so that subsequent processing
				// of connections and failures do not ignore
				// the request.
				connReq.updateState(ConnPending)
				log.Debugf("Reconnecting to %s",
					connReq)
				pending[msg.id] = connReq
				cm.handleFailedConn(connReq, nil)

			case handleFailed:
				connReq := msg.c

				if _, ok := pending[connReq.id]; !ok {
					log.Debugf("Ignoring connection for "+
						"canceled conn req: %s", connReq)
					continue
				}

				connReq.updateState(ConnFailing)
				if shouldWriteConnFailedLog(msg.err) {
					log.Debugf("Failed to connect to %s: %s",
						connReq, msg.err)
				}
				cm.handleFailedConn(connReq, msg.err)

				if cm.cfg.OnConnectionFailed != nil {
					cm.cfg.OnConnectionFailed(connReq)
				}
			}

		case <-cm.quit:
			break out
		}
	}

	cm.wg.Done()
	log.Trace("Connection handler done")
}

// NotifyConnectionRequestComplete notifies the connection
// manager that a peer had been successfully connected and
// marked as good.
func (cm *ConnManager) NotifyConnectionRequestComplete() {
	cm.failedAttempts = 0
}

// NewConnReq creates a new connection request and connects to the
// corresponding address.
func (cm *ConnManager) NewConnReq() {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	c := &ConnReq{}
	atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))

	// Submit a request of a pending connection attempt to the connection
	// manager. By registering the id before the connection is even
	// established, we'll be able to later cancel the connection via the
	// Remove method.
	done := make(chan struct{})
	select {
	case cm.requests <- registerPending{c, done}:
	case <-cm.quit:
		return
	}

	// Wait for the registration to successfully add the pending conn req to
	// the conn manager's internal state.
	select {
	case <-done:
	case <-cm.quit:
		return
	}
	err := cm.associateAddressToConnReq(c)
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
		}
		return
	}

	cm.connect(c)
}

func (cm *ConnManager) associateAddressToConnReq(c *ConnReq) error {
	cm.addressMtx.Lock()
	defer cm.addressMtx.Unlock()

	addr, err := cm.getNewAddress()
	if err != nil {
		return err
	}

	cm.markAddressAsUsed(addr)
	c.Addr = addr
	return nil
}

// Connect assigns an id and dials a connection to the address of the
// connection request.
func (cm *ConnManager) Connect(c *ConnReq) error {
	err := func() error {
		cm.addressMtx.Lock()
		defer cm.addressMtx.Unlock()

		if cm.isAddressUsed(c.Addr) {
			return fmt.Errorf("address %s is already in use", c.Addr)
		}
		cm.markAddressAsUsed(c.Addr)
		return nil
	}()
	if err != nil {
		return err
	}

	cm.connect(c)
	return nil
}

// connect assigns an id and dials a connection to the address of the
// connection request. This function assumes that the connection address
// has checked and already marked as used.
func (cm *ConnManager) connect(c *ConnReq) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	if atomic.LoadUint64(&c.id) == 0 {
		atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))

		// Submit a request of a pending connection attempt to the
		// connection manager. By registering the id before the
		// connection is even established, we'll be able to later
		// cancel the connection via the Remove method.
		done := make(chan struct{})
		select {
		case cm.requests <- registerPending{c, done}:
		case <-cm.quit:
			return
		}

		// Wait for the registration to successfully add the pending
		// conn req to the conn manager's internal state.
		select {
		case <-done:
		case <-cm.quit:
			return
		}
	}

	log.Debugf("Attempting to connect to %s", c)

	conn, err := cm.cfg.Dial(c.Addr)
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
		}
		return
	}

	select {
	case cm.requests <- handleConnected{c, conn}:
	case <-cm.quit:
	}
}

// Disconnect disconnects the connection corresponding to the given connection
// id. If permanent, the connection will be retried with an increasing backoff
// duration.
func (cm *ConnManager) Disconnect(id uint64) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	select {
	case cm.requests <- handleDisconnected{id, true}:
	case <-cm.quit:
	}
}

// Remove removes the connection corresponding to the given connection id from
// known connections.
//
// NOTE: This method can also be used to cancel a lingering connection attempt
// that hasn't yet succeeded.
func (cm *ConnManager) Remove(id uint64) {
	if atomic.LoadInt32(&cm.stop) != 0 {
		return
	}

	select {
	case cm.requests <- handleDisconnected{id, false}:
	case <-cm.quit:
	}
}

// listenHandler accepts incoming connections on a given listener. It must be
// run as a goroutine.
func (cm *ConnManager) listenHandler(listener net.Listener) {
	log.Infof("Server listening on %s", listener.Addr())
	for atomic.LoadInt32(&cm.stop) == 0 {
		conn, err := listener.Accept()
		if err != nil {
			// Only log the error if not forcibly shutting down.
			if atomic.LoadInt32(&cm.stop) == 0 {
				log.Errorf("Can't accept connection: %s", err)
			}
			continue
		}
		spawn(func() {
			cm.cfg.OnAccept(conn)
		})
	}

	cm.wg.Done()
	log.Tracef("Listener handler done for %s", listener.Addr())
}

// Start launches the connection manager and begins connecting to the network.
func (cm *ConnManager) Start() {
	// Already started?
	if atomic.AddInt32(&cm.start, 1) != 1 {
		return
	}

	log.Trace("Connection manager started")
	cm.wg.Add(1)
	spawn(cm.connHandler)

	// Start all the listeners so long as the caller requested them and
	// provided a callback to be invoked when connections are accepted.
	if cm.cfg.OnAccept != nil {
		for _, listener := range cm.cfg.Listeners {
			// Declaring this variable is necessary as it needs be declared in the same
			// scope of the anonymous function below it.
			listenerCopy := listener
			cm.wg.Add(1)
			spawn(func() {
				cm.listenHandler(listenerCopy)
			})
		}
	}

	for i := atomic.LoadUint64(&cm.connReqCount); i < uint64(cm.cfg.TargetOutbound); i++ {
		spawn(cm.NewConnReq)
	}
}

// Wait blocks until the connection manager halts gracefully.
func (cm *ConnManager) Wait() {
	cm.wg.Wait()
}

// Stop gracefully shuts down the connection manager.
func (cm *ConnManager) Stop() {
	if atomic.AddInt32(&cm.stop, 1) != 1 {
		log.Warnf("Connection manager already stopped")
		return
	}

	// Stop all the listeners. There will not be any listeners if
	// listening is disabled.
	for _, listener := range cm.cfg.Listeners {
		// Ignore the error since this is shutdown and there is no way
		// to recover anyways.
		_ = listener.Close()
	}

	close(cm.quit)
	log.Trace("Connection manager stopped")
}

func (cm *ConnManager) getNewAddress() (*net.TCPAddr, error) {
	for tries := 0; tries < 100; tries++ {
		addr := cm.cfg.AddrManager.GetAddress()
		if addr == nil {
			break
		}

		// Check if there's already a connection to the same address.
		netAddr := addr.NetAddress().TCPAddress()
		if cm.isAddressUsed(netAddr) {
			continue
		}

		// Address will not be invalid, local or unroutable
		// because addrmanager rejects those on addition.
		// Just check that we don't already have an address
		// in the same group so that we are not connecting
		// to the same network segment at the expense of
		// others.
		//
		// Networks that accept unroutable connections are exempt
		// from this rule, since they're meant to run within a
		// private subnet, like 10.0.0.0/16.
		if !config.ActiveConfig().NetParams().AcceptUnroutable && cm.isOutboundGroupUsed(netAddr) {
			continue
		}

		// only allow recent nodes (10mins) after we failed 30
		// times
		if tries < 30 && time.Since(addr.LastAttempt().ToNativeTime()) < 10*time.Minute {
			continue
		}

		// allow nondefault ports after 50 failed tries.
		if tries < 50 && fmt.Sprintf("%d", netAddr.Port) !=
			config.ActiveConfig().NetParams().DefaultPort {
			continue
		}

		return netAddr, nil
	}
	return nil, ErrNoAddress
}

// New returns a new connection manager.
// Use Start to start connecting to the network.
func New(cfg *Config) (*ConnManager, error) {
	if cfg.Dial == nil {
		return nil, errors.WithStack(ErrDialNil)
	}
	if cfg.AddrManager == nil {
		return nil, errors.WithStack(ErrAddressManagerNil)
	}
	// Default to sane values
	if cfg.RetryDuration <= 0 {
		cfg.RetryDuration = defaultRetryDuration
	}
	cm := ConnManager{
		cfg:                *cfg, // Copy so caller can't mutate
		requests:           make(chan interface{}),
		quit:               make(chan struct{}),
		usedAddresses:      make(map[string]struct{}),
		usedOutboundGroups: make(map[string]int64),
	}
	return &cm, nil
}
