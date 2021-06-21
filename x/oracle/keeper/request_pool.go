package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func (k Keeper) DepositRequestPool(ctx sdk.Context, requestKey string, portID string, channelID string, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoins(ctx, sender, types.GetEscrowAddress(requestKey, portID, channelID), amount)
}

func (k Keeper) GetRequestPoolBalances(ctx sdk.Context, requestKey string, portID string, channelID string) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, types.GetEscrowAddress(requestKey, portID, channelID))
}
