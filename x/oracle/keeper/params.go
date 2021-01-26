package keeper

import (
	"github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
