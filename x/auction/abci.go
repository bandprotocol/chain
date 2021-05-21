package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	auctionkeeper "github.com/GeoDB-Limited/odin-core/x/auction/keeper"
)

// handleBeginBlock re-calculates and saves the rolling seed value based on block hashes.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k auctionkeeper.Keeper) {
	accumulatedPaymentsForData := k.GetAccumulatedPaymentsForData(ctx)
	auctionThreshold := k.GetAuctionStartThreshold(ctx)
	if accumulatedPaymentsForData.IsAllGTE(auctionThreshold) {
		if err := k.StartAuction(ctx); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to start auction").Error())
			return
		}
		accumulatedPaymentsForData = accumulatedPaymentsForData.Sub(auctionThreshold)
		k.SetAccumulatedPaymentsForData(ctx, accumulatedPaymentsForData)
	}
}

// handleEndBlock cleans up the state during end block. See comment in the implementation!
func handleEndBlock(ctx sdk.Context, k auctionkeeper.Keeper) {
	auctionStatus := k.GetAuctionStatus(ctx)
	currentBlockHeight := uint64(ctx.BlockHeight())
	if currentBlockHeight == auctionStatus.FinishBlock {
		if err := k.FinishAuction(ctx); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "failed to finish auction").Error())
		}
	}

}
