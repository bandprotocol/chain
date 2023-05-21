package client

import (
	"context"
	"fmt"
	"time"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Client struct {
	client    rpcclient.Client // RPC client for communication with the node.
	context   client.Context   // Context that holds the client's configuration and context.
	txFactory tx.Factory       // Factory for creating and handling transactions.

	maxTry       uint64        // Maximum number of tries to submit a transaction.
	timeout      time.Duration // Timeout duration for waiting for transaction commits.
	pollInterval time.Duration // Duration between each poll for transaction status.

	gasAdjustStart float64 // Initial value for adjusting the gas price.
	gasAdjustStep  float64 // Step value for adjusting the gas price.
}

// New creates a new instance of the Client.
// It returns the created Client instance and an error if the initialization fails.
func New(cfg *cylinder.Config, kr keyring.Keyring) (*Client, error) {
	// Create a new HTTP client for the specified node URI
	c, err := httpclient.New(cfg.NodeURI, "/websocket")
	if err != nil {
		return nil, err
	}

	// Start the client to establish a connection
	err = c.Start()
	if err != nil {
		return nil, err
	}

	// Create a new client context and configure it with necessary parameters
	ctx := client.Context{}.
		WithClient(c).
		WithChainID(cfg.ChainID).
		WithCodec(band.MakeEncodingConfig().Marshaler).
		WithTxConfig(band.MakeEncodingConfig().TxConfig).
		WithBroadcastMode(flags.BroadcastSync).
		WithInterfaceRegistry(band.MakeEncodingConfig().InterfaceRegistry).
		WithKeyring(kr)

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
		client:       c,
		context:      ctx,
		txFactory:    txf,
		timeout:      cfg.BroadcastTimeout,
		pollInterval: cfg.RPCPollInterval,
		maxTry:       cfg.MaxTry,
		// TODO-CYLINDER: TUNE THESE NUMBERS / MOVE TO CONFIG
		gasAdjustStart: 1.4,
		gasAdjustStep:  0.2,
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
func (c *Client) GetTxFromTxHash(
	txHash string,
) (*sdk.TxResponse, error) {
	var err error
	for start := time.Now(); time.Since(start) < c.timeout; {
		txRes, err := authtx.QueryTx(c.context, txHash)
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
func (c *Client) QueryGroup(groupID tss.GroupID) (*GroupResponse, error) {
	queryClient := types.NewQueryClient(c.context)

	gr, err := queryClient.Group(context.Background(), &types.QueryGroupRequest{
		GroupId: uint64(groupID),
	})
	if err != nil {
		return nil, err
	}

	return NewGroupResponse(gr), nil
}

// BroadcastAndConfirm broadcasts and confirms the messages by signing and submitting them using the provided key.
// It returns the transaction response or an error. It retries broadcasting and confirming up to maxTry times.
func (c *Client) BroadcastAndConfirm(key *keyring.Record, msgs []sdk.Msg) (res *sdk.TxResponse, err error) {
	gasAdjust := c.gasAdjustStart

	for try := uint64(1); try <= c.maxTry; try++ {
		time.Sleep(c.pollInterval)

		// sign and broadcast the messages
		res, err = c.Broadcast(
			key,
			msgs,
			gasAdjust,
		)
		if err != nil {
			continue
		}

		if res.Code == 0 {
			// query transaction to get status
			res, err = c.GetTxFromTxHash(res.TxHash)
			if err != nil {
				continue
			}

			if res.Code == 0 {
				return
			}
		}

		if res.Codespace == sdkerrors.RootCodespace && res.Code == sdkerrors.ErrOutOfGas.ABCICode() {
			gasAdjust += c.gasAdjustStep
		}
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

	txf := c.txFactory.WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithGasAdjustment(gasAdjust)

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

	err = tx.Sign(txf, key.Name, txb, true)
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
