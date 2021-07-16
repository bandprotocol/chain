package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	auctionkeeper "github.com/GeoDB-Limited/odin-core/x/auction/keeper"
)

// handleBeginBlock re-calculates and saves the rolling seed value based on block hashes.
func handleBeginBlock(ctx sdk.Context, k auctionkeeper.Keeper) {
	auctionStatus := k.GetAuctionStatus(ctx)
	accumulatedPaymentsForData := k.GetAccumulatedPaymentsForData(ctx)
	auctionThreshold := k.GetAuctionStartThreshold(ctx)

	if accumulatedPaymentsForData.IsAllGTE(auctionThreshold) {
		if !auctionStatus.Pending {
			if err := k.StartAuction(ctx); err != nil {
				k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to start auction").Error())
			}
		}
		return
	}

	if auctionStatus.Pending {
		if err := k.FinishAuction(ctx); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to finish auction").Error())
		}
	}
}
