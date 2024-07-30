package keeper

import (
	"context"
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

// SubmitSignals registers new signals and updates feeds.
func (ms msgServer) SubmitSignals(
	goCtx context.Context,
	req *types.MsgSubmitSignals,
) (*types.MsgSubmitSignalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the delegator address from Bech32 format to sdk.AccAddress
	delegator, err := sdk.AccAddressFromBech32(req.Delegator)
	if err != nil {
		return nil, err
	}

	// Check if the number of submitted signals exceeds the maximum allowed feeds
	if len(req.Signals) > int(ms.GetParams(ctx).MaxCurrentFeeds) {
		return nil, types.ErrSubmittedSignalsTooLarge.Wrapf(
			"maximum number of signals is %d but received %d",
			ms.GetParams(ctx).MaxCurrentFeeds, len(req.Signals),
		)
	}

	// Check whether the delegator has enough delegation for the submitted signals
	err = ms.CheckDelegatorDelegation(ctx, delegator, req.Signals)
	if err != nil {
		return nil, err
	}

	// RegisterNewSignals delete previous signals and register new signals then returns feed power differences
	signalIDToPowerDiff, err := ms.RegisterNewSignals(ctx, delegator, req.Signals)
	if err != nil {
		return nil, err
	}

	// Sort keys to guarantee order of signalIDToPowerDiff iteration
	keys := make([]string, 0, len(signalIDToPowerDiff))
	for k := range signalIDToPowerDiff {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Iterate over each signal ID, updating the total power and ensuring no negative power values
	for _, signalID := range keys {
		// Get the power difference for the current signal ID
		powerDiff := signalIDToPowerDiff[signalID]

		// Retrieve the total power of the current signal ID from the store
		signalTotalPower, err := ms.GetSignalTotalPower(ctx, signalID)
		if err != nil {
			// Initialize a new signal with zero power if the signal ID does not exist
			signalTotalPower = types.Signal{
				ID:    signalID,
				Power: 0,
			}
		}

		// Update the total power of the signal by adding the power difference
		signalTotalPower.Power += powerDiff

		// Ensure the total power of the signal is not negative
		if signalTotalPower.Power < 0 {
			return nil, types.ErrPowerNegative
		}

		// Save the updated signal total power back to the store
		ms.SetSignalTotalPower(ctx, signalTotalPower)
	}

	// Return an empty response indicating success
	return &types.MsgSubmitSignalsResponse{}, nil
}

// SubmitSignalPrices register new validator prices.
func (ms msgServer) SubmitSignalPrices(
	goCtx context.Context,
	req *types.MsgSubmitSignalPrices,
) (*types.MsgSubmitSignalPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime().Unix()
	blockHeight := ctx.BlockHeight()

	if len(req.Prices) > int(ms.Keeper.GetParams(ctx).MaxCurrentFeeds) {
		return nil, types.ErrSignalPricesTooLarge
	}

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	// check if it's in top bonded validators.
	err = ms.ValidateSubmitSignalPricesRequest(ctx, blockTime, req, val)
	if err != nil {
		return nil, err
	}

	cooldownTime := ms.GetParams(ctx).CooldownTime
	supportedFeeds := ms.GetCurrentFeeds(ctx)
	supportedFeedsMap := make(map[string]int)
	for idx, feed := range supportedFeeds.Feeds {
		supportedFeedsMap[feed.SignalID] = idx
	}

	valPrices := make([]types.ValidatorPrice, len(supportedFeedsMap))
	prevValPrices, err := ms.GetValidatorPriceList(ctx, val)
	if err == nil {
		for _, valPrice := range prevValPrices.ValidatorPrices {
			idx, ok := supportedFeedsMap[valPrice.SignalID]
			if ok {
				valPrices[idx] = valPrice
			}
		}
	}

	for _, price := range req.Prices {
		idx, ok := supportedFeedsMap[price.SignalID]
		if !ok {
			return nil, types.ErrSignalIDNotSupported.Wrapf(
				"signal_id: %s",
				price.SignalID,
			)
		}

		// check if price is not too fast
		valPrice := valPrices[idx]
		if valPrice.SignalID != "" && blockTime < valPrice.Timestamp+cooldownTime {
			return nil, types.ErrPriceSubmitTooEarly
		}

		valPrice = ms.NewValidatorPrice(val, price, blockTime, blockHeight)
		valPrices[idx] = valPrice
		emitEventSubmitSignalPrice(ctx, valPrice)
	}

	if err := ms.SetValidatorPriceList(ctx, val, valPrices); err != nil {
		return nil, err
	}

	return &types.MsgSubmitSignalPricesResponse{}, nil
}

// UpdateReferenceSourceConfig updates reference source configuration.
func (ms msgServer) UpdateReferenceSourceConfig(
	goCtx context.Context,
	req *types.MsgUpdateReferenceSourceConfig,
) (*types.MsgUpdateReferenceSourceConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin := ms.GetParams(ctx).Admin
	if admin != req.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			req.Admin,
		)
	}

	if err := ms.SetReferenceSourceConfig(ctx, req.ReferenceSourceConfig); err != nil {
		return nil, err
	}

	emitEventUpdateReferenceSourceConfig(ctx, req.ReferenceSourceConfig)

	return &types.MsgUpdateReferenceSourceConfigResponse{}, nil
}

// UpdateParams updates the feeds module params.
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

	emitEventUpdateParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
