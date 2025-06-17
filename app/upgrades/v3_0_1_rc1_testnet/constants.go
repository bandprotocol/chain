package v3_0_1_rc1_testnet

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name.
const UpgradeName = "v3_0_1_rc1_testnet"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
