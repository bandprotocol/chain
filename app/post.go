package band

import (
	feemarketpost "github.com/skip-mev/feemarket/x/feemarket/post"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/app/mempool"
)

// PostHandlerOptions are the options required for constructing a FeeMarket PostHandler.
type PostHandlerOptions struct {
	AccountKeeper            feemarketpost.AccountKeeper
	BankKeeper               feemarketpost.BankKeeper
	FeeMarketKeeper          feemarketpost.FeeMarketKeeper
	IgnorePostDecoratorLanes []*mempool.Lane
}

// NewPostHandler returns a PostHandler chain with the fee deduct decorator.
func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	if !UseFeeMarketDecorator {
		return nil, nil
	}

	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for post builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for post builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "fee market keeper is required for post builder")
	}

	postDecorators := []sdk.PostDecorator{
		NewIgnorePostDecorator(
			feemarketpost.NewFeeMarketDeductDecorator(
				options.AccountKeeper,
				options.BankKeeper,
				options.FeeMarketKeeper,
			),
			options.IgnorePostDecoratorLanes...,
		),
	}

	return sdk.ChainPostDecorators(postDecorators...), nil
}

// IgnorePostDecorator is a post decorator that wraps an existing post decorator. It allows
// for the post decorator to be ignored for specified lanes.
type IgnorePostDecorator struct {
	decorator sdk.PostDecorator
	lanes     []*mempool.Lane
}

// NewIgnorePostDecorator returns a new IgnorePostDecorator instance.
func NewIgnorePostDecorator(decorator sdk.PostDecorator, lanes ...*mempool.Lane) *IgnorePostDecorator {
	return &IgnorePostDecorator{
		decorator: decorator,
		lanes:     lanes,
	}
}

// NewIgnorePostDecorator is a wrapper that implements the sdk.PostDecorator interface,
// providing two execution paths for processing transactions:
// 1. If the transaction contained in any of the provided lanes, the decorator is skipped.
// 2. Otherwise, the wrapped decorator is executed.
func (ig IgnorePostDecorator) PostHandle(
	ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler,
) (sdk.Context, error) {
	// IgnorePostDecorator is only used for check tx.
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate, success)
	}

	for _, lane := range ig.lanes {
		if lane.Contains(tx) {
			return next(ctx, tx, simulate, success)
		}
	}

	return ig.decorator.PostHandle(ctx, tx, simulate, success, next)
}
