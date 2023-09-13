package globalfee

import (
	"testing"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/globalfee/keeper"
	"github.com/bandprotocol/chain/v2/x/globalfee/types"
)

func TestDefaultGenesis(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	gotJSON := AppModuleBasic{}.DefaultGenesis(encCfg.Codec)
	assert.JSONEq(t, `{"params":{"minimum_gas_prices":[]}}`, string(gotJSON))
}

func TestValidateGenesis(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	specs := map[string]struct {
		src    string
		expErr bool
	}{
		"all good": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}]}}`,
		},
		"empty minimum": {
			src: `{"params":{"minimum_gas_prices":[]}}`,
		},
		"minimum not set": {
			src: `{"params":{}}`,
		},
		"zero amount not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"0"}]}}`,
			expErr: true,
		},
		"duplicate denoms not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"negative amounts not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"-1"}]}}`,
			expErr: true,
		},
		"denom must be sorted": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ZLX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"sorted denoms is allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ZLX", "amount":"2"}]}}`,
			expErr: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := AppModuleBasic{}.ValidateGenesis(encCfg.Codec, nil, []byte(spec.src))
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestInitExportGenesis(t *testing.T) {
	specs := map[string]struct {
		src string
		exp types.GenesisState
	}{
		"single fee": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}]}}`,
			exp: types.GenesisState{
				Params: types.Params{MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)))},
			},
		},
		"multiple fee options": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}, {"denom":"BLX", "amount":"0.001"}]}}`,
			exp: types.GenesisState{
				Params: types.Params{MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)),
					sdk.NewDecCoinFromDec("BLX", sdk.NewDecWithPrec(1, 3)))},
			},
		},
		"no fee set": {
			src: `{"params":{}}`,
			exp: types.GenesisState{Params: types.Params{MinimumGasPrices: sdk.DecCoins{}}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, encCfg, keeper := setupTestStore(t)
			m := NewAppModule(keeper)
			m.InitGenesis(ctx, encCfg.Codec, []byte(spec.src))
			gotJSON := m.ExportGenesis(ctx, encCfg.Codec)
			var got types.GenesisState
			require.NoError(t, encCfg.Codec.UnmarshalJSON(gotJSON, &got))
			assert.Equal(t, spec.exp, got, string(gotJSON))
		})
	}
}

func setupTestStore(t *testing.T) (sdk.Context, moduletestutil.TestEncodingConfig, keeper.Keeper) {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	ms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	globalfeeKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeKey,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	return ctx, encCfg, globalfeeKeeper
}
