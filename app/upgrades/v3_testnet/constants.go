package v3_testnet

import (
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name.
const (
	UpgradeName = "v3"

	Denom = "uband"

	// BlockMaxBytes is the max bytes for a block, 3mb
	BlockMaxBytes = int64(3_000_000)

	// BlockMaxGas is the max gas allowed in a block
	BlockMaxGas = int64(50_000_000)
)

var (
	Upgrade = upgrades.Upgrade{
		UpgradeName:          UpgradeName,
		CreateUpgradeHandler: CreateUpgradeHandler,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{
				feemarkettypes.StoreKey,
			},
			Deleted: []string{
				"globalfee",
			},
		},
	}

	// MinimumGasPrice is the minimum gas price for transactions
	MinimumGasPrice = sdkmath.LegacyNewDecWithPrec(25, 4) // 0.0025uband
)
