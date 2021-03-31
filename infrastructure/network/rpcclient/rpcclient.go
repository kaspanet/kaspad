package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient/grpcclient"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"time"
)

const defaultTimeout = 30 * time.Second

// RPCClient is an RPC client
type RPCClient struct {
	*grpcclient.GRPCClient

	rpcAddress      string
	rpcRouter       *rpcRouter
	shouldReconnect bool

	timeout time.Duration
}

// NewRPCClient creates a new RPC client
func NewRPCClient(rpcAddress string) (*RPCClient, error) {
	rpcClient := &RPCClient{
		rpcAddress:      rpcAddress,
		timeout:         defaultTimeout,
		shouldReconnect: true,
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
	rpcRouter, err := buildRPCRouter()
	if err != nil {
		return errors.Wrapf(err, "error creating the RPC router")
	}
	rpcClient.AttachRouter(rpcRouter.router)

	c.GRPCClient = rpcClient
	c.rpcRouter = rpcRouter

	log.Infof("Connected to server %s", c.rpcAddress)
	return nil
}

func (c *RPCClient) disconnect() error {
	c.rpcRouter.router.Close()
	return c.GRPCClient.Disconnect()
}

func (c *RPCClient) Reconnect() error {
	if !c.shouldReconnect {
		return errors.Errorf("Cannot reconnect to a closed client")
	}
	return c.disconnect()
}

func (c *RPCClient) handleClientDisconnected() {
	if c.shouldReconnect {
		for {
			err := c.connect()
			if err == nil {
				return
			}
			log.Warnf("Could not automatically reconnect to %s: %s", c.rpcAddress, err)

			const retryDelay = 10 * time.Second
			log.Warnf("Retrying in %s", retryDelay)
			time.Sleep(retryDelay)
		}
	}
}

// SetTimeout sets the timeout by which to wait for RPC responses
func (c *RPCClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Close closes the RPC client
func (c *RPCClient) Close() {
	c.shouldReconnect = false
	c.rpcRouter.router.Close()
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
