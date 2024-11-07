package keeper

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

// AllocateTokens allocates a portion of fee collected in the previous blocks to members in the
// current group. Note that this reward is also subjected to comm tax and this reward is
// calculate after allocation in oracle module.
func (k Keeper) AllocateTokens(ctx sdk.Context) {
	gid := k.GetCurrentGroup(ctx).GroupID
	if gid == tss.GroupID(0) {
		return
	}

	// Get all active members in the current group.
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
	tssRewardRatio := math.LegacyNewDecWithPrec(int64(k.GetParams(ctx).RewardPercentage), 2)
	tssRewardInt, _ := totalFee.MulDecTruncate(tssRewardRatio).TruncateDecimal()

	// Transfer the reward from fee collector to distr module.
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, distrtypes.ModuleName, tssRewardInt)
	if err != nil {
		panic(err)
	}

	// Convert the transferred tokens back to DecCoins for internal distr allocations.
	tssReward := sdk.NewDecCoinsFromCoins(tssRewardInt...)
	communityTax, err := k.distrKeeper.GetCommunityTax(ctx)
	if err != nil {
		panic(err)
	}
	rewardMultiplier := math.LegacyOneDec().Sub(communityTax)

	// calculate the reward for each active member.
	n := math.LegacyNewDec(int64(len(validMembers)))
	powerFraction := math.LegacyNewDec(1).QuoTruncate(n)
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
	communityFund := tssRewardInt.Sub(rewardInt.MulInt(math.NewInt(int64(len(validMembers))))...)
	err = k.distrKeeper.FundCommunityPool(
		ctx,
		communityFund,
		k.authKeeper.GetModuleAccount(ctx, distrtypes.ModuleName).GetAddress(),
	)
	if err != nil {
		panic(err)
	}
}
