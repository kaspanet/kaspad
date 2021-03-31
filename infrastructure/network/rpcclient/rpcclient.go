package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient/grpcclient"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

const defaultTimeout = 30 * time.Second

// RPCClient is an RPC client
type RPCClient struct {
	*grpcclient.GRPCClient

	rpcAddress     string
	rpcRouter      *rpcRouter
	isConnected    uint32
	isClosed       uint32
	isReconnecting uint32

	timeout time.Duration
}

// NewRPCClient creates a new RPC client
func NewRPCClient(rpcAddress string) (*RPCClient, error) {
	rpcClient := &RPCClient{
		rpcAddress: rpcAddress,
		timeout:    defaultTimeout,
	}
	err := rpcClient.connect()
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}

func (c *RPCClient) connect() error {
	rpcClient, err := grpcclient.Connect(c.rpcAddress)
	if err != nil {
		return errors.Wrapf(err, "error connecting to address %s", c.rpcAddress)
	}
	rpcClient.SetOnDisconnectedHandler(c.handleClientDisconnected)
	rpcClient.SetOnErrorHandler(c.handleClientError)
	rpcRouter, err := buildRPCRouter()
	if err != nil {
		return errors.Wrapf(err, "error creating the RPC router")
	}

	atomic.StoreUint32(&c.isConnected, 1)
	rpcClient.AttachRouter(rpcRouter.router)

	c.GRPCClient = rpcClient
	c.rpcRouter = rpcRouter

	log.Infof("Connected to %s", c.rpcAddress)
	return nil
}

func (c *RPCClient) disconnect() error {
	c.rpcRouter.router.Close()
	err := c.GRPCClient.Disconnect()
	if err != nil {
		return err
	}
	log.Infof("Disconnected from %s", c.rpcAddress)
	return nil
}

func (c *RPCClient) Reconnect() error {
	if atomic.LoadUint32(&c.isClosed) == 1 {
		return errors.Errorf("Cannot reconnect from a closed client")
	}

	// Protect against multiple threads attempting to reconnect at the same time
	swapped := atomic.CompareAndSwapUint32(&c.isReconnecting, 0, 1)
	if !swapped {
		// Already reconnecting
		return nil
	}
	defer atomic.StoreUint32(&c.isReconnecting, 0)

	log.Warnf("Attempting to reconnect to %s", c.rpcAddress)

	// Disconnect if we're connected
	if atomic.LoadUint32(&c.isConnected) == 1 {
		err := c.disconnect()
		if err != nil {
			return err
		}
	}

	// Attempt to connect until we succeed
	for {
		err := c.connect()
		if err == nil {
			return nil
		}
		log.Warnf("Could not automatically reconnect to %s: %s", c.rpcAddress, err)

		const retryDelay = 10 * time.Second
		log.Warnf("Retrying in %s", retryDelay)
		time.Sleep(retryDelay)
	}
}

func (c *RPCClient) handleClientDisconnected() {
	atomic.StoreUint32(&c.isConnected, 0)
	if atomic.LoadUint32(&c.isClosed) == 0 {
		err := c.Reconnect()
		if err != nil {
			panic(err)
		}
	}
}

func (c *RPCClient) handleClientError(err error) {
	log.Warnf("Received error from client: %s", err)
	c.handleClientDisconnected()
}

// SetTimeout sets the timeout by which to wait for RPC responses
func (c *RPCClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Close closes the RPC client
func (c *RPCClient) Close() error {
	swapped := atomic.CompareAndSwapUint32(&c.isClosed, 0, 1)
	if !swapped {
		return errors.Errorf("Cannot close a client that had already been closed")
	}
	c.rpcRouter.router.Close()
	return nil
}

// Address returns the address the RPC client connected to
func (c *RPCClient) Address() string {
	return c.rpcAddress
}

func (c *RPCClient) route(command appmessage.MessageCommand) *routerpkg.Route {
	return c.rpcRouter.routes[command]
}

// ErrRPC is an error in the RPC protocol
var ErrRPC = errors.New("rpc error")

func (c *RPCClient) convertRPCError(rpcError *appmessage.RPCError) error {
	return errors.Wrap(ErrRPC, rpcError.Message)
}

// SetLogger uses a specified Logger to output package logging info
func (c *RPCClient) SetLogger(backend *logger.Backend, level logger.Level) {
	const logSubsystem = "RPCC"
	log = backend.Logger(logSubsystem)
	log.SetLevel(level)
	spawn = panics.GoroutineWrapperFunc(log)
}
