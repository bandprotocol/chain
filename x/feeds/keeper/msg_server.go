package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/feeds MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

func (ms msgServer) SignalSymbols(
	goCtx context.Context,
	req *types.MsgSignalSymbols,
) (*types.MsgSignalSymbolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegator, err := sdk.AccAddressFromBech32(req.Delegator)
	if err != nil {
		return nil, err
	}

	// check whether delegator has enough delegation for signals
	err = ms.CheckDelegatorDelegation(ctx, delegator, req.Signals)
	if err != nil {
		return nil, err
	}

	// delete previous signal, decrease symbol power by the previous signals
	symbolToIntervalDiff := make(map[string]int64)
	symbolToIntervalDiff, err = ms.RemoveDelegatorPreviousSignals(ctx, delegator, symbolToIntervalDiff)
	if err != nil {
		return nil, err
	}

	// increase symbol power by the new signals
	symbolToIntervalDiff, err = ms.RegisterDelegatorSignals(ctx, delegator, req.Signals, symbolToIntervalDiff)
	if err != nil {
		return nil, err
	}

	// update interval timestamp for interval-changed symbols
	err = ms.UpdateSymbolIntervalTimestamp(ctx, symbolToIntervalDiff)
	if err != nil {
		return nil, err
	}

	// emit events for the signaling operation.
	for _, signal := range req.Signals {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSignalSymbols,
				sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
				sdk.NewAttribute(types.AttributeKeySymbol, signal.Symbol),
				sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", signal.Power)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
	}

	return &types.MsgSignalSymbolsResponse{}, nil
}

func (ms msgServer) SubmitPrices(
	goCtx context.Context,
	req *types.MsgSubmitPrices,
) (*types.MsgSubmitPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime().Unix()

	// check if it's in top bonded validators.
	err := ms.ValidateSubmitPricesRequest(ctx, blockTime, req)
	if err != nil {
		return nil, err
	}

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	transitionTime := ms.GetParams(ctx).TransitionTime

	for _, price := range req.Prices {
		priceVal, err := ms.NewPriceValidator(ctx, blockTime, price, val, transitionTime)
		if err != nil {
			return nil, err
		}

		_ = ms.SetPriceValidator(ctx, priceVal)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitPrice,
				sdk.NewAttribute(types.AttributeKeyValidator, priceVal.Validator),
				sdk.NewAttribute(types.AttributeKeySymbol, priceVal.Symbol),
				sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", priceVal.Price)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", priceVal.Timestamp)),
			),
		)
	}

	return &types.MsgSubmitPricesResponse{}, nil
}

func (ms msgServer) UpdatePriceService(
	goCtx context.Context,
	req *types.MsgUpdatePriceService,
) (*types.MsgUpdatePriceServiceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin := ms.GetParams(ctx).Admin
	if admin != req.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			req.Admin,
		)
	}

	if err := ms.SetPriceService(ctx, req.PriceService); err != nil {
		return nil, err
	}

	return &types.MsgUpdatePriceServiceResponse{}, nil
}

func (ms msgServer) UpdateParams(
	goCtx context.Context,
	req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if ms.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			ms.authority,
			req.Authority,
		)
	}

	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
