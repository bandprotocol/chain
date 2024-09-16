package v2_4

import (
	storetypes "cosmossdk.io/store/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
)

const UpgradeName = "v2_4"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{icahosttypes.StoreKey},
	},
}
