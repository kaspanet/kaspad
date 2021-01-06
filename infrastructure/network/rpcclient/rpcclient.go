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

	rpcAddress string
	rpcRouter  *rpcRouter

	timeout time.Duration
}

// NewRPCClient creates a new RPC client
func NewRPCClient(rpcAddress string) (*RPCClient, error) {
	rpcClient, err := grpcclient.Connect(rpcAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to address %s", rpcAddress)
	}
	rpcRouter, err := buildRPCRouter()
	if err != nil {
		return nil, errors.Wrapf(err, "error creating the RPC router")
	}
	rpcClient.AttachRouter(rpcRouter.router)

	log.Infof("Connected to server %s", rpcAddress)

	return &RPCClient{
		GRPCClient: rpcClient,
		rpcAddress: rpcAddress,
		rpcRouter:  rpcRouter,
		timeout:    defaultTimeout,
	}, nil
}

// SetTimeout sets the timeout by which to wait for RPC responses
func (c *RPCClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Close closes the RPC client
func (c *RPCClient) Close() {
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
