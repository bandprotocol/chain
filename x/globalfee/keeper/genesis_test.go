package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/x/globalfee"
	"github.com/bandprotocol/chain/v2/x/globalfee/keeper"
	"github.com/bandprotocol/chain/v2/x/globalfee/types"
)

type GenesisTestSuite struct {
	suite.Suite

	sdkCtx sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
	key    *storetypes.KVStoreKey
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(globalfee.AppModuleBasic{})

	// gomock initializations
	s.cdc = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	s.sdkCtx = testCtx.Ctx
	s.key = key

	s.keeper = keeper.NewKeeper(s.cdc, key, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func (s *GenesisTestSuite) TestImportExportGenesis() {
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
			exp: types.GenesisState{Params: types.Params{MinimumGasPrices: nil}},
		},
	}
	for name, spec := range specs {
		s.Run(name, func() {
			genesisState := spec.exp
			s.keeper.InitGenesis(s.sdkCtx, &genesisState)

			params := s.keeper.GetParams(s.sdkCtx)
			s.Require().Equal(genesisState.Params, params)

			genesisState2 := s.keeper.ExportGenesis(s.sdkCtx)
			s.Require().Equal(&genesisState, genesisState2)
		})
	}
}
