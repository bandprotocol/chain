package v3_rc4

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name.
const UpgradeName = "v3_rc4"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
