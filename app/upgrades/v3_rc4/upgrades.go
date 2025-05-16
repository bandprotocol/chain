package v3_rc4

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v3/app/keepers"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Set param key table for params module migration
		ctx := sdk.UnwrapSDKContext(c)

		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Set minter permission for tunnel module account
		acc := keepers.AccountKeeper.GetModuleAccount(ctx, tunneltypes.ModuleName)
		baseAcc := authtypes.NewBaseAccount(
			acc.GetAddress(),
			acc.GetPubKey(),
			acc.GetAccountNumber(),
			acc.GetSequence(),
		)
		macc := authtypes.NewModuleAccount(baseAcc, tunneltypes.ModuleName, authtypes.Minter)
		keepers.AccountKeeper.SetModuleAccount(ctx, macc)

		return vm, nil
	}
}
