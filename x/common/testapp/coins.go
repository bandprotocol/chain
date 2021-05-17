package testapp

import sdk "github.com/cosmos/cosmos-sdk/types"

// nolint
var (
	EmptyCoins               = sdk.Coins(nil)
	Coin1geo                 = sdk.NewInt64Coin(DefaultHelperDenom, 1)
	Coin10odin               = sdk.NewInt64Coin(DefaultBondDenom, 10)
	Coin100000000geo         = sdk.NewInt64Coin(DefaultHelperDenom, 100000000)
	Coins1000000odin         = sdk.NewCoins(sdk.NewInt64Coin(DefaultBondDenom, 1000000))
	Coins99999999odin        = sdk.NewCoins(sdk.NewInt64Coin(DefaultBondDenom, 99999999))
	Coin100000000odin        = sdk.NewInt64Coin(DefaultBondDenom, 100000000)
	Coin10000000000odin      = sdk.NewInt64Coin(DefaultBondDenom, 10000000000)
	Coins100000000odin       = sdk.NewCoins(Coin100000000odin)
	Coins10000000000odin     = sdk.NewCoins(Coin10000000000odin)
	DefaultDataProvidersPool = sdk.NewCoins(Coin100000000odin)
	DefaultCommunityPool     = sdk.NewCoins(Coin100000000geo, Coin100000000odin)
)
