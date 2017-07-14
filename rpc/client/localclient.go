package client

import (
	"context"

	"github.com/pkg/errors"

	data "github.com/tendermint/go-wire/data"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	tmquery "github.com/tendermint/tmlibs/pubsub/query"
)

/*
Local is a Client implementation that directly executes the rpc
functions on a given node, without going through HTTP or GRPC.

This implementation is useful for:

* Running tests against a node in-process without the overhead
of going through an http server
* Communication between an ABCI app and Tendermint core when they
are compiled in process.

For real clients, you probably want to use client.HTTP.  For more
powerful control during testing, you probably want the "client/mock" package.
*/
type Local struct {
	node *nm.Node

	*types.EventBus
	subscriptions map[string]*tmquery.Query
}

// NewLocal configures a client that calls the Node directly.
//
// Note that given how rpc/core works with package singletons, that
// you can only have one node per process.  So make sure test cases
// don't run in parallel, or try to simulate an entire network in
// one process...
func NewLocal(node *nm.Node) *Local {
	node.ConfigureRPC()
	return &Local{
		node:          node,
		EventBus:      node.EventBus(),
		subscriptions: make(map[string]*tmquery.Query),
	}
}

func (c *Local) _assertIsClient() Client {
	return c
}

func (c *Local) _assertIsNetworkClient() NetworkClient {
	return c
}

func (c *Local) _assertIsEventsClient() EventsClient {
	return c
}

func (Local) Status() (*ctypes.ResultStatus, error) {
	return core.Status()
}

func (Local) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return core.ABCIInfo()
}

func (Local) ABCIQuery(path string, data data.Bytes, prove bool) (*ctypes.ResultABCIQuery, error) {
	return core.ABCIQuery(path, data, prove)
}

func (Local) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return core.BroadcastTxCommit(tx)
}

func (Local) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return core.BroadcastTxAsync(tx)
}

func (Local) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return core.BroadcastTxSync(tx)
}

func (Local) NetInfo() (*ctypes.ResultNetInfo, error) {
	return core.NetInfo()
}

func (Local) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	return core.DumpConsensusState()
}

func (Local) DialSeeds(seeds []string) (*ctypes.ResultDialSeeds, error) {
	return core.UnsafeDialSeeds(seeds)
}

func (Local) BlockchainInfo(minHeight, maxHeight int) (*ctypes.ResultBlockchainInfo, error) {
	return core.BlockchainInfo(minHeight, maxHeight)
}

func (Local) Genesis() (*ctypes.ResultGenesis, error) {
	return core.Genesis()
}

func (Local) Block(height int) (*ctypes.ResultBlock, error) {
	return core.Block(height)
}

func (Local) Commit(height int) (*ctypes.ResultCommit, error) {
	return core.Commit(height)
}

func (Local) Validators() (*ctypes.ResultValidators, error) {
	return core.Validators()
}

func (Local) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	return core.Tx(hash, prove)
}

func (c *Local) Subscribe(ctx context.Context, query string, out chan<- interface{}) error {
	q, err := tmquery.New(query)
	if err != nil {
		return errors.Wrap(err, "failed to parse query")
	}
	c.subscriptions[query] = q
	return c.EventBus.Subscribe(ctx, "rpclocalclient", q, out)
}

func (c *Local) Unsubscribe(ctx context.Context, query string) error {
	q, ok := c.subscriptions[query]
	if ok {
		return c.EventBus.Unsubscribe(ctx, "rpclocalclient", q)
	}
	return nil
}

func (c *Local) UnsubscribeAll(ctx context.Context) error {
	return c.EventBus.UnsubscribeAll(ctx, "rpclocalclient")
}
