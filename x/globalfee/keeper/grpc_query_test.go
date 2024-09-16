package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/globalfee/keeper"
	"github.com/bandprotocol/chain/v3/x/globalfee/types"
)

func TestQueryParams(t *testing.T) {
	specs := map[string]struct {
		setupStore func(ctx sdk.Context, k keeper.Keeper)
		expMin     sdk.DecCoins
	}{
		"one coin": {
			setupStore: func(ctx sdk.Context, k keeper.Keeper) {
				_ = k.SetParams(ctx, types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", math.OneInt())),
				})
			},
			expMin: sdk.NewDecCoins(sdk.NewDecCoin("ALX", math.OneInt())),
		},
		"multiple coins": {
			setupStore: func(ctx sdk.Context, k keeper.Keeper) {
				_ = k.SetParams(ctx, types.Params{
					MinimumGasPrices: sdk.NewDecCoins(
						sdk.NewDecCoin("ALX", math.OneInt()),
						sdk.NewDecCoin("BLX", math.NewInt(2)),
					),
				})
			},
			expMin: sdk.NewDecCoins(sdk.NewDecCoin("ALX", math.OneInt()), sdk.NewDecCoin("BLX", math.NewInt(2))),
		},
		"no min gas price set": {
			setupStore: func(ctx sdk.Context, k keeper.Keeper) {
				_ = k.SetParams(ctx, types.Params{})
			},
		},
		"no param set": {
			setupStore: func(ctx sdk.Context, k keeper.Keeper) {
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			encCfg := moduletestutil.MakeTestEncodingConfig()
			key := storetypes.NewKVStoreKey(types.StoreKey)
			ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx

			k := keeper.NewKeeper(
				encCfg.Codec,
				key,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			)

			q := keeper.Querier{Keeper: k}
			spec.setupStore(ctx, k)
			gotResp, gotErr := q.Params(ctx, nil)

			require.NoError(t, gotErr)
			require.NotNil(t, gotResp)
			assert.Equal(t, spec.expMin, gotResp.Params.MinimumGasPrices)
		})
	}
}
