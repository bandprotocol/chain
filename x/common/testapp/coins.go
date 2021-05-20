package testapp

import sdk "github.com/cosmos/cosmos-sdk/types"

// nolint
var (
	EmptyCoins               = sdk.Coins(nil)
	Coin1minigeo             = sdk.NewInt64Coin(DefaultHelperDenom, 1)
	Coin10loki               = sdk.NewInt64Coin(DefaultBondDenom, 10)
	Coin100000000minigeo     = sdk.NewInt64Coin(DefaultHelperDenom, 100000000)
	Coins1000000loki         = sdk.NewCoins(sdk.NewInt64Coin(DefaultBondDenom, 1000000))
	Coins99999999loki        = sdk.NewCoins(sdk.NewInt64Coin(DefaultBondDenom, 99999999))
	Coin100000000loki        = sdk.NewInt64Coin(DefaultBondDenom, 100000000)
	Coin10000000000loki      = sdk.NewInt64Coin(DefaultBondDenom, 10000000000)
	Coins100000000loki       = sdk.NewCoins(Coin100000000loki)
	Coins10000000000loki     = sdk.NewCoins(Coin10000000000loki)
	DefaultDataProvidersPool = sdk.NewCoins(Coin100000000loki)
	DefaultCommunityPool     = sdk.NewCoins(Coin100000000minigeo, Coin100000000loki)
)
