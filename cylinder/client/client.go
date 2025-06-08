package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	cylinderctx "github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

type Client struct {
	client    rpcclient.Client // RPC client for communication with the node.
	context   client.Context   // Context that holds the client's configuration and context.
	txFactory tx.Factory       // Factory for creating and handling transactions.

	maxTry       uint64        // Maximum number of tries to submit a transaction and query.
	timeout      time.Duration // Timeout duration for waiting for transaction commits.
	pollInterval time.Duration // Duration between each poll for transaction status or query result.

	gasAdjustStart float64 // Initial value for adjusting the gas price.
	gasAdjustStep  float64 // Step value for adjusting the gas price.
}

// New creates a new instance of the Client.
// It returns the created Client instance and an error if the initialization fails.
func New(cylinderCtx *cylinderctx.Context) (*Client, error) {
	cfg := cylinderCtx.Config

	// Create a new HTTP client for the specified node URI
	c, err := httpclient.New(cfg.NodeURI, "/websocket")
	if err != nil {
		return nil, err
	}

	// Start the client to establish a connection
	if err = c.Start(); err != nil {
		return nil, err
	}

	// Create a new client context and configure it with necessary parameters
	ctx := client.Context{}.
		WithClient(c).
		WithChainID(cfg.ChainID).
		WithCodec(cylinderCtx.Cdc).
		WithTxConfig(cylinderCtx.TxConfig).
		WithBroadcastMode(flags.BroadcastSync).
		WithInterfaceRegistry(cylinderCtx.InterfaceRegistry).
		WithKeyring(cylinderCtx.Keyring)

	// Create a new transaction factory and configure it with necessary parameters
	txf := tx.Factory{}.
		WithTxConfig(ctx.TxConfig).
		WithChainID(ctx.ChainID).
		WithKeybase(ctx.Keyring).
		WithAccountRetriever(ctx.AccountRetriever).
		WithGasPrices(cfg.GasPrices).
		WithSimulateAndExecute(true)

	// Create and return the Client instance with the initialized fields
	return &Client{
		client:         c,
		context:        ctx,
		txFactory:      txf,
		timeout:        cfg.BroadcastTimeout,
		pollInterval:   cfg.RPCPollInterval,
		maxTry:         cfg.MaxTry,
		gasAdjustStart: cfg.GasAdjustStart,
		gasAdjustStep:  cfg.GasAdjustStep,
	}, nil
}

// Subscribe subscribes to an event query with the provided subscriber and query string.
// It returns a channel of ResultEvent to receive the subscribed events and an error if any.
func (c *Client) Subscribe(subscriber, query string, outCapacity ...int) (out <-chan ctypes.ResultEvent, err error) {
	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()

	return c.client.Subscribe(ctx, subscriber, query, outCapacity...)
}

// GetTxFromTxHash retrieves the transaction response for the given transaction hash.
// It waits for the transaction to be committed and returns the transaction response or an error if it exceeds timeout.
func (c *Client) GetTxFromTxHash(txHash string) (txRes *sdk.TxResponse, err error) {
	for start := time.Now(); time.Since(start) < c.timeout; {
		txRes, err = authtx.QueryTx(c.context, txHash)
		if err != nil {
			time.Sleep(c.pollInterval)
			continue
		}

		return txRes, nil
	}

	return nil, err
}

// QueryGroup queries the group information with the given group ID.
// It returns the group response or an error.
func (c *Client) QueryGroup(groupID tss.GroupID) (*GroupResult, error) {
	queryClient := tsstypes.NewQueryClient(c.context)
	input := &tsstypes.QueryGroupRequest{
		GroupId: uint64(groupID),
	}

	gr, err := queryWithRetry(queryClient.Group, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return NewGroupResult(gr), nil
}

// QuerySigning queries the signing information with the given signing ID.
// It returns the signing response or an error.
func (c *Client) QuerySigning(signingID tss.SigningID) (*SigningResponse, error) {
	queryClient := tsstypes.NewQueryClient(c.context)
	input := &tsstypes.QuerySigningRequest{
		SigningId: uint64(signingID),
	}

	sr, err := queryWithRetry(queryClient.Signing, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return NewSigningResponse(sr), nil
}

// QueryDE queries the DE information with the given address.
// It returns the de response or an error.
func (c *Client) QueryDE(address string, offset uint64, limit uint64) (*DEResponse, error) {
	queryClient := tsstypes.NewQueryClient(c.context)
	input := &tsstypes.QueryDERequest{
		Address: address,
		Pagination: &query.PageRequest{
			Offset:     offset,
			Limit:      limit,
			CountTotal: true,
		},
	}

	der, err := queryWithRetry(queryClient.DE, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return NewDEResponse(der), nil
}

// QueryAllDE queries all DEs with the given address.
func (c *Client) QueryAllDE(address string) ([]tsstypes.DE, error) {
	des := make([]tsstypes.DE, 0)
	queryClient := tsstypes.NewQueryClient(c.context)

	var nextKey []byte
	for {
		input := &tsstypes.QueryDERequest{
			Address: address,
			Pagination: &query.PageRequest{
				Key: nextKey,
			},
		}

		res, err := queryWithRetry(queryClient.DE, input, c.maxTry, c.pollInterval)
		if err != nil {
			return nil, err
		}

		des = append(des, res.DEs...)

		nextKey = res.GetPagination().GetNextKey()
		if len(nextKey) == 0 {
			break
		}
	}

	return des, nil
}

// QueryMember queries the member information of the given address.
// It returns the member information on current and incoming group or an error.
func (c *Client) QueryMember(address string) (*bandtsstypes.QueryMemberResponse, error) {
	queryClient := bandtsstypes.NewQueryClient(c.context)
	input := &bandtsstypes.QueryMemberRequest{
		Address: address,
	}

	res, err := queryWithRetry(queryClient.Member, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryPendingGroups queries the all pending groups with the given address.
// It returns the QueryPendingSignsResponse or an error.
func (c *Client) QueryPendingGroups(address string) (*tsstypes.QueryPendingGroupsResponse, error) {
	queryClient := tsstypes.NewQueryClient(c.context)
	input := &tsstypes.QueryPendingGroupsRequest{
		Address: address,
	}

	res, err := queryWithRetry(queryClient.PendingGroups, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryPendingSignings queries the all pending signings with the given address.
// It returns the QueryPendingSignsResponse or an error.
func (c *Client) QueryPendingSignings(address string) (*tsstypes.QueryPendingSigningsResponse, error) {
	queryClient := tsstypes.NewQueryClient(c.context)
	input := &tsstypes.QueryPendingSigningsRequest{
		Address: address,
	}

	res, err := queryWithRetry(queryClient.PendingSignings, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryTssParams queries the current tss parameters.
func (c *Client) QueryTssParams() (*tsstypes.Params, error) {
	queryClient := tsstypes.NewQueryClient(c.context)
	input := &tsstypes.QueryParamsRequest{}

	res, err := queryWithRetry(queryClient.Params, input, c.maxTry, c.pollInterval)
	if err != nil {
		return nil, err
	}

	return &res.Params, nil
}

// BroadcastAndConfirm broadcasts and confirms the messages by signing and submitting them using the provided key.
// It returns the transaction response or an error. It retries broadcasting and confirming up to maxTry times.
func (c *Client) BroadcastAndConfirm(
	logger *logger.Logger,
	key *keyring.Record,
	msgs []sdk.Msg,
) (res *sdk.TxResponse, err error) {
	gasAdjust := c.gasAdjustStart

	for try := uint64(1); try <= c.maxTry; try++ {
		// sign and broadcast the messages
		res, err = c.Broadcast(
			key,
			msgs,
			gasAdjust,
		)
		time.Sleep(c.pollInterval)
		if err != nil {
			logger.Debug(":anxious_face_with_sweat: Try %d: Failed to broadcast msgs with error: %s", try, err.Error())
			continue
		}

		if res.Code == 0 {
			// query transaction to get status
			res, err = c.GetTxFromTxHash(res.TxHash)
			if err != nil {
				logger.Debug(
					":anxious_face_with_sweat: Try %d: Failed to get tx from hash with error: %s",
					try,
					err.Error(),
				)
				continue
			}

			if res.Code == 0 {
				return
			}
		}

		if res.Codespace == sdkerrors.RootCodespace && res.Code == sdkerrors.ErrOutOfGas.ABCICode() {
			gasAdjust += c.gasAdjustStep
			logger.Debug(
				":anxious_face_with_sweat: Try %d: Bumping gas since tx is out of gas: new gad adjustment %f",
				try,
				gasAdjust,
			)
		}

		logger.Debug(
			":anxious_face_with_sweat: Try %d: Transaction is not successful with error code: codespace: %s, code: %d",
			try,
			res.Codespace,
			res.Code,
		)
	}

	return
}

// Broadcast signs and broadcasts the provided messages using the given key.
// It adjusts the gas according to the gasAdjust parameter and returns the transaction response or an error.
func (c *Client) Broadcast(key *keyring.Record, msgs []sdk.Msg, gasAdjust float64) (*sdk.TxResponse, error) {
	acc, err := c.QueryAccount(key)
	if err != nil {
		return nil, fmt.Errorf("unable to get account: %w", err)
	}

	address, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	memo := fmt.Sprintf("cylinder: %s", version.Version)

	txf := c.txFactory.WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithGasAdjustment(gasAdjust).
		WithFromName(key.Name).
		WithMemo(memo)

	execMsg := authz.NewMsgExec(address, msgs)

	_, adjusted, err := tx.CalculateGas(c.context, txf, &execMsg)
	if err != nil {
		return nil, err
	}

	// Set the gas amount on the transaction factory
	txf = txf.WithGas(adjusted)

	txb, err := txf.BuildUnsignedTx(&execMsg)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(context.Background(), txf, key.Name, txb, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := c.context.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	// Broadcast to a node
	res, err := c.context.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryAccount queries the account information associated with the given key.
// It returns the account or an error if the account retrieval fails.
func (c *Client) QueryAccount(key *keyring.Record) (client.Account, error) {
	address, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	acc, err := authtypes.AccountRetriever{}.GetAccount(c.context, address)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// Stop stops the client by terminating the underlying RPC client connection.
// It returns an error if the client cannot be stopped.
func (c *Client) Stop() error {
	return c.client.Stop()
}

// queryWithRetry performs a query with retry and sleeps for the given poll interval if the query fails.
func queryWithRetry[T any, I any](
	queryFunc func(ctx context.Context, input I, opts ...grpc.CallOption) (T, error),
	input I,
	maxTry uint64,
	pollInterval time.Duration,
) (res T, err error) {
	for try := uint64(1); try <= maxTry; try++ {
		res, err = queryFunc(context.Background(), input)
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}

		return res, nil
	}

	return res, err
}
