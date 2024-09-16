package v2_5

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
)

const UpgradeName = "v2_5"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
