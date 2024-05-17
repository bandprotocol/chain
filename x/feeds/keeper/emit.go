package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
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

func emitEventUpdateSupportedFeeds(ctx sdk.Context, supportedFeeds types.SupportedFeeds) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateSupportedFeeds,
			sdk.NewAttribute(
				types.AttributeKeyLastUpdateTimestamp,
				fmt.Sprintf("%d", supportedFeeds.LastUpdateTimestamp),
			),
			sdk.NewAttribute(types.AttributeKeyLastUpdateBlock, fmt.Sprintf("%d", supportedFeeds.LastUpdateBlock))),
	)
}

func emitEventSubmitPrice(ctx sdk.Context, valPrice types.ValidatorPrice) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitPrice,
			sdk.NewAttribute(types.AttributeKeyPriceStatus, valPrice.PriceStatus.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valPrice.Validator),
			sdk.NewAttribute(types.AttributeKeySignalID, valPrice.SignalID),
			sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", valPrice.Price)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", valPrice.Timestamp)),
		),
	)
}

func emitEventUpdatePriceService(ctx sdk.Context, priceService types.PriceService) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdatePriceService,
			sdk.NewAttribute(types.AttributeKeyHash, priceService.Hash),
			sdk.NewAttribute(types.AttributeKeyVersion, priceService.Version),
			sdk.NewAttribute(types.AttributeKeyURL, priceService.Url),
		),
	)
}

func emitEventUpdateParams(ctx sdk.Context, params types.Params) {
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyParams, params.String()),
	))
}
