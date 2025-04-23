package v3

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/bandprotocol/chain/v3/app/keepers"
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

		oracleParams := keepers.OracleKeeper.GetParams(ctx)
		oracleParams.MaxCalldataSize = 512
		oracleParams.MaxReportDataSize = 512
		err = keepers.OracleKeeper.SetParams(ctx, oracleParams)
		if err != nil {
			return nil, err
		}

		return vm, nil
	}
}
