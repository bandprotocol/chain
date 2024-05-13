package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func emitEventUpdateFeed(ctx sdk.Context, feed types.Feed) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateFeed,
			sdk.NewAttribute(types.AttributeKeySignalID, feed.SignalID),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", feed.Power)),
			sdk.NewAttribute(types.AttributeKeyInterval, fmt.Sprintf("%d", feed.Interval)),
			sdk.NewAttribute(
				types.AttributeKeyLastIntervalUpdateTimestamp,
				fmt.Sprintf("%d", feed.LastIntervalUpdateTimestamp),
			),
			sdk.NewAttribute(
				types.AttributeKeyDeviationInThousandth,
				fmt.Sprintf("%d", feed.DeviationInThousandth),
			),
		),
	)
}

func emitEventDeleteFeed(ctx sdk.Context, feed types.Feed) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeleteFeed,
			sdk.NewAttribute(types.AttributeKeySignalID, feed.SignalID),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", feed.Power)),
			sdk.NewAttribute(types.AttributeKeyInterval, fmt.Sprintf("%d", feed.Interval)),
			sdk.NewAttribute(
				types.AttributeKeyLastIntervalUpdateTimestamp,
				fmt.Sprintf("%d", feed.LastIntervalUpdateTimestamp),
			),
			sdk.NewAttribute(
				types.AttributeKeyDeviationInThousandth,
				fmt.Sprintf("%d", feed.DeviationInThousandth),
			),
		),
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
