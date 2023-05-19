package v2_6

import (
	"github.com/bandprotocol/chain/v2/app/keepers"
	"github.com/bandprotocol/chain/v2/app/upgrades"
	"github.com/bandprotocol/chain/v2/x/globalfee"
	globalfeetypes "github.com/bandprotocol/chain/v2/x/globalfee/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"

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
			// Specifying the whole list instead of adding and removing. Less fragile.
			AllowMessages: ICAAllowMessages,
		}
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		minGasPriceGenesisState := &globalfeetypes.GenesisState{
			Params: globalfeetypes.Params{
				MinimumGasPrices: sdk.DecCoins{sdk.NewDecCoinFromDec("uband", sdk.NewDecWithPrec(25, 4))},
			},
		}
		am.GetSubspace(globalfee.ModuleName).SetParamSet(ctx, &minGasPriceGenesisState.Params)

		// set version of globalfee so that it won't run initgenesis again
		fromVM["globalfee"] = 1

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
