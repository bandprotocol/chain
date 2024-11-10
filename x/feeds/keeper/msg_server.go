package keeper

import (
	"context"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
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

// Vote votes signals.
func (ms msgServer) Vote(
	goCtx context.Context,
	msg *types.MsgVote,
) (*types.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// convert the voter address from Bech32 format to sdk.AccAddress
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, err
	}

	// check if the number of submitted signals exceeds the maximum allowed feeds
	if len(msg.Signals) > int(ms.GetParams(ctx).MaxCurrentFeeds) {
		return nil, types.ErrSubmittedSignalsTooLarge.Wrapf(
			"maximum number of signals is %d but received %d",
			ms.GetParams(ctx).MaxCurrentFeeds, len(msg.Signals),
		)
	}

	// lock the voter's power equal to the sum of the signal powers
	err = ms.LockVoterPower(ctx, voter, msg.Signals)
	if err != nil {
		return nil, err
	}

	// RegisterNewSignals deletes previous signals and registers new signals then returns feed power differences
	signalIDToPowerDiff := ms.UpdateVoteAndReturnPowerDiff(ctx, voter, msg.Signals)

	// sort keys to guarantee order of signalIDToPowerDiff iteration
	keys := make([]string, 0, len(signalIDToPowerDiff))
	for k := range signalIDToPowerDiff {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// iterate over each signal ID, updating the total power and ensuring no negative power values
	for _, signalID := range keys {
		// get the power difference for the current signal ID
		powerDiff := signalIDToPowerDiff[signalID]

		// retrieve the total power of the current signal ID from the store
		signalTotalPower, err := ms.GetSignalTotalPower(ctx, signalID)
		if err != nil {
			// initialize a new signal with zero power if the signal ID does not exist
			signalTotalPower = types.NewSignal(
				signalID,
				0,
			)
		}

		// update the total power of the signal by adding the power difference
		signalTotalPower.Power += powerDiff

		// ensure the total power of the signal is not negative
		if signalTotalPower.Power < 0 {
			return nil, types.ErrPowerNegative
		}

		// save the updated signal total power back to the store
		ms.SetSignalTotalPower(ctx, signalTotalPower)
	}

	// return an empty response indicating success
	return &types.MsgVoteResponse{}, nil
}

// SubmitSignalPrices submits new validator prices.
func (ms msgServer) SubmitSignalPrices(
	goCtx context.Context,
	msg *types.MsgSubmitSignalPrices,
) (*types.MsgSubmitSignalPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime().Unix()
	blockHeight := ctx.BlockHeight()

	// check if the number of signal prices exceeds the maximum allowed feeds
	if len(msg.SignalPrices) > int(ms.Keeper.GetParams(ctx).MaxCurrentFeeds) {
		return nil, types.ErrSignalPricesTooLarge
	}

	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	// check if the validator is required to send prices
	if err := ms.ValidateValidatorRequiredToSend(ctx, val); err != nil {
		return nil, err
	}

	// check if the timestamp is not too far from the block time
	if types.AbsInt64(msg.Timestamp-blockTime) > ms.GetParams(ctx).AllowableBlockTimeDiscrepancy {
		return nil, types.ErrInvalidTimestamp.Wrapf(
			"block_time: %d, timestamp: %d",
			blockTime,
			msg.Timestamp,
		)
	}

	// get current feeds
	currentFeeds := ms.GetCurrentFeeds(ctx)
	currentFeedsMap := make(map[string]int)
	for idx, feed := range currentFeeds.Feeds {
		currentFeedsMap[feed.SignalID] = idx
	}

	valPrices := make([]types.ValidatorPrice, len(currentFeedsMap))
	prevValPrices, err := ms.GetValidatorPriceList(ctx, val)
	if err == nil {
		for _, valPrice := range prevValPrices.ValidatorPrices {
			idx, ok := currentFeedsMap[valPrice.SignalID]
			if ok {
				valPrices[idx] = valPrice
			}
		}
	}

	cooldownTime := ms.GetParams(ctx).CooldownTime
	for _, signalPrice := range msg.SignalPrices {
		idx, ok := currentFeedsMap[signalPrice.SignalID]
		if !ok {
			return nil, types.ErrSignalIDNotSupported.Wrapf(
				"signal_id: %s",
				signalPrice.SignalID,
			)
		}

		// check if price is not too fast
		valPrice := valPrices[idx]
		if valPrice.SignalPriceStatus != types.SignalPriceStatusUnspecified &&
			blockTime < valPrice.Timestamp+cooldownTime {
			return nil, types.ErrPriceSubmitTooEarly
		}

		valPrice = types.NewValidatorPrice(signalPrice, blockTime, blockHeight)
		valPrices[idx] = valPrice
		emitEventSubmitSignalPrice(ctx, val, valPrice)
	}

	if err := ms.SetValidatorPriceList(ctx, val, valPrices); err != nil {
		return nil, err
	}

	return &types.MsgSubmitSignalPricesResponse{}, nil
}

// UpdateReferenceSourceConfig updates reference source configuration.
func (ms msgServer) UpdateReferenceSourceConfig(
	goCtx context.Context,
	msg *types.MsgUpdateReferenceSourceConfig,
) (*types.MsgUpdateReferenceSourceConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the request is from the admin
	admin := ms.GetParams(ctx).Admin
	if admin != msg.Admin {
		return nil, types.ErrInvalidSigner.Wrapf(
			"invalid admin; expected %s, got %s",
			admin,
			msg.Admin,
		)
	}

	// update the reference source configuration
	if err := ms.SetReferenceSourceConfig(ctx, msg.ReferenceSourceConfig); err != nil {
		return nil, err
	}

	emitEventUpdateReferenceSourceConfig(ctx, msg.ReferenceSourceConfig)

	return &types.MsgUpdateReferenceSourceConfigResponse{}, nil
}

// UpdateParams updates the feeds module params.
func (ms msgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the request is from the authority
	if ms.GetAuthority() != msg.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			ms.GetAuthority(),
			msg.Authority,
		)
	}

	// update the parameters
	if err := ms.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	emitEventUpdateParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
