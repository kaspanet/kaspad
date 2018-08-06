package server

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/addrmgr"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/blockdag/indexers"
	"github.com/daglabs/btcd/config"
	"github.com/daglabs/btcd/connmgr"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/mempool"
	"github.com/daglabs/btcd/mining"
	"github.com/daglabs/btcd/mining/cpuminer"
	"github.com/daglabs/btcd/netsync"
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/wire"
	"github.com/daglabs/btcutil"
	"github.com/daglabs/btcutil/bloom"
)

// RelayMsg packages an inventory vector along with the newly discovered
// inventory so the relay has access to that information.
type RelayMsg struct {
	InvVect *wire.InvVect
	Data    interface{}
}

// Peer extends the peer to maintain state shared by the server and
// the blockmanager.
type Peer struct {
	// The following variables must only be used atomically
	NonSafeFeeFilter int64

	*peer.Peer

	connReq         *connmgr.ConnReq
	Server          *Server
	Persistent      bool
	ContinueHash    *daghash.Hash
	RelayMtx        sync.Mutex
	DisableRelayTx  bool
	SentAddrs       bool
	IsWhitelisted   bool
	Filter          *bloom.Filter
	KnownAddresses  map[string]struct{}
	DynamicBanScore connmgr.DynamicBanScore
	Quit            chan struct{}
	// The following chans are used to sync blockmanager and server.
	TxProcessed    chan struct{}
	BlockProcessed chan struct{}
}

// RPCServer provides a concurrent safe RPC server to a chain server.
type RPCServer struct {
	Started                int32
	Shutdown               int32
	Cfg                    RPCServerConfig
	Authsha                [sha256.Size]byte
	Limitauthsha           [sha256.Size]byte
	NtfnMgr                *WSNotificationManager
	NumClients             int32
	StatusLines            map[int]string
	StatusLock             sync.RWMutex
	WG                     sync.WaitGroup
	GbtWorkState           *GbtWorkState
	HelpCacher             *HelpCacher
	RequestProcessShutdown chan struct{}
	Quit                   chan int
}

// Server provides a bitcoin server for handling communications to and from
// bitcoin peers.
type Server struct {
	// The following variables must only be used atomically.
	// Putting the uint64s first makes them 64-bit aligned for 32-bit systems.
	BytesReceived uint64 // Total bytes received from all peers since start.
	BytesSent     uint64 // Total bytes sent by all peers since start.
	Started       int32
	Shutdown      int32
	ShutdownSched int32
	StartupTime   int64

	DagParams            *dagconfig.Params
	AddrManager          *addrmgr.AddrManager
	ConnManager          *connmgr.ConnManager
	SigCache             *txscript.SigCache
	RPCServer            *RPCServer
	SyncManager          *netsync.SyncManager
	DAG                  *blockdag.BlockDAG
	TXMemPool            *mempool.TxPool
	CPUMiner             *cpuminer.CPUMiner
	ModifyRebroadcastInv chan interface{}
	NewPeers             chan *Peer
	DonePeers            chan *Peer
	BanPeers             chan *Peer
	Query                chan interface{}
	RelayInv             chan RelayMsg
	Broadcast            chan BroadcastMsg
	PeerHeightsUpdate    chan UpdatePeerHeightsMsg
	WG                   sync.WaitGroup
	Quit                 chan struct{}
	NAT                  NAT
	DB                   database.DB
	TimeSource           blockdag.MedianTimeSource
	Services             wire.ServiceFlag

	// The following fields are used for optional indexes.  They will be nil
	// if the associated index is not enabled.  These fields are set during
	// initial creation of the server and never changed afterwards, so they
	// do not need to be protected for concurrent access.
	TxIndex   *indexers.TxIndex
	AddrIndex *indexers.AddrIndex
	CfIndex   *indexers.CfIndex

	// The fee estimator keeps track of how long transactions are left in
	// the mempool before they are mined into blocks.
	FeeEstimator *mempool.FeeEstimator

	// cfCheckptCaches stores a cached slice of filter headers for cfcheckpt
	// messages for each filter type.
	CfCheckptCaches    map[wire.FilterType][]CfHeaderKV
	CfCheckptCachesMtx sync.RWMutex
}

// RPCServerConfig is a descriptor containing the RPC server configuration.
type RPCServerConfig struct {
	// Listeners defines a slice of listeners for which the RPC server will
	// take ownership of and accept connections.  Since the RPC server takes
	// ownership of these listeners, they will be closed when the RPC server
	// is stopped.
	Listeners []net.Listener

	// StartupTime is the unix timestamp for when the server that is hosting
	// the RPC server started.
	StartupTime int64

	// ConnMgr defines the connection manager for the RPC server to use.  It
	// provides the RPC server with a means to do things such as add,
	// remove, connect, disconnect, and query peers as well as other
	// connection-related data and tasks.
	ConnMgr RPCServerConnManager

	// SyncMgr defines the sync manager for the RPC server to use.
	SyncMgr RPCServerSyncManager

	// These fields allow the RPC server to interface with the local block
	// chain data and state.
	TimeSource  blockdag.MedianTimeSource
	DAG         *blockdag.BlockDAG
	ChainParams *dagconfig.Params
	DB          database.DB

	// TxMemPool defines the transaction memory pool to interact with.
	TxMemPool *mempool.TxPool

	// These fields allow the RPC server to interface with mining.
	//
	// Generator produces block templates and the CPUMiner solves them using
	// the CPU.  CPU mining is typically only useful for test purposes when
	// doing regression or simulation testing.
	Generator *mining.BlkTmplGenerator
	CPUMiner  *cpuminer.CPUMiner

	// These fields define any optional indexes the RPC server can make use
	// of to provide additional data when queried.
	TxIndex   *indexers.TxIndex
	AddrIndex *indexers.AddrIndex
	CfIndex   *indexers.CfIndex

	// The fee estimator keeps track of how long transactions are left in
	// the mempool before they are mined into blocks.
	FeeEstimator *mempool.FeeEstimator
}

// RPCServerConnManager represents a connection manager for use with the RPC
// server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type RPCServerConnManager interface {
	// Connect adds the provided address as a new outbound peer.  The
	// permanent flag indicates whether or not to make the peer persistent
	// and reconnect if the connection is lost.  Attempting to connect to an
	// already existing peer will return an error.
	Connect(addr string, permanent bool) error

	// RemoveByID removes the peer associated with the provided id from the
	// list of persistent peers.  Attempting to remove an id that does not
	// exist will return an error.
	RemoveByID(id int32) error

	// RemoveByAddr removes the peer associated with the provided address
	// from the list of persistent peers.  Attempting to remove an address
	// that does not exist will return an error.
	RemoveByAddr(addr string) error

	// DisconnectByID disconnects the peer associated with the provided id.
	// This applies to both inbound and outbound peers.  Attempting to
	// remove an id that does not exist will return an error.
	DisconnectByID(id int32) error

	// DisconnectByAddr disconnects the peer associated with the provided
	// address.  This applies to both inbound and outbound peers.
	// Attempting to remove an address that does not exist will return an
	// error.
	DisconnectByAddr(addr string) error

	// ConnectedCount returns the number of currently connected peers.
	ConnectedCount() int32

	// NetTotals returns the sum of all bytes received and sent across the
	// network for all peers.
	NetTotals() (uint64, uint64)

	// ConnectedPeers returns an array consisting of all connected peers.
	ConnectedPeers() []RPCServerPeer

	// PersistentPeers returns an array consisting of all the persistent
	// peers.
	PersistentPeers() []RPCServerPeer

	// BroadcastMessage sends the provided message to all currently
	// connected peers.
	BroadcastMessage(msg wire.Message)

	// AddRebroadcastInventory adds the provided inventory to the list of
	// inventories to be rebroadcast at random intervals until they show up
	// in a block.
	AddRebroadcastInventory(iv *wire.InvVect, data interface{})

	// RelayTransactions generates and relays inventory vectors for all of
	// the passed transactions to all connected peers.
	RelayTransactions(txns []*mempool.TxDesc)
}

// RPCServerSyncManager represents a sync manager for use with the RPC server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type RPCServerSyncManager interface {
	// IsCurrent returns whether or not the sync manager believes the chain
	// is current as compared to the rest of the network.
	IsCurrent() bool

	// SubmitBlock submits the provided block to the network after
	// processing it locally.
	SubmitBlock(block *btcutil.Block, flags blockdag.BehaviorFlags) (bool, error)

	// Pause pauses the sync manager until the returned channel is closed.
	Pause() chan<- struct{}

	// SyncPeerID returns the ID of the peer that is currently the peer being
	// used to sync from or 0 if there is none.
	SyncPeerID() int32

	// LocateHeaders returns the headers of the blocks after the first known
	// block in the provided locators until the provided stop hash or the
	// current tip is reached, up to a max of wire.MaxBlockHeadersPerMsg
	// hashes.
	LocateHeaders(locators []*daghash.Hash, hashStop *daghash.Hash) []wire.BlockHeader
}

// WSNotificationManager is a connection and notification manager used for
// websockets.  It allows websocket clients to register for notifications they
// are interested in.  When an event happens elsewhere in the code such as
// transactions being added to the memory pool or block connects/disconnects,
// the notification manager is provided with the relevant details needed to
// figure out which websocket clients need to be notified based on what they
// have registered for and notifies them accordingly.  It is also used to keep
// track of all connected websocket clients.
type WSNotificationManager struct {
	// server is the RPC server the notification manager is associated with.
	server *RPCServer

	// queueNotification queues a notification for handling.
	queueNotification chan interface{}

	// notificationMsgs feeds notificationHandler with notifications
	// and client (un)registeration requests from a queue as well as
	// registeration and unregisteration requests from clients.
	notificationMsgs chan interface{}

	// Access channel for current number of connected clients.
	numClients chan int

	// Shutdown handling
	wg   sync.WaitGroup
	quit chan struct{}
}

// GbtWorkState houses state that is used in between multiple RPC invocations to
// getblocktemplate.
type GbtWorkState struct {
	sync.Mutex
	lastTxUpdate  time.Time
	lastGenerated time.Time
	prevHash      *daghash.Hash
	minTimestamp  time.Time
	template      *mining.BlockTemplate
	NotifyMap     map[daghash.Hash]map[int64]chan struct{}
	TimeSource    blockdag.MedianTimeSource
}

// HelpCacher provides a concurrent safe type that provides help and usage for
// the RPC server commands and caches the results for future calls.
type HelpCacher struct {
	sync.Mutex
	Usage      string
	MethodHelp map[string]string
}

// BroadcastMsg provides the ability to house a bitcoin message to be broadcast
// to all connected peers except specified excluded peers.
type BroadcastMsg struct {
	message      wire.Message
	excludePeers []*Peer
}

// UpdatePeerHeightsMsg is a message sent from the blockmanager to the server
// after a new block has been accepted. The purpose of the message is to update
// the heights of peers that were known to announce the block before we
// connected it to the main chain or recognized it as an orphan. With these
// updates, peer heights will be kept up to date, allowing for fresh data when
// selecting sync peer candidacy.
type UpdatePeerHeightsMsg struct {
	newHash    *daghash.Hash
	newHeight  int32
	originPeer *peer.Peer
}

// CfHeaderKV is a tuple of a filter header and its associated block hash. The
// struct is used to cache cfcheckpt responses.
type CfHeaderKV struct {
	BlockHash    daghash.Hash
	FilterHeader daghash.Hash
}

// RPCServerPeer represents a peer for use with the RPC server.
//
// The interface contract requires that all of these methods are safe for
// concurrent access.
type RPCServerPeer interface {
	// ToPeer returns the underlying peer instance.
	ToPeer() *peer.Peer

	// IsTxRelayDisabled returns whether or not the peer has disabled
	// transaction relay.
	IsTxRelayDisabled() bool

	// BanScore returns the current integer value that represents how close
	// the peer is to being banned.
	BanScore() uint32
}

// BTCDLookup resolves the IP of the given host using the correct DNS lookup
// function depending on the configuration options.  For example, addresses will
// be resolved using tor when the --proxy flag was specified unless --noonion
// was also specified in which case the normal system DNS resolver will be used.
//
// Any attempt to resolve a tor address (.onion) will return an error since they
// are not intended to be resolved outside of the tor proxy.
func BTCDLookup(host string) ([]net.IP, error) {
	if strings.HasSuffix(host, ".onion") {
		return nil, fmt.Errorf("attempt to resolve tor address %s", host)
	}

	return config.MainConfig().Lookup(host)
}

// AddBytesReceived adds the passed number of bytes to the total bytes received
// counter for the server.  It is safe for concurrent access.
func (s *Server) AddBytesReceived(bytesReceived uint64) {
	atomic.AddUint64(&s.BytesReceived, bytesReceived)
}

// ConnectedCount returns the number of currently connected peers.
func (s *Server) ConnectedCount() int32 {
	replyChan := make(chan int32)

	s.Query <- GetConnCountMsg{Reply: replyChan}

	return <-replyChan
}

// NetTotals returns the sum of all bytes received and sent across the network
// for all peers.  It is safe for concurrent access.
func (s *Server) NetTotals() (uint64, uint64) {
	return atomic.LoadUint64(&s.BytesReceived),
		atomic.LoadUint64(&s.BytesSent)
}

// BroadcastMessage sends msg to all peers currently connected to the server
// except those in the passed peers to exclude.
func (s *Server) BroadcastMessage(msg wire.Message, exclPeers ...*Peer) {
	// XXX: Need to determine if this is an alert that has already been
	// broadcast and refrain from broadcasting again.
	bmsg := BroadcastMsg{message: msg, excludePeers: exclPeers}
	s.Broadcast <- bmsg
}

// AddRebroadcastInventory adds 'iv' to the list of inventories to be
// rebroadcasted at random intervals until they show up in a block.
func (s *Server) AddRebroadcastInventory(iv *wire.InvVect, data interface{}) {
	// Ignore if shutting down.
	if atomic.LoadInt32(&s.Shutdown) != 0 {
		return
	}

	s.ModifyRebroadcastInv <- BroadcastInventoryAdd{InvVect: iv, Data: data}
}

type ConnectNodeMsg struct {
	Addr      string
	Permanent bool
	Reply     chan error
}

type RemoveNodeMsg struct {
	Cmp   func(*Peer) bool
	Reply chan error
}

type GetConnCountMsg struct {
	Reply chan int32
}

type GetPeersMsg struct {
	Reply chan []*Peer
}

type GetOutboundGroup struct {
	Key   string
	Reply chan int
}

type GetAddedNodesMsg struct {
	Reply chan []*Peer
}

type DisconnectNodeMsg struct {
	Cmp   func(*Peer) bool
	Reply chan error
}

// BroadcastInventoryAdd is a type used to declare that the InvVect it contains
// needs to be added to the rebroadcast map
type BroadcastInventoryAdd RelayMsg

// RelayTransactions generates and relays inventory vectors for all of the
// passed transactions to all connected peers.
func (s *Server) RelayTransactions(txns []*mempool.TxDesc) {
	for _, txD := range txns {
		iv := wire.NewInvVect(wire.InvTypeTx, txD.Tx.Hash())
		s.RelayInventory(iv, txD)
	}
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (s *Server) RelayInventory(invVect *wire.InvVect, data interface{}) {
	s.RelayInv <- RelayMsg{InvVect: invVect, Data: data}
}
