package common

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

type Hooks []Hook

// Hook is an interface of hook that can process along with abci application
type Hook interface {
	AfterInitChain(ctx sdk.Context, req *abci.RequestInitChain, res *abci.ResponseInitChain)
	// Receive context that has been finalized on that block
	AfterBeginBlock(ctx sdk.Context, req *abci.RequestFinalizeBlock, events []abci.Event)
	AfterDeliverTx(ctx sdk.Context, tx sdk.Tx, res *abci.ExecTxResult)
	AfterEndBlock(ctx sdk.Context, events []abci.Event)
	RequestSearch(req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, bool, error)
	RequestPrice(req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, bool, error)
	BeforeCommit()
}

func (h Hooks) AfterInitChain(ctx sdk.Context, req *abci.RequestInitChain, res *abci.ResponseInitChain) {
	for _, hook := range h {
		hook.AfterInitChain(ctx, req, res)
	}
}

func (h Hooks) AfterBeginBlock(ctx sdk.Context, req *abci.RequestFinalizeBlock, events []abci.Event) {
	for _, hook := range h {
		hook.AfterBeginBlock(ctx, req, events)
	}
}

func (h Hooks) AfterDeliverTx(ctx sdk.Context, tx sdk.Tx, res *abci.ExecTxResult) {
	for _, hook := range h {
		hook.AfterDeliverTx(ctx, tx, res)
	}
}

func (h Hooks) AfterEndBlock(ctx sdk.Context, events []abci.Event) {
	for _, hook := range h {
		hook.AfterEndBlock(ctx, events)
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
