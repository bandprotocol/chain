package keeper

import (
	"context"
	"fmt"
	"sort"

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

	// calculate power different of each signal by decresing signal power with previous signal
	signalIDToPowerDiff := ms.CalculateDelegatorSignalsPowerDiff(ctx, delegator, req.Signals)

	// sort keys to guarantee order of signalIDToPowerDiff iteration
	keys := make([]string, 0, len(signalIDToPowerDiff))
	for k := range signalIDToPowerDiff {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, signalID := range keys {
		powerDiff := signalIDToPowerDiff[signalID]
		feed, err := ms.GetFeed(ctx, signalID)
		if err != nil {
			feed = types.Feed{
				SignalID:                    signalID,
				Power:                       0,
				Interval:                    0,
				LastIntervalUpdateTimestamp: ctx.BlockTime().Unix(),
			}
		} else {
			// before changing in feed, delete feed by power index
			ms.DeleteFeedByPowerIndex(ctx, feed)
		}
		feed.Power += powerDiff

		if feed.Power < 0 {
			return nil, types.ErrPowerNegative
		}

		feed.Interval, feed.DeviationInThousandth = calculateIntervalAndDeviation(feed.Power, ms.GetParams(ctx))
		ms.SetFeed(ctx, feed)
	}

	return &types.MsgSubmitSignalsResponse{}, nil
}

func (ms msgServer) SubmitPrices(
	goCtx context.Context,
	req *types.MsgSubmitPrices,
) (*types.MsgSubmitPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime().Unix()

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	// check if it's in top bonded validators.
	err = ms.ValidateSubmitPricesRequest(ctx, blockTime, req)
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdatePriceService,
			sdk.NewAttribute(types.AttributeKeyHash, req.PriceService.Hash),
			sdk.NewAttribute(types.AttributeKeyVersion, req.PriceService.Version),
			sdk.NewAttribute(types.AttributeKeyURL, req.PriceService.Url),
		),
	)

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

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, req.Params.String()),
	))

	return &types.MsgUpdateParamsResponse{}, nil
}
