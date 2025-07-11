package v3_1

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/bandprotocol/chain/v3/app/keepers"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		err = keepers.BandtssKeeper.SetParams(ctx, bandtsstypes.DefaultParams())
		if err != nil {
			return nil, err
		}

		return vm, nil
	}
}
