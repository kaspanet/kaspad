// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/pkg/errors"
)

func init() {
	// Override the max retry duration when running tests.
	maxRetryDuration = 2 * time.Millisecond
}

// mockAddr mocks a network address
type mockAddr struct {
	net, address string
}

func (m mockAddr) Network() string { return m.net }
func (m mockAddr) String() string  { return m.address }

// mockConn mocks a network connection by implementing the net.Conn interface.
type mockConn struct {
	io.Reader
	io.Writer
	io.Closer

	// local network, address for the connection.
	lnet, laddr string

	// remote network, address for the connection.
	rAddr net.Addr
}

// LocalAddr returns the local address for the connection.
func (c mockConn) LocalAddr() net.Addr {
	return &mockAddr{c.lnet, c.laddr}
}

// RemoteAddr returns the remote address for the connection.
func (c mockConn) RemoteAddr() net.Addr {
	return &mockAddr{c.rAddr.Network(), c.rAddr.String()}
}

// Close handles closing the connection.
func (c mockConn) Close() error {
	return nil
}

func (c mockConn) SetDeadline(t time.Time) error      { return nil }
func (c mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c mockConn) SetWriteDeadline(t time.Time) error { return nil }

// mockDialer mocks the net.Dial interface by returning a mock connection to
// the given address.
func mockDialer(addr net.Addr) (net.Conn, error) {
	r, w := io.Pipe()
	c := &mockConn{rAddr: addr}
	c.Reader = r
	c.Writer = w
	return c, nil
}

// TestNewConfig tests that new ConnManager config is validated as expected.
func TestNewConfig(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	_, err := New(&Config{})
	if !errors.Is(err, ErrDialNil) {
		t.Fatalf("New expected error: %s, got %s", ErrDialNil, err)
	}

	_, err = New(&Config{
		Dial: mockDialer,
	})
	if !errors.Is(err, ErrAddressManagerNil) {
		t.Fatalf("New expected error: %s, got %s", ErrAddressManagerNil, err)
	}

	amgr, teardown := addressManagerForTest(t, "TestNewConfig", 10)
	defer teardown()

	_, err = New(&Config{
		Dial:        mockDialer,
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("New unexpected error: %v", err)
	}
}

// TestStartStop tests that the connection manager starts and stops as
// expected.
func TestStartStop(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	connected := make(chan *ConnReq)
	disconnected := make(chan *ConnReq)

	amgr, teardown := addressManagerForTest(t, "TestStartStop", 10)
	defer teardown()

	cmgr, err := New(&Config{
		TargetOutbound: 1,
		AddrManager:    amgr,
		Dial:           mockDialer,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- c
		},
		OnDisconnection: func(c *ConnReq) {
			disconnected <- c
		},
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()
	gotConnReq := <-connected
	cmgr.Stop()
	// already stopped
	cmgr.Stop()
	// ignored
	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	err = cmgr.Connect(cr)
	if err != nil {
		t.Fatalf("Connect error: %s", err)
	}
	if cr.ID() != 0 {
		t.Fatalf("start/stop: got id: %v, want: 0", cr.ID())
	}
	cmgr.Disconnect(gotConnReq.ID())
	cmgr.Remove(gotConnReq.ID())
	select {
	case <-disconnected:
		t.Fatalf("start/stop: unexpected disconnection")
	case <-time.Tick(10 * time.Millisecond):
		break
	}
}

func overrideActiveConfig() func() {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	return func() {
		// Give some extra time to all open NewConnReq goroutines
		// to finish before restoring the active config to prevent
		// potential panics.
		time.Sleep(10 * time.Millisecond)

		config.SetActiveConfig(originalActiveCfg)
	}
}

func addressManagerForTest(t *testing.T, testName string, numAddresses uint8) (*addrmgr.AddrManager, func()) {
	amgr, teardown := createEmptyAddressManagerForTest(t, testName)

	for i := uint8(0); i < numAddresses; i++ {
		ip := fmt.Sprintf("173.%d.115.66:16511", i)
		err := amgr.AddAddressByIP(ip, nil)
		if err != nil {
			t.Fatalf("AddAddressByIP unexpectedly failed to add IP %s: %s", ip, err)
		}
	}

	return amgr, teardown
}

func createEmptyAddressManagerForTest(t *testing.T, testName string) (*addrmgr.AddrManager, func()) {
	path, err := ioutil.TempDir("", fmt.Sprintf("%s-database", testName))
	if err != nil {
		t.Fatalf("createEmptyAddressManagerForTest: TempDir unexpectedly "+
			"failed: %s", err)
	}

	err = dbaccess.Open(path)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	return addrmgr.New(nil, nil), func() {
		// Wait for the connection manager to finish, so it'll
		// have access to the address manager as long as it's
		// alive.
		time.Sleep(10 * time.Millisecond)

		err := dbaccess.Close()
		if err != nil {
			t.Fatalf("error closing the database: %s", err)
		}
	}
}

// TestConnectMode tests that the connection manager works in the connect mode.
//
// In connect mode, automatic connections are disabled, so we test that
// requests using Connect are handled and that no other connections are made.
func TestConnectMode(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	connected := make(chan *ConnReq)
	amgr, teardown := addressManagerForTest(t, "TestConnectMode", 10)
	defer teardown()

	cmgr, err := New(&Config{
		TargetOutbound: 0,
		Dial:           mockDialer,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- c
		},
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	cmgr.Start()
	cmgr.Connect(cr)
	gotConnReq := <-connected
	wantID := cr.ID()
	gotID := gotConnReq.ID()
	if gotID != wantID {
		t.Fatalf("connect mode: %v - want ID %v, got ID %v", cr.Addr, wantID, gotID)
	}
	gotState := cr.State()
	wantState := ConnEstablished
	if gotState != wantState {
		t.Fatalf("connect mode: %v - want state %v, got state %v", cr.Addr, wantState, gotState)
	}
	select {
	case c := <-connected:
		t.Fatalf("connect mode: got unexpected connection - %v", c.Addr)
	case <-time.After(time.Millisecond):
		break
	}
	cmgr.Stop()
	cmgr.Wait()
}

// TestTargetOutbound tests the target number of outbound connections.
//
// We wait until all connections are established, then test they there are the
// only connections made.
func TestTargetOutbound(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	const numAddressesInAddressManager = 10
	targetOutbound := uint32(numAddressesInAddressManager - 2)
	connected := make(chan *ConnReq)

	amgr, teardown := addressManagerForTest(t, "TestTargetOutbound", 10)
	defer teardown()

	cmgr, err := New(&Config{
		TargetOutbound: targetOutbound,
		Dial:           mockDialer,
		AddrManager:    amgr,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- c
		},
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()
	for i := uint32(0); i < targetOutbound; i++ {
		<-connected
	}

	select {
	case c := <-connected:
		t.Fatalf("target outbound: got unexpected connection - %v", c.Addr)
	case <-time.After(time.Millisecond):
		break
	}
	cmgr.Stop()
	cmgr.Wait()
}

// TestDuplicateOutboundConnections tests that connection requests cannot use an already used address.
// It checks it by creating one connection request for each address in the address manager, so that
// the next connection request will have to fail because no unused address will be available.
func TestDuplicateOutboundConnections(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	const numAddressesInAddressManager = 10
	targetOutbound := uint32(numAddressesInAddressManager - 1)
	connected := make(chan struct{})
	failedConnections := make(chan struct{})

	amgr, teardown := addressManagerForTest(t, "TestDuplicateOutboundConnections", 10)
	defer teardown()

	cmgr, err := New(&Config{
		TargetOutbound: targetOutbound,
		Dial:           mockDialer,
		AddrManager:    amgr,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- struct{}{}
		},
		OnConnectionFailed: func(_ *ConnReq) {
			failedConnections <- struct{}{}
		},
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()
	for i := uint32(0); i < targetOutbound; i++ {
		<-connected
	}

	time.Sleep(time.Millisecond)

	// Here we check that making a manual connection request beyond the target outbound connection
	// doesn't fail, so we can know that the reason such connection request will fail is an address
	// related issue.
	cmgr.NewConnReq()
	select {
	case <-connected:
		break
	case <-time.After(time.Millisecond):
		t.Fatalf("connection request unexpectedly didn't connect")
	}

	select {
	case <-failedConnections:
		t.Fatalf("a connection request unexpectedly failed")
	case <-time.After(time.Millisecond):
		break
	}

	// After we created numAddressesInAddressManager connection requests, this request should fail
	// because there aren't any more available addresses.
	cmgr.NewConnReq()
	select {
	case <-connected:
		t.Fatalf("connection request unexpectedly succeeded")
	case <-time.After(time.Millisecond):
		t.Fatalf("connection request didn't fail as expected")
	case <-failedConnections:
		break
	}

	cmgr.Stop()
	cmgr.Wait()
}

// TestSameOutboundGroupConnections tests that connection requests cannot use an address with an already used
// address CIDR group.
// It checks it by creating an address manager with only two addresses, that both belong to the same CIDR group
// and checks that the second connection request fails.
func TestSameOutboundGroupConnections(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	amgr, teardown := createEmptyAddressManagerForTest(t, "TestSameOutboundGroupConnections")
	defer teardown()

	err := amgr.AddAddressByIP("173.190.115.66:16511", nil)
	if err != nil {
		t.Fatalf("AddAddressByIP unexpectedly failed: %s", err)
	}

	err = amgr.AddAddressByIP("173.190.115.67:16511", nil)
	if err != nil {
		t.Fatalf("AddAddressByIP unexpectedly failed: %s", err)
	}

	connected := make(chan struct{})
	failedConnections := make(chan struct{})
	cmgr, err := New(&Config{
		TargetOutbound: 0,
		Dial:           mockDialer,
		AddrManager:    amgr,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- struct{}{}
		},
		OnConnectionFailed: func(_ *ConnReq) {
			failedConnections <- struct{}{}
		},
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}

	cmgr.Start()

	cmgr.NewConnReq()
	select {
	case <-connected:
		break
	case <-time.After(time.Millisecond):
		t.Fatalf("connection request unexpectedly didn't connect")
	}

	select {
	case <-failedConnections:
		t.Fatalf("a connection request unexpectedly failed")
	case <-time.After(time.Millisecond):
		break
	}

	cmgr.NewConnReq()
	select {
	case <-connected:
		t.Fatalf("connection request unexpectedly succeeded")
	case <-time.After(time.Millisecond):
		t.Fatalf("connection request didn't fail as expected")
	case <-failedConnections:
		break
	}

	cmgr.Stop()
	cmgr.Wait()
}

// TestRetryPermanent tests that permanent connection requests are retried.
//
// We make a permanent connection request using Connect, disconnect it using
// Disconnect and we wait for it to be connected back.
func TestRetryPermanent(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	connected := make(chan *ConnReq)
	disconnected := make(chan *ConnReq)

	amgr, teardown := addressManagerForTest(t, "TestRetryPermanent", 10)
	defer teardown()

	cmgr, err := New(&Config{
		RetryDuration:  time.Millisecond,
		TargetOutbound: 0,
		Dial:           mockDialer,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- c
		},
		OnDisconnection: func(c *ConnReq) {
			disconnected <- c
		},
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}

	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	go cmgr.Connect(cr)
	cmgr.Start()
	gotConnReq := <-connected
	wantID := cr.ID()
	gotID := gotConnReq.ID()
	if gotID != wantID {
		t.Fatalf("retry: %v - want ID %v, got ID %v", cr.Addr, wantID, gotID)
	}
	gotState := cr.State()
	wantState := ConnEstablished
	if gotState != wantState {
		t.Fatalf("retry: %v - want state %v, got state %v", cr.Addr, wantState, gotState)
	}

	cmgr.Disconnect(cr.ID())
	gotConnReq = <-disconnected
	wantID = cr.ID()
	gotID = gotConnReq.ID()
	if gotID != wantID {
		t.Fatalf("retry: %v - want ID %v, got ID %v", cr.Addr, wantID, gotID)
	}
	gotState = cr.State()
	wantState = ConnPending
	if gotState != wantState {
		// There is a small chance that connection has already been established,
		// so check for that as well
		if gotState != ConnEstablished {
			t.Fatalf("retry: %v - want state %v, got state %v", cr.Addr, wantState, gotState)
		}
	}

	gotConnReq = <-connected
	wantID = cr.ID()
	gotID = gotConnReq.ID()
	if gotID != wantID {
		t.Fatalf("retry: %v - want ID %v, got ID %v", cr.Addr, wantID, gotID)
	}
	gotState = cr.State()
	wantState = ConnEstablished
	if gotState != wantState {
		t.Fatalf("retry: %v - want state %v, got state %v", cr.Addr, wantState, gotState)
	}

	cmgr.Remove(cr.ID())
	gotConnReq = <-disconnected

	// Wait for status to be updated
	time.Sleep(10 * time.Millisecond)
	wantID = cr.ID()
	gotID = gotConnReq.ID()
	if gotID != wantID {
		t.Fatalf("retry: %v - want ID %v, got ID %v", cr.Addr, wantID, gotID)
	}
	gotState = cr.State()
	wantState = ConnDisconnected
	if gotState != wantState {
		t.Fatalf("retry: %v - want state %v, got state %v", cr.Addr, wantState, gotState)
	}
	cmgr.Stop()
	cmgr.Wait()
}

// TestMaxRetryDuration tests the maximum retry duration.
//
// We have a timed dialer which initially returns err but after RetryDuration
// hits maxRetryDuration returns a mock conn.
func TestMaxRetryDuration(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	networkUp := make(chan struct{})
	time.AfterFunc(5*time.Millisecond, func() {
		close(networkUp)
	})
	timedDialer := func(addr net.Addr) (net.Conn, error) {
		select {
		case <-networkUp:
			return mockDialer(addr)
		default:
			return nil, errors.New("network down")
		}
	}

	amgr, teardown := addressManagerForTest(t, "TestMaxRetryDuration", 10)
	defer teardown()

	connected := make(chan *ConnReq)
	cmgr, err := New(&Config{
		RetryDuration:  time.Millisecond,
		TargetOutbound: 0,
		Dial:           timedDialer,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- c
		},
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}

	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	go cmgr.Connect(cr)
	cmgr.Start()
	// retry in 1ms
	// retry in 2ms - max retry duration reached
	// retry in 2ms - timedDialer returns mockDial
	select {
	case <-connected:
	case <-time.Tick(100 * time.Millisecond):
		t.Fatalf("max retry duration: connection timeout")
	}
	cmgr.Stop()
	cmgr.Wait()
}

// TestNetworkFailure tests that the connection manager handles a network
// failure gracefully.
func TestNetworkFailure(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	var dials uint32
	errDialer := func(net net.Addr) (net.Conn, error) {
		atomic.AddUint32(&dials, 1)
		return nil, errors.New("network down")
	}

	amgr, teardown := addressManagerForTest(t, "TestNetworkFailure", 10)
	defer teardown()

	cmgr, err := New(&Config{
		TargetOutbound: 5,
		RetryDuration:  5 * time.Millisecond,
		Dial:           errDialer,
		AddrManager:    amgr,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			t.Fatalf("network failure: got unexpected connection - %v", c.Addr)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()
	time.Sleep(10 * time.Millisecond)
	cmgr.Stop()
	cmgr.Wait()
	wantMaxDials := uint32(75)
	if atomic.LoadUint32(&dials) > wantMaxDials {
		t.Fatalf("network failure: unexpected number of dials - got %v, want < %v",
			atomic.LoadUint32(&dials), wantMaxDials)
	}
}

// TestStopFailed tests that failed connections are ignored after connmgr is
// stopped.
//
// We have a dailer which sets the stop flag on the conn manager and returns an
// err so that the handler assumes that the conn manager is stopped and ignores
// the failure.
func TestStopFailed(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	done := make(chan struct{}, 1)
	waitDialer := func(addr net.Addr) (net.Conn, error) {
		done <- struct{}{}
		time.Sleep(time.Millisecond)
		return nil, errors.New("network down")
	}

	amgr, teardown := addressManagerForTest(t, "TestStopFailed", 10)
	defer teardown()

	cmgr, err := New(&Config{
		Dial:        waitDialer,
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()
	go func() {
		<-done
		atomic.StoreInt32(&cmgr.stop, 1)
		time.Sleep(2 * time.Millisecond)
		atomic.StoreInt32(&cmgr.stop, 0)
		cmgr.Stop()
	}()
	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	go cmgr.Connect(cr)
	cmgr.Wait()
}

// TestRemovePendingConnection tests that it's possible to cancel a pending
// connection, removing its internal state from the ConnMgr.
func TestRemovePendingConnection(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	// Create a ConnMgr instance with an instance of a dialer that'll never
	// succeed.
	wait := make(chan struct{})
	indefiniteDialer := func(addr net.Addr) (net.Conn, error) {
		<-wait
		return nil, errors.Errorf("error")
	}

	amgr, teardown := addressManagerForTest(t, "TestRemovePendingConnection", 10)
	defer teardown()

	cmgr, err := New(&Config{
		Dial:        indefiniteDialer,
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()

	// Establish a connection request to a random IP we've chosen.
	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	go cmgr.Connect(cr)

	time.Sleep(10 * time.Millisecond)

	if cr.State() != ConnPending {
		t.Fatalf("pending request hasn't been registered, status: %v",
			cr.State())
	}

	// The request launched above will actually never be able to establish
	// a connection. So we'll cancel it _before_ it's able to be completed.
	cmgr.Remove(cr.ID())

	time.Sleep(10 * time.Millisecond)

	// Now examine the status of the connection request, it should read a
	// status of failed.
	if cr.State() != ConnCanceled {
		t.Fatalf("request wasn't canceled, status is: %v", cr.State())
	}

	close(wait)
	cmgr.Stop()
	cmgr.Wait()
}

// TestCancelIgnoreDelayedConnection tests that a canceled connection request will
// not execute the on connection callback, even if an outstanding retry
// succeeds.
func TestCancelIgnoreDelayedConnection(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	retryTimeout := 10 * time.Millisecond

	// Setup a dialer that will continue to return an error until the
	// connect chan is signaled, the dial attempt immediately after will
	// succeed in returning a connection.
	connect := make(chan struct{})
	failingDialer := func(addr net.Addr) (net.Conn, error) {
		select {
		case <-connect:
			return mockDialer(addr)
		default:
		}

		return nil, errors.Errorf("error")
	}

	connected := make(chan *ConnReq)

	amgr, teardown := addressManagerForTest(t, "TestCancelIgnoreDelayedConnection", 10)
	defer teardown()

	cmgr, err := New(&Config{
		Dial:          failingDialer,
		RetryDuration: retryTimeout,
		OnConnection: func(c *ConnReq, conn net.Conn) {
			connected <- c
		},
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()

	// Establish a connection request to a random IP we've chosen.
	cr := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
	}
	cmgr.Connect(cr)

	// Allow for the first retry timeout to elapse.
	time.Sleep(2 * retryTimeout)

	// Connection be marked as failed, even after reattempting to
	// connect.
	if cr.State() != ConnFailing {
		t.Fatalf("failing request should have status failed, status: %v",
			cr.State())
	}

	// Remove the connection, and then immediately allow the next connection
	// to succeed.
	cmgr.Remove(cr.ID())
	close(connect)

	// Allow the connection manager to process the removal.
	time.Sleep(5 * time.Millisecond)

	// Now examine the status of the connection request, it should read a
	// status of canceled.
	if cr.State() != ConnCanceled {
		t.Fatalf("request wasn't canceled, status is: %v", cr.State())
	}

	// Finally, the connection manager should not signal the on-connection
	// callback, since we explicitly canceled this request. We give a
	// generous window to ensure the connection manager's lienar backoff is
	// allowed to properly elapse.
	select {
	case <-connected:
		t.Fatalf("on-connect should not be called for canceled req")
	case <-time.After(5 * retryTimeout):
	}
	cmgr.Stop()
	cmgr.Wait()
}

// mockListener implements the net.Listener interface and is used to test
// code that deals with net.Listeners without having to actually make any real
// connections.
type mockListener struct {
	localAddr   string
	provideConn chan net.Conn
}

// Accept returns a mock connection when it receives a signal via the Connect
// function.
//
// This is part of the net.Listener interface.
func (m *mockListener) Accept() (net.Conn, error) {
	for conn := range m.provideConn {
		return conn, nil
	}
	return nil, errors.New("network connection closed")
}

// Close closes the mock listener which will cause any blocked Accept
// operations to be unblocked and return errors.
//
// This is part of the net.Listener interface.
func (m *mockListener) Close() error {
	close(m.provideConn)
	return nil
}

// Addr returns the address the mock listener was configured with.
//
// This is part of the net.Listener interface.
func (m *mockListener) Addr() net.Addr {
	return &mockAddr{"tcp", m.localAddr}
}

// Connect fakes a connection to the mock listener from the provided remote
// address. It will cause the Accept function to return a mock connection
// configured with the provided remote address and the local address for the
// mock listener.
func (m *mockListener) Connect(ip string, port int) {
	m.provideConn <- &mockConn{
		laddr: m.localAddr,
		lnet:  "tcp",
		rAddr: &net.TCPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
	}
}

// newMockListener returns a new mock listener for the provided local address
// and port. No ports are actually opened.
func newMockListener(localAddr string) *mockListener {
	return &mockListener{
		localAddr:   localAddr,
		provideConn: make(chan net.Conn),
	}
}

// TestListeners ensures providing listeners to the connection manager along
// with an accept callback works properly.
func TestListeners(t *testing.T) {
	restoreConfig := overrideActiveConfig()
	defer restoreConfig()

	// Setup a connection manager with a couple of mock listeners that
	// notify a channel when they receive mock connections.
	receivedConns := make(chan net.Conn)
	listener1 := newMockListener("127.0.0.1:16111")
	listener2 := newMockListener("127.0.0.1:9333")
	listeners := []net.Listener{listener1, listener2}

	amgr, teardown := addressManagerForTest(t, "TestListeners", 10)
	defer teardown()

	cmgr, err := New(&Config{
		Listeners: listeners,
		OnAccept: func(conn net.Conn) {
			receivedConns <- conn
		},
		Dial:        mockDialer,
		AddrManager: amgr,
	})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err)
	}
	cmgr.Start()

	// Fake a couple of mock connections to each of the listeners.
	go func() {
		for i, listener := range listeners {
			l := listener.(*mockListener)
			l.Connect("127.0.0.1", 10000+i*2)
			l.Connect("127.0.0.1", 10000+i*2+1)
		}
	}()

	// Tally the receive connections to ensure the expected number are
	// received. Also, fail the test after a timeout so it will not hang
	// forever should the test not work.
	expectedNumConns := len(listeners) * 2
	var numConns int
out:
	for {
		select {
		case <-receivedConns:
			numConns++
			if numConns == expectedNumConns {
				break out
			}

		case <-time.After(time.Millisecond * 50):
			t.Fatalf("Timeout waiting for %d expected connections",
				expectedNumConns)
		}
	}

	cmgr.Stop()
	cmgr.Wait()
}

// TestConnReqString ensures that ConnReq.String() does not crash
func TestConnReqString(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ConnReq.String crashed %v", r)
		}
	}()
	cr1 := &ConnReq{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 18555,
		},
		Permanent: true,
	}
	_ = cr1.String()
	cr2 := &ConnReq{}
	_ = cr2.String()
}
