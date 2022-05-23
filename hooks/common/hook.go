package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

type Hooks []Hook

// Hook is an interface of hook that can process along with abci application
type Hook interface {
	AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, res abci.ResponseInitChain)
	AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock)
	AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx)
	AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock)
	RequestSearch(req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, bool, error)
	RequestPrice(req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, bool, error)
	BeforeCommit()
}

func (h Hooks) AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, res abci.ResponseInitChain) {
	for _, hook := range h {
		hook.AfterInitChain(ctx, req, res)
	}
}

func (h Hooks) AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	for _, hook := range h {
		hook.AfterBeginBlock(ctx, req, res)
	}
}

func (h Hooks) AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
	for _, hook := range h {
		hook.AfterDeliverTx(ctx, req, res)
	}
}

func (h Hooks) AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) {
	for _, hook := range h {
		hook.AfterEndBlock(ctx, req, res)
	}
}

func (h Hooks) BeforeCommit() {
	for _, hook := range h {
		hook.BeforeCommit()
	}
}

func (h Hooks) RequestSearch(req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, bool, error) {
	for _, hook := range h {
		res, hit, err := hook.RequestSearch(req)
		if hit {
			return res, true, err
		}
	}

	return nil, false, nil
}

func (h Hooks) RequestPrice(req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, bool, error) {
	for _, hook := range h {
		res, hit, err := hook.RequestPrice(req)
		if hit {
			return res, true, err
		}
	}

	return nil, false, nil
}
