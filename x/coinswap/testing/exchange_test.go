package testing

import (
	swaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	"github.com/GeoDB-Limited/odin-core/x/common/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	geo         = "geo"
	odin        = "odin"
	initialRate = 10
)

func TestKeeper_ExchangeDenom(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(false, true)

	app.CoinswapKeeper.SetInitialRate(ctx, sdk.NewDec(initialRate))
	app.CoinswapKeeper.SetParams(ctx, swaptypes.DefaultParams())

	err := app.CoinswapKeeper.ExchangeDenom(ctx, geo, odin, sdk.NewInt64Coin(geo, 10), testapp.Alice.Address)

	assert.NoError(t, err, "exchange denom failed")
}
