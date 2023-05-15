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
	client    rpcclient.Client
	context   client.Context
	txFactory tx.Factory

	maxTry       uint64
	timeout      time.Duration
	pollInterval time.Duration

	gasAdjustStart float64
	gasAdjustStep  float64
}

// TODO-CYLINDER: TBD: SHOULD ADD LOG IN THIS LEVEL? e.g. DEBUG
func New(cfg *cylinder.Config, kr keyring.Keyring) (*Client, error) {
	c, err := httpclient.New(cfg.NodeURI, "/websocket")
	if err != nil {
		return nil, err
	}

	err = c.Start()
	if err != nil {
		return nil, err
	}

	ctx := client.Context{}.
		WithClient(c).
		WithChainID(cfg.ChainID).
		WithCodec(band.MakeEncodingConfig().Marshaler).
		WithTxConfig(band.MakeEncodingConfig().TxConfig).
		WithBroadcastMode(flags.BroadcastSync).
		WithInterfaceRegistry(band.MakeEncodingConfig().InterfaceRegistry).
		WithKeyring(kr)

	txf := tx.Factory{}.
		WithTxConfig(ctx.TxConfig).
		WithChainID(ctx.ChainID).
		WithKeybase(ctx.Keyring).
		WithAccountRetriever(ctx.AccountRetriever).
		WithGasPrices(cfg.GasPrices).
		WithSimulateAndExecute(true)

	return &Client{
		client:       c,
		context:      ctx,
		txFactory:    txf,
		timeout:      cfg.BroadcastTimeout,
		pollInterval: cfg.RPCPollInterval,
		maxTry:       cfg.MaxTry,
		// TODO-CYLINDER: REVISIT TO THINK IF SHOULD IN CONFIG OR NOT
		gasAdjustStart: 1.4,
		gasAdjustStep:  0.2,
	}, nil
}

func (c *Client) Subscribe(subscriber, query string, outCapacity ...int) (out <-chan ctypes.ResultEvent, err error) {
	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()

	return c.client.Subscribe(ctx, subscriber, query, outCapacity...)
}

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

func (c *Client) QueryGroup(
	groupID tss.GroupID,
) (*types.QueryGroupResponse, error) {
	queryClient := types.NewQueryClient(c.context)
	return queryClient.Group(context.Background(), &types.QueryGroupRequest{
		GroupId: uint64(groupID),
	})
}

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

func (c *Client) Broadcast(
	key *keyring.Record, msgs []sdk.Msg, gasAdjust float64,
) (*sdk.TxResponse, error) {
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

	// broadcast to a Tendermint node
	res, err := c.context.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

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

func (c *Client) Stop() error {
	return c.client.Stop()
}
