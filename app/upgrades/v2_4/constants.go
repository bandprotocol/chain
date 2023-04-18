package v2_4

import (
	"github.com/bandprotocol/chain/v2/app/upgrades"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
)

const UpgradeName = "v2_4"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{icahosttypes.StoreKey},
	},
}
