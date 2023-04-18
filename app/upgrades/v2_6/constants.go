package v2_6

import (
	"github.com/bandprotocol/chain/v2/app/upgrades"
	"github.com/bandprotocol/chain/v2/x/globalfee"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

const UpgradeName = "v2_6"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{group.StoreKey, globalfee.ModuleName},
	},
}
