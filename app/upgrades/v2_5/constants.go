package v2_5

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/bandprotocol/chain/v2/app/upgrades"
)

const UpgradeName = "v2_5"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
