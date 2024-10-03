package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/x/globalfee"
	"github.com/bandprotocol/chain/v3/x/globalfee/keeper"
	"github.com/bandprotocol/chain/v3/x/globalfee/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	globalfeeKeeper keeper.Keeper
	ctx             sdk.Context
	msgServer       types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(globalfee.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	s.globalfeeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.Require().Equal(testCtx.Ctx.Logger().With("module", "x/"+types.ModuleName),
		s.globalfeeKeeper.Logger(testCtx.Ctx))

	err := s.globalfeeKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	s.msgServer = keeper.NewMsgServerImpl(s.globalfeeKeeper)
}

func (s *IntegrationTestSuite) TestParams() {
	testCases := []struct {
		name      string
		input     types.Params
		expectErr string
	}{
		{
			name: "set full valid params",
			input: types.Params{
				MinimumGasPrices: sdk.NewDecCoins(
					sdk.NewDecCoin("ALX", math.NewInt(1)),
					sdk.NewDecCoinFromDec("BLX", math.LegacyNewDecWithPrec(1, 3)),
				),
			},
			expectErr: "",
		},
		{
			name: "set empty coin",
			input: types.Params{
				MinimumGasPrices: sdk.DecCoins(nil),
			},
			expectErr: "",
		},
		{
			name: "set invalid denom",
			input: types.Params{
				MinimumGasPrices: []sdk.DecCoin{
					{
						Denom:  "1AAAA",
						Amount: math.LegacyNewDecFromInt(math.NewInt(1)),
					},
				},
			},
			expectErr: "invalid denom",
		},
		{
			name: "set negative value",
			input: types.Params{
				MinimumGasPrices: []sdk.DecCoin{
					{
						Denom:  "AAAA",
						Amount: math.LegacyNewDecFromInt(math.NewInt(-1)),
					},
				},
			},
			expectErr: "is not positive",
		},
		{
			name: "set duplicated denom",
			input: types.Params{
				MinimumGasPrices: []sdk.DecCoin{
					{
						Denom:  "AAAA",
						Amount: math.LegacyNewDecFromInt(math.NewInt(1)),
					},
					{
						Denom:  "AAAA",
						Amount: math.LegacyNewDecFromInt(math.NewInt(2)),
					},
				},
			},
			expectErr: "duplicate denomination",
		},
		{
			name: "set unsorted denom",
			input: types.Params{
				MinimumGasPrices: []sdk.DecCoin{
					{
						Denom:  "BBBB",
						Amount: math.LegacyNewDecFromInt(math.NewInt(1)),
					},
					{
						Denom:  "AAAA",
						Amount: math.LegacyNewDecFromInt(math.NewInt(2)),
					},
				},
			},
			expectErr: "is not sorted",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expected := s.globalfeeKeeper.GetParams(s.ctx)
			err := s.globalfeeKeeper.SetParams(s.ctx, tc.input)
			if tc.expectErr != "" {
				s.Require().ErrorContains(err, tc.expectErr)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := s.globalfeeKeeper.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}
