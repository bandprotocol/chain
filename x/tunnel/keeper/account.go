package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// TODO: patch account generation follow ibc-v8
// https://github.com/cosmos/ibc-go/blob/v8.5.1/modules/apps/27-interchain-accounts/types/account.go#L46
func (k Keeper) GenerateAccount(ctx sdk.Context, key string) (sdk.AccAddress, error) {
	header := ctx.BlockHeader()

	buf := []byte(key)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	moduleCred, err := authtypes.NewModuleCredential(types.ModuleName, []byte(types.TunnelAccountsKey), buf)
	if err != nil {
		return nil, err
	}

	tunnelAccAddr := sdk.AccAddress(moduleCred.Address())

	if acc := k.authKeeper.GetAccount(ctx, tunnelAccAddr); acc != nil {
		// this should not happen
		return nil, types.ErrAccountAlreadyExist.Wrapf(
			"existing account for newly generated key account address %s",
			tunnelAccAddr.String(),
		)
	}

	tunnelAcc, err := authtypes.NewBaseAccountWithPubKey(moduleCred)
	if err != nil {
		return nil, err
	}

	k.authKeeper.NewAccount(ctx, tunnelAcc)
	k.authKeeper.SetAccount(ctx, tunnelAcc)

	return tunnelAccAddr, nil
}
