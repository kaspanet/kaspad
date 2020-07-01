// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package peer

import (
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/util/testtools"

	"github.com/pkg/errors"

	"github.com/btcsuite/go-socks/socks"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// conn mocks a network connection by implementing the net.Conn interface. It
// is used to test peer connection without actually opening a network
// connection.
type conn struct {
	io.Reader
	io.Writer
	io.Closer

	// local network, address for the connection.
	lnet, laddr string

	// remote network, address for the connection.
	rnet, raddr string

	// mocks socks proxy if true
	proxy bool
}

// LocalAddr returns the local address for the connection.
func (c conn) LocalAddr() net.Addr {
	return &addr{c.lnet, c.laddr}
}

// Remote returns the remote address for the connection.
func (c conn) RemoteAddr() net.Addr {
	if !c.proxy {
		return &addr{c.rnet, c.raddr}
	}
	host, strPort, _ := net.SplitHostPort(c.raddr)
	port, _ := strconv.Atoi(strPort)
	return &socks.ProxiedAddr{
		Net:  c.rnet,
		Host: host,
		Port: port,
	}
}

// Close handles closing the connection.
func (c conn) Close() error {
	if c.Closer == nil {
		return nil
	}
	return c.Closer.Close()
}

func (c conn) SetDeadline(t time.Time) error      { return nil }
func (c conn) SetReadDeadline(t time.Time) error  { return nil }
func (c conn) SetWriteDeadline(t time.Time) error { return nil }

// addr mocks a network address
type addr struct {
	net, address string
}

func (m addr) Network() string { return m.net }
func (m addr) String() string  { return m.address }

// pipe turns two mock connections into a full-duplex connection similar to
// net.Pipe to allow pipe's with (fake) addresses.
func pipe(c1, c2 *conn) (*conn, *conn) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	c1.Writer = w1
	c1.Closer = w1
	c2.Reader = r1
	c1.Reader = r2
	c2.Writer = w2
	c2.Closer = w2

	return c1, c2
}

// peerStats holds the expected peer stats used for testing peer.
type peerStats struct {
	wantUserAgent       string
	wantServices        wire.ServiceFlag
	wantProtocolVersion uint32
	wantConnected       bool
	wantVersionKnown    bool
	wantVerAckReceived  bool
	wantLastPingTime    time.Time
	wantLastPingNonce   uint64
	wantLastPingMicros  int64
	wantTimeOffset      int64
	wantBytesSent       uint64
	wantBytesReceived   uint64
}

// testPeer tests the given peer's flags and stats
func testPeer(t *testing.T, p *Peer, s peerStats) {
	if p.UserAgent() != s.wantUserAgent {
		t.Errorf("testPeer: wrong UserAgent - got %v, want %v", p.UserAgent(), s.wantUserAgent)
		return
	}

	if p.Services() != s.wantServices {
		t.Errorf("testPeer: wrong Services - got %v, want %v", p.Services(), s.wantServices)
		return
	}

	if !p.LastPingTime().Equal(s.wantLastPingTime) {
		t.Errorf("testPeer: wrong LastPingTime - got %v, want %v", p.LastPingTime(), s.wantLastPingTime)
		return
	}

	if p.LastPingNonce() != s.wantLastPingNonce {
		t.Errorf("testPeer: wrong LastPingNonce - got %v, want %v", p.LastPingNonce(), s.wantLastPingNonce)
		return
	}

	if p.LastPingMicros() != s.wantLastPingMicros {
		t.Errorf("testPeer: wrong LastPingMicros - got %v, want %v", p.LastPingMicros(), s.wantLastPingMicros)
		return
	}

	if p.VerAckReceived() != s.wantVerAckReceived {
		t.Errorf("testPeer: wrong VerAckReceived - got %v, want %v", p.VerAckReceived(), s.wantVerAckReceived)
		return
	}

	if p.VersionKnown() != s.wantVersionKnown {
		t.Errorf("testPeer: wrong VersionKnown - got %v, want %v", p.VersionKnown(), s.wantVersionKnown)
		return
	}

	if p.ProtocolVersion() != s.wantProtocolVersion {
		t.Errorf("testPeer: wrong ProtocolVersion - got %v, want %v", p.ProtocolVersion(), s.wantProtocolVersion)
		return
	}

	// Allow for a deviation of 1s.
	secondsInMs := time.Second.Milliseconds()
	if p.TimeOffset() > s.wantTimeOffset+secondsInMs && p.TimeOffset() < s.wantTimeOffset-secondsInMs {
		t.Errorf("testPeer: wrong TimeOffset - got %v, want between %v and %v", s.wantTimeOffset-secondsInMs,
			s.wantTimeOffset, s.wantTimeOffset+secondsInMs)
		return
	}

	if p.BytesSent() != s.wantBytesSent {
		t.Errorf("testPeer: wrong BytesSent - got %v, want %v", p.BytesSent(), s.wantBytesSent)
		return
	}

	if p.BytesReceived() != s.wantBytesReceived {
		t.Errorf("testPeer: wrong BytesReceived - got %v, want %v", p.BytesReceived(), s.wantBytesReceived)
		return
	}

	if p.Connected() != s.wantConnected {
		t.Errorf("testPeer: wrong Connected - got %v, want %v", p.Connected(), s.wantConnected)
		return
	}

	stats := p.StatsSnapshot()

	if p.ID() != stats.ID {
		t.Errorf("testPeer: wrong ID - got %v, want %v", p.ID(), stats.ID)
		return
	}

	if p.Addr() != stats.Addr {
		t.Errorf("testPeer: wrong Addr - got %v, want %v", p.Addr(), stats.Addr)
		return
	}

	if p.LastSend() != stats.LastSend {
		t.Errorf("testPeer: wrong LastSend - got %v, want %v", p.LastSend(), stats.LastSend)
		return
	}

	if p.LastRecv() != stats.LastRecv {
		t.Errorf("testPeer: wrong LastRecv - got %v, want %v", p.LastRecv(), stats.LastRecv)
		return
	}
}

// TestPeerConnection tests connection between inbound and outbound peers.
func TestPeerConnection(t *testing.T) {
	inPeerVerack, outPeerVerack, inPeerOnWriteVerack, outPeerOnWriteVerack :=
		make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{}, 1)

	inPeerCfg := &Config{
		Listeners: MessageListeners{
			OnVerAck: func(p *Peer, msg *wire.MsgVerAck) {
				inPeerVerack <- struct{}{}
			},
			OnWrite: func(p *Peer, bytesWritten int, msg wire.Message,
				err error) {
				if _, ok := msg.(*wire.MsgVerAck); ok {
					inPeerOnWriteVerack <- struct{}{}
				}
			},
		},
		UserAgentName:     "peer",
		UserAgentVersion:  "1.0",
		UserAgentComments: []string{"comment"},
		DAGParams:         &dagconfig.MainnetParams,
		ProtocolVersion:   wire.ProtocolVersion, // Configure with older version
		Services:          0,
		SelectedTipHash:   fakeSelectedTipFn,
		SubnetworkID:      nil,
	}
	outPeerCfg := &Config{
		Listeners: MessageListeners{
			OnVerAck: func(p *Peer, msg *wire.MsgVerAck) {
				outPeerVerack <- struct{}{}
			},
			OnWrite: func(p *Peer, bytesWritten int, msg wire.Message,
				err error) {
				if _, ok := msg.(*wire.MsgVerAck); ok {
					outPeerOnWriteVerack <- struct{}{}
				}
			},
		},
		UserAgentName:     "peer",
		UserAgentVersion:  "1.0",
		UserAgentComments: []string{"comment"},
		DAGParams:         &dagconfig.MainnetParams,
		ProtocolVersion:   wire.ProtocolVersion + 1,
		Services:          wire.SFNodeNetwork,
		SelectedTipHash:   fakeSelectedTipFn,
		SubnetworkID:      nil,
	}

	wantStats1 := peerStats{
		wantUserAgent:       wire.DefaultUserAgent + "peer:1.0(comment)/",
		wantServices:        0,
		wantProtocolVersion: wire.ProtocolVersion,
		wantConnected:       true,
		wantVersionKnown:    true,
		wantVerAckReceived:  true,
		wantLastPingTime:    time.Time{},
		wantLastPingNonce:   uint64(0),
		wantLastPingMicros:  int64(0),
		wantTimeOffset:      int64(0),
		wantBytesSent:       195, // 171 version + 24 verack
		wantBytesReceived:   195,
	}
	wantStats2 := peerStats{
		wantUserAgent:       wire.DefaultUserAgent + "peer:1.0(comment)/",
		wantServices:        wire.SFNodeNetwork,
		wantProtocolVersion: wire.ProtocolVersion,
		wantConnected:       true,
		wantVersionKnown:    true,
		wantVerAckReceived:  true,
		wantLastPingTime:    time.Time{},
		wantLastPingNonce:   uint64(0),
		wantLastPingMicros:  int64(0),
		wantTimeOffset:      int64(0),
		wantBytesSent:       195, // 171 version + 24 verack
		wantBytesReceived:   195,
	}

	tests := []struct {
		name  string
		setup func() (*Peer, *Peer, error)
	}{
		{
			"basic handshake",
			func() (*Peer, *Peer, error) {
				inPeer, outPeer, err := setupPeers(inPeerCfg, outPeerCfg)
				if err != nil {
					return nil, nil, err
				}

				// wait for 4 veracks
				if !testtools.WaitTillAllCompleteOrTimeout(time.Second,
					inPeerVerack, inPeerOnWriteVerack, outPeerVerack, outPeerOnWriteVerack) {

					return nil, nil, errors.New("handshake timeout")
				}
				return inPeer, outPeer, nil
			},
		},
		{
			"socks proxy",
			func() (*Peer, *Peer, error) {
				inConn, outConn := pipe(
					&conn{raddr: "10.0.0.1:16111", proxy: true},
					&conn{raddr: "10.0.0.2:16111"},
				)
				inPeer, outPeer, err := setupPeersWithConns(inPeerCfg, outPeerCfg, inConn, outConn)
				if err != nil {
					return nil, nil, err
				}

				// wait for 4 veracks
				if !testtools.WaitTillAllCompleteOrTimeout(time.Second,
					inPeerVerack, inPeerOnWriteVerack, outPeerVerack, outPeerOnWriteVerack) {

					return nil, nil, errors.New("handshake timeout")
				}
				return inPeer, outPeer, nil
			},
		},
	}
	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		inPeer, outPeer, err := test.setup()
		if err != nil {
			t.Errorf("TestPeerConnection setup #%d: unexpected err %v", i, err)
			return
		}
		testPeer(t, inPeer, wantStats2)
		testPeer(t, outPeer, wantStats1)

		inPeer.Disconnect()
		outPeer.Disconnect()
		inPeer.WaitForDisconnect()
		outPeer.WaitForDisconnect()
	}
}

// TestPeerListeners tests that the peer listeners are called as expected.
func TestPeerListeners(t *testing.T) {
	inPeerVerack, outPeerVerack := make(chan struct{}, 1), make(chan struct{}, 1)
	ok := make(chan wire.Message, 20)
	inPeerCfg := &Config{
		Listeners: MessageListeners{
			OnGetAddr: func(p *Peer, msg *wire.MsgGetAddr) {
				ok <- msg
			},
			OnAddr: func(p *Peer, msg *wire.MsgAddr) {
				ok <- msg
			},
			OnPing: func(p *Peer, msg *wire.MsgPing) {
				ok <- msg
			},
			OnPong: func(p *Peer, msg *wire.MsgPong) {
				ok <- msg
			},
			OnTx: func(p *Peer, msg *wire.MsgTx) {
				ok <- msg
			},
			OnBlock: func(p *Peer, msg *wire.MsgBlock, buf []byte) {
				ok <- msg
			},
			OnInv: func(p *Peer, msg *wire.MsgInv) {
				ok <- msg
			},
			OnNotFound: func(p *Peer, msg *wire.MsgNotFound) {
				ok <- msg
			},
			OnGetData: func(p *Peer, msg *wire.MsgGetData) {
				ok <- msg
			},
			OnGetBlockInvs: func(p *Peer, msg *wire.MsgGetBlockInvs) {
				ok <- msg
			},
			OnFeeFilter: func(p *Peer, msg *wire.MsgFeeFilter) {
				ok <- msg
			},
			OnFilterAdd: func(p *Peer, msg *wire.MsgFilterAdd) {
				ok <- msg
			},
			OnFilterClear: func(p *Peer, msg *wire.MsgFilterClear) {
				ok <- msg
			},
			OnFilterLoad: func(p *Peer, msg *wire.MsgFilterLoad) {
				ok <- msg
			},
			OnMerkleBlock: func(p *Peer, msg *wire.MsgMerkleBlock) {
				ok <- msg
			},
			OnVersion: func(p *Peer, msg *wire.MsgVersion) {
				ok <- msg
			},
			OnVerAck: func(p *Peer, msg *wire.MsgVerAck) {
				inPeerVerack <- struct{}{}
			},
			OnReject: func(p *Peer, msg *wire.MsgReject) {
				ok <- msg
			},
		},
		UserAgentName:     "peer",
		UserAgentVersion:  "1.0",
		UserAgentComments: []string{"comment"},
		DAGParams:         &dagconfig.MainnetParams,
		Services:          wire.SFNodeBloom,
		SelectedTipHash:   fakeSelectedTipFn,
		SubnetworkID:      nil,
	}

	outPeerCfg := &Config{}
	*outPeerCfg = *inPeerCfg // copy inPeerCfg
	outPeerCfg.Listeners.OnVerAck = func(p *Peer, msg *wire.MsgVerAck) {
		outPeerVerack <- struct{}{}
	}

	inPeer, outPeer, err := setupPeers(inPeerCfg, outPeerCfg)
	if err != nil {
		t.Errorf("TestPeerListeners: %v", err)
	}
	// wait for 2 veracks
	if !testtools.WaitTillAllCompleteOrTimeout(time.Second, inPeerVerack, outPeerVerack) {
		t.Errorf("TestPeerListeners: Timout waiting for veracks")
	}

	tests := []struct {
		listener string
		msg      wire.Message
	}{
		{
			"OnGetAddr",
			wire.NewMsgGetAddr(false, nil),
		},
		{
			"OnAddr",
			wire.NewMsgAddr(false, nil),
		},
		{
			"OnPing",
			wire.NewMsgPing(42),
		},
		{
			"OnPong",
			wire.NewMsgPong(42),
		},
		{
			"OnTx",
			wire.NewNativeMsgTx(wire.TxVersion, nil, nil),
		},
		{
			"OnBlock",
			wire.NewMsgBlock(wire.NewBlockHeader(1,
				[]*daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, 1, 1)),
		},
		{
			"OnInv",
			wire.NewMsgInv(),
		},
		{
			"OnNotFound",
			wire.NewMsgNotFound(),
		},
		{
			"OnGetData",
			wire.NewMsgGetData(),
		},
		{
			"OnGetBlockInvs",
			wire.NewMsgGetBlockInvs(&daghash.Hash{}, &daghash.Hash{}),
		},
		{
			"OnFeeFilter",
			wire.NewMsgFeeFilter(15000),
		},
		{
			"OnFilterAdd",
			wire.NewMsgFilterAdd([]byte{0x01}),
		},
		{
			"OnFilterClear",
			wire.NewMsgFilterClear(),
		},
		{
			"OnFilterLoad",
			wire.NewMsgFilterLoad([]byte{0x01}, 10, 0, wire.BloomUpdateNone),
		},
		{
			"OnMerkleBlock",
			wire.NewMsgMerkleBlock(wire.NewBlockHeader(1,
				[]*daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, 1, 1)),
		},
		// only one version message is allowed
		// only one verack message is allowed
		{
			"OnReject",
			wire.NewMsgReject("block", wire.RejectDuplicate, "dupe block"),
		},
	}
	t.Logf("Running %d tests", len(tests))
	for _, test := range tests {
		// Queue the test message
		outPeer.QueueMessage(test.msg, nil)
		select {
		case <-ok:
		case <-time.After(time.Second * 1):
			t.Errorf("TestPeerListeners: %s timeout", test.listener)
			return
		}
	}
	inPeer.Disconnect()
	outPeer.Disconnect()
}

// TestOutboundPeer tests that the outbound peer works as expected.
func TestOutboundPeer(t *testing.T) {
	peerCfg := &Config{
		SelectedTipHash: func() *daghash.Hash {
			return &daghash.ZeroHash
		},
		UserAgentName:     "peer",
		UserAgentVersion:  "1.0",
		UserAgentComments: []string{"comment"},
		DAGParams:         &dagconfig.MainnetParams,
		Services:          0,
		SubnetworkID:      nil,
	}

	_, p, err := setupPeers(peerCfg, peerCfg)
	if err != nil {
		t.Fatalf("TestOuboundPeer: unexpected err in setupPeers - %v\n", err)
	}

	// Test trying to connect for a second time and make sure nothing happens.
	err = p.AssociateConnection(p.conn)
	if err != nil {
		t.Fatalf("AssociateConnection for the second time didn't return nil")
	}
	p.Disconnect()

	// Test Queue Inv
	fakeBlockHash := &daghash.Hash{0: 0x00, 1: 0x01}
	fakeInv := wire.NewInvVect(wire.InvTypeBlock, fakeBlockHash)

	// Should be noops as the peer could not connect.
	p.QueueInventory(fakeInv)
	p.AddKnownInventory(fakeInv)
	p.QueueInventory(fakeInv)

	fakeMsg := wire.NewMsgVerAck()
	p.QueueMessage(fakeMsg, nil)
	done := make(chan struct{})
	p.QueueMessage(fakeMsg, done)
	<-done
	p.Disconnect()

	// Test SelectedTipHashAndBlueScore
	var selectedTipHash = func() *daghash.Hash {
		hashStr := "14a0810ac680a3eb3f82edc878cea25ec41d6b790744e5daeef"
		hash, err := daghash.NewHashFromStr(hashStr)
		if err != nil {
			t.Fatalf("daghash.NewHashFromStr: %s", err)
		}
		return hash
	}

	peerCfg.SelectedTipHash = selectedTipHash

	_, p1, err := setupPeers(peerCfg, peerCfg)
	if err != nil {
		t.Fatalf("TestOuboundPeer: unexpected err in setupPeers - %v\n", err)
	}

	// Test Queue Inv after connection
	p1.QueueInventory(fakeInv)
	p1.Disconnect()

	// Test regression
	peerCfg.DAGParams = &dagconfig.RegressionNetParams
	peerCfg.Services = wire.SFNodeBloom
	_, p2, err := setupPeers(peerCfg, peerCfg)
	if err != nil {
		t.Fatalf("NewOutboundPeer: unexpected err - %v\n", err)
	}

	// Test PushXXX
	var addrs []*wire.NetAddress
	for i := 0; i < 5; i++ {
		na := wire.NetAddress{}
		addrs = append(addrs, &na)
	}
	if _, err := p2.PushAddrMsg(addrs, nil); err != nil {
		t.Fatalf("PushAddrMsg: unexpected err %v\n", err)
	}
	if err := p2.PushGetBlockInvsMsg(&daghash.Hash{}, &daghash.Hash{}); err != nil {
		t.Fatalf("PushGetBlockInvsMsg: unexpected err %v\n", err)
	}

	p2.PushRejectMsg("block", wire.RejectMalformed, "malformed", nil, false)
	p2.PushRejectMsg("block", wire.RejectInvalid, "invalid", nil, false)

	// Test Queue Messages
	p2.QueueMessage(wire.NewMsgGetAddr(false, nil), nil)
	p2.QueueMessage(wire.NewMsgPing(1), nil)
	p2.QueueMessage(wire.NewMsgGetData(), nil)
	p2.QueueMessage(wire.NewMsgFeeFilter(20000), nil)

	p2.Disconnect()
}

// Tests that the node disconnects from peers with an unsupported protocol
// version.
func TestUnsupportedVersionPeer(t *testing.T) {
	peerCfg := &Config{
		UserAgentName:     "peer",
		UserAgentVersion:  "1.0",
		UserAgentComments: []string{"comment"},
		DAGParams:         &dagconfig.MainnetParams,
		Services:          0,
		SelectedTipHash:   fakeSelectedTipFn,
	}

	localNA := wire.NewNetAddressIPPort(
		net.ParseIP("10.0.0.1:16111"),
		uint16(16111),
		wire.SFNodeNetwork,
	)
	remoteNA := wire.NewNetAddressIPPort(
		net.ParseIP("10.0.0.2:16111"),
		uint16(16111),
		wire.SFNodeNetwork,
	)
	localConn, remoteConn := pipe(
		&conn{laddr: "10.0.0.1:16111", raddr: "10.0.0.2:16111"},
		&conn{laddr: "10.0.0.2:16111", raddr: "10.0.0.1:16111"},
	)

	p, err := NewOutboundPeer(peerCfg, "10.0.0.1:16111")
	if err != nil {
		t.Fatalf("NewOutboundPeer: unexpected err - %v\n", err)
	}

	go func() {
		err := p.AssociateConnection(localConn)
		wantErrorMessage := "protocol version must be 1 or greater"
		if err == nil {
			t.Fatalf("No error from AssociateConnection to invalid protocol version")
		}
		gotErrorMessage := err.Error()
		if !strings.Contains(gotErrorMessage, wantErrorMessage) {
			t.Fatalf("Wrong error message from AssociateConnection to invalid protocol version.\nWant: '%s'\nGot: '%s'",
				wantErrorMessage, gotErrorMessage)
		}
	}()

	// Read outbound messages to peer into a channel
	outboundMessages := make(chan wire.Message)
	go func() {
		for {
			_, msg, _, err := wire.ReadMessageN(
				remoteConn,
				p.ProtocolVersion(),
				peerCfg.DAGParams.Net,
			)
			if err == io.EOF {
				close(outboundMessages)
				return
			}
			if err != nil {
				t.Errorf("Error reading message from local node: %v\n", err)
				return
			}

			outboundMessages <- msg
		}
	}()

	// Read version message sent to remote peer
	select {
	case msg := <-outboundMessages:
		if _, ok := msg.(*wire.MsgVersion); !ok {
			t.Fatalf("Expected version message, got [%s]", msg.Command())
		}
	case <-time.After(time.Second):
		t.Fatal("Peer did not send version message")
	}

	// Remote peer writes version message advertising invalid protocol version 0
	invalidVersionMsg := wire.NewMsgVersion(remoteNA, localNA, 0, &daghash.ZeroHash, nil)
	invalidVersionMsg.ProtocolVersion = 0

	_, err = wire.WriteMessageN(
		remoteConn.Writer,
		invalidVersionMsg,
		uint32(invalidVersionMsg.ProtocolVersion),
		peerCfg.DAGParams.Net,
	)
	if err != nil {
		t.Fatalf("wire.WriteMessageN: unexpected err - %v\n", err)
	}

	// Expect peer to disconnect automatically
	disconnected := make(chan struct{})
	go func() {
		p.WaitForDisconnect()
		disconnected <- struct{}{}
	}()

	select {
	case <-disconnected:
		close(disconnected)
	case <-time.After(time.Second):
		t.Fatal("Peer did not automatically disconnect")
	}

	// Expect no further outbound messages from peer
	select {
	case msg, chanOpen := <-outboundMessages:
		if chanOpen {
			t.Fatalf("Expected no further messages, received [%s]", msg.Command())
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for remote reader to close")
	}
}

func init() {
	// Allow self connection when running the tests.
	TstAllowSelfConns()
}

func fakeSelectedTipFn() *daghash.Hash {
	return &daghash.Hash{0x12, 0x34}
}

func setupPeers(inPeerCfg, outPeerCfg *Config) (inPeer *Peer, outPeer *Peer, err error) {
	inConn, outConn := pipe(
		&conn{raddr: "10.0.0.1:16111"},
		&conn{raddr: "10.0.0.2:16111"},
	)
	return setupPeersWithConns(inPeerCfg, outPeerCfg, inConn, outConn)
}

func setupPeersWithConns(inPeerCfg, outPeerCfg *Config, inConn, outConn *conn) (inPeer *Peer, outPeer *Peer, err error) {
	inPeer = NewInboundPeer(inPeerCfg)
	inPeerDone := make(chan struct{})
	var inPeerErr error
	go func() {
		inPeerErr = inPeer.AssociateConnection(inConn)
		inPeerDone <- struct{}{}
	}()

	outPeer, err = NewOutboundPeer(outPeerCfg, outConn.raddr)
	if err != nil {
		return nil, nil, err
	}
	outPeerDone := make(chan struct{})
	var outPeerErr error
	go func() {
		outPeerErr = outPeer.AssociateConnection(outConn)
		outPeerDone <- struct{}{}
	}()

	// wait for AssociateConnection to complete in all instances
	if !testtools.WaitTillAllCompleteOrTimeout(2*time.Second, inPeerDone, outPeerDone) {
		return nil, nil, errors.New("handshake timeout")
	}

	if inPeerErr != nil && outPeerErr != nil {
		return nil, nil, errors.Errorf("both inPeer and outPeer failed connecting: \nInPeer: %+v\nOutPeer: %+v",
			inPeerErr, outPeerErr)
	}
	if inPeerErr != nil {
		return nil, nil, inPeerErr
	}
	if outPeerErr != nil {
		return nil, nil, outPeerErr
	}

	return inPeer, outPeer, nil
}
