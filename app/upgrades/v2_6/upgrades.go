package v2_6

import (
	"github.com/bandprotocol/chain/v2/app/keepers"
	"github.com/bandprotocol/chain/v2/app/upgrades"
	globalfeetypes "github.com/bandprotocol/chain/v2/x/globalfee/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	am upgrades.AppManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			// specifying the whole list instead of adding and removing. Less fragile.
			AllowMessages: ICAAllowMessages,
		}
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		err = keepers.GlobalfeeKeeper.SetParams(ctx, globalfeetypes.Params{
			MinimumGasPrices: sdk.DecCoins{sdk.NewDecCoinFromDec("uband", sdk.NewDecWithPrec(25, 4))},
		})
		if err != nil {
			return nil, err
		}

		return vm, nil
	}
}
