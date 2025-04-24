package v3_rc3

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name.
const UpgradeName = "v3.0.0-rc3"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
