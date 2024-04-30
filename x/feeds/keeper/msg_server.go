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

func (ms msgServer) SubmitSignals(
	goCtx context.Context,
	req *types.MsgSubmitSignals,
) (*types.MsgSubmitSignalsResponse, error) {
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

	// delete previous signal, decrease feed power by the previous signals
	signalIDToIntervalDiff := make(map[string]int64)
	prevSignals := ms.GetDelegatorSignals(ctx, delegator)
	signalIDToIntervalDiff, err = ms.RemoveSignals(ctx, prevSignals, signalIDToIntervalDiff)
	if err != nil {
		return nil, err
	}

	// increase feed power by the new signals
	signalIDToIntervalDiff, err = ms.RegisterDelegatorSignals(ctx, delegator, req.Signals, signalIDToIntervalDiff)
	if err != nil {
		return nil, err
	}

	// update interval timestamp for interval-changed signal ids
	ms.UpdateFeedIntervalTimestamp(ctx, signalIDToIntervalDiff)

	// delete feed that has zero power
	for _, signal := range prevSignals {
		feed, err := ms.GetFeed(ctx, signal.ID)
		if err != nil {
			// if feed is not existed, no need to delete
			continue
		}
		if feed.Power == 0 {
			ms.DeleteFeed(ctx, feed.SignalID)
			ms.DeleteFeedByPowerIndex(ctx, feed)
		}
	}

	// emit events for the signaling operation.
	for _, signal := range req.Signals {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitSignals,
				sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
				sdk.NewAttribute(types.AttributeKeySignalID, signal.ID),
				sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", signal.Power)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
	}

	return &types.MsgSubmitSignalsResponse{}, nil
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

	cooldownTime := ms.GetParams(ctx).CooldownTime

	for _, price := range req.Prices {
		priceVal, err := ms.NewPriceValidator(ctx, blockTime, price, val, cooldownTime)
		if err != nil {
			return nil, err
		}

		err = ms.SetPriceValidator(ctx, priceVal)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitPrice,
				sdk.NewAttribute(types.AttributeKeyPriceOption, priceVal.PriceOption.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, priceVal.Validator),
				sdk.NewAttribute(types.AttributeKeySignalID, priceVal.SignalID),
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
