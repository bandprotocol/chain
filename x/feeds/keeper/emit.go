package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func emitEventUpdateSignalTotalPower(ctx sdk.Context, signal types.Signal) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateSignalTotalPower,
			sdk.NewAttribute(types.AttributeKeySignalID, signal.ID),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", signal.Power)),
		),
	)
}

func emitEventDeleteSignalTotalPower(ctx sdk.Context, signal types.Signal) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeleteSignalTotalPower,
			sdk.NewAttribute(types.AttributeKeySignalID, signal.ID),
		),
	)
}

func emitEventUpdateCurrentFeeds(ctx sdk.Context, currentFeeds types.CurrentFeeds) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateCurrentFeeds,
			sdk.NewAttribute(
				types.AttributeKeyLastUpdateTimestamp,
				fmt.Sprintf("%d", currentFeeds.LastUpdateTimestamp),
			),
			sdk.NewAttribute(types.AttributeKeyLastUpdateBlock, fmt.Sprintf("%d", currentFeeds.LastUpdateBlock))),
	)
}

func emitEventSubmitSignalPrice(ctx sdk.Context, valPrice types.ValidatorPrice) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitSignalPrice,
			sdk.NewAttribute(types.AttributeKeySignalPriceStatus, valPrice.SignalPriceStatus.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valPrice.Validator),
			sdk.NewAttribute(types.AttributeKeySignalID, valPrice.SignalID),
			sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", valPrice.Price)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", valPrice.Timestamp)),
		),
	)
}

func emitEventUpdateReferenceSourceConfig(ctx sdk.Context, referenceSourceConfig types.ReferenceSourceConfig) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateReferenceSourceConfig,
			sdk.NewAttribute(types.AttributeKeyRegistryIPFSHash, referenceSourceConfig.RegistryIPFSHash),
			sdk.NewAttribute(types.AttributeKeyRegistryVersion, referenceSourceConfig.RegistryVersion),
		),
	)
}

func emitEventUpdateParams(ctx sdk.Context, params types.Params) {
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, params.String()),
	))
}

func emitEventUpdatePrice(ctx sdk.Context, price types.Price) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdatePrice,
			sdk.NewAttribute(types.AttributeKeySignalID, price.SignalID),
			sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", price.Price)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", price.Timestamp)),
		),
	)
}
