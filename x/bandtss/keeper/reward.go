package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// AllocateTokens allocates a portion of fee collected in the previous blocks to members in the
// current group. Note that this reward is also subjected to comm tax and this reward is
// calculate after allocation in oracle module.
func (k Keeper) AllocateTokens(ctx sdk.Context) {
	gid := k.GetCurrentGroupID(ctx)
	if gid == tss.GroupID(0) {
		return
	}

	members := k.tssKeeper.MustGetMembers(ctx, gid)
	validMembers := make([]sdk.AccAddress, 0, len(members))
	for _, m := range members {
		acc := sdk.MustAccAddressFromBech32(m.Address)

		deQueue := k.tssKeeper.GetDEQueue(ctx, acc)
		if deQueue.Tail > deQueue.Head && m.IsActive {
			validMembers = append(validMembers, acc)
		}
	}

	// No active members performing tss tasks, nothing needs to be done here.
	if len(validMembers) == 0 {
		return
	}

	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	totalFee := sdk.NewDecCoinsFromCoins(k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())...)

	// Compute the fee allocated for tss module.
	tssRewardRatio := sdk.NewDecWithPrec(int64(k.GetParams(ctx).RewardPercentage), 2)
	tssRewardInt, _ := totalFee.MulDecTruncate(tssRewardRatio).TruncateDecimal()

	// Transfer the reward from fee collector to distr module.
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, distrtypes.ModuleName, tssRewardInt)
	if err != nil {
		panic(err)
	}

	// Convert the transferred tokens back to DecCoins for internal distr allocations.
	tssReward := sdk.NewDecCoinsFromCoins(tssRewardInt...)
	rewardMultiplier := sdk.OneDec().Sub(k.distrKeeper.GetCommunityTax(ctx))

	// calculate the reward for each active member.
	n := sdk.NewDec(int64(len(validMembers)))
	powerFraction := sdk.NewDec(1).QuoTruncate(n)
	reward := tssReward.MulDecTruncate(rewardMultiplier).MulDecTruncate(powerFraction)
	rewardInt, _ := reward.TruncateDecimal()

	// Allocate non-community pool tokens to active members.
	for _, acc := range validMembers {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, acc, rewardInt)
		if err != nil {
			panic(err)
		}
	}

	// Allocate the remaining coins to the community pool.
	remaining := tssReward.Sub(sdk.NewDecCoinsFromCoins(rewardInt...).MulDecTruncate(n))
	feePool := k.distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	k.distrKeeper.SetFeePool(ctx, feePool)
}
