package v2_5

import (
	"github.com/bandprotocol/chain/v2/app/upgrades"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

const UpgradeName = "v2_5"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
