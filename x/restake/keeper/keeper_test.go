package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/restake/keeper"
	restaketestutil "github.com/bandprotocol/chain/v3/x/restake/testutil"
	"github.com/bandprotocol/chain/v3/x/restake/types"
)

var (
	// staked power
	// - 50 -> address 1
	// - 10 -> address 3
	// - 0  -> others
	ValidAddress1 = sdk.AccAddress("1000000001")
	ValidAddress2 = sdk.AccAddress("1000000002")
	ValidAddress3 = sdk.AccAddress("1000000003")

	ActiveVaultKey   = "active_vault_key"
	InactiveVaultKey = "inactive_vault_key"
	InvalidVaultKey  = "invalid_key"

	LiquidStakerAddress = sdk.AccAddress("12345678901234567890123456789012")

	ValAddress = sdk.ValAddress("4000000001")
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	storeKey      storetypes.StoreKey
	restakeKeeper keeper.Keeper
	accountKeeper *restaketestutil.MockAccountKeeper
	bankKeeper    *restaketestutil.MockBankKeeper
	stakingKeeper *restaketestutil.MockStakingKeeper

	stakingHooks stakingtypes.StakingHooks

	queryClient types.QueryClient
	msgServer   types.MsgServer

	validVaults []types.Vault
	validLocks  []types.Lock
	validParams types.Params
	validStakes []types.Stake
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.validParams = types.NewParams([]string{"uband"})
	suite.validStakes = []types.Stake{
		{
			StakerAddress: ValidAddress1.String(),
			Coins:         sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(50))),
		},
		{
			StakerAddress: ValidAddress3.String(),
			Coins:         sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(10))),
		},
	}
	suite.validVaults = []types.Vault{
		{
			Key:      ActiveVaultKey,
			IsActive: true,
		},
		{
			Key:      InactiveVaultKey,
			IsActive: false,
		},
	}

	suite.validLocks = []types.Lock{
		{
			StakerAddress: ValidAddress1.String(),
			Key:           ActiveVaultKey,
			Power:         sdkmath.NewInt(100),
		},
		{
			StakerAddress: ValidAddress1.String(),
			Key:           InactiveVaultKey,
			Power:         sdkmath.NewInt(50),
		},
		{
			StakerAddress: ValidAddress2.String(),
			Key:           ActiveVaultKey,
			Power:         sdkmath.NewInt(10),
		},
	}

	key := storetypes.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()
	moduleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := restaketestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().
		GetModuleAddress(types.ModuleName).
		Return(moduleAccount.GetAddress()).
		AnyTimes()
	accountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(moduleAccount).
		AnyTimes()
	accountKeeper.EXPECT().
		SetModuleAccount(gomock.Any(), moduleAccount).
		Return().
		AnyTimes()
	suite.accountKeeper = accountKeeper

	bankKeeper := restaketestutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().
		GetAllBalances(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	bankKeeper.EXPECT().
		GetAllBalances(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	suite.bankKeeper = bankKeeper

	stakingKeeper := restaketestutil.NewMockStakingKeeper(ctrl)
	suite.stakingKeeper = stakingKeeper

	suite.restakeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	suite.restakeKeeper.InitGenesis(suite.ctx, types.DefaultGenesisState())

	suite.stakingHooks = suite.restakeKeeper.Hooks()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, encCfg.InterfaceRegistry)
	queryServer := keeper.Querier{
		Keeper: &suite.restakeKeeper,
	}

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, queryServer)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient
	suite.msgServer = keeper.NewMsgServerImpl(&suite.restakeKeeper)
}

func (suite *KeeperTestSuite) setupState() {
	err := suite.restakeKeeper.SetParams(suite.ctx, suite.validParams)
	suite.Require().NoError(err)

	for _, vault := range suite.validVaults {
		suite.restakeKeeper.SetVault(suite.ctx, vault)
	}

	for _, lock := range suite.validLocks {
		suite.restakeKeeper.SetLock(suite.ctx, lock)
	}

	for _, stake := range suite.validStakes {
		suite.restakeKeeper.SetStake(suite.ctx, stake)
	}
}

func (suite *KeeperTestSuite) TestGetTotalPower() {
	ctx := suite.ctx
	suite.setupState()

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(1e18), nil).
		Times(1)
	expPower, _ := sdkmath.NewIntFromString("1000000000000000050")
	power, err := suite.restakeKeeper.GetTotalPower(ctx, ValidAddress1)
	suite.Require().Equal(expPower, power)
	suite.Require().NoError(err)

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress2).
		Return(sdkmath.NewInt(1e18), nil).
		Times(1)
	expPower, _ = sdkmath.NewIntFromString("1000000000000000000")
	power, err = suite.restakeKeeper.GetTotalPower(ctx, ValidAddress2)
	suite.Require().Equal(expPower, power)
	suite.Require().NoError(err)

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress3).
		Return(sdkmath.NewInt(10), nil).
		Times(1)
	expPower, _ = sdkmath.NewIntFromString("20")
	power, err = suite.restakeKeeper.GetTotalPower(ctx, ValidAddress3)
	suite.Require().Equal(expPower, power)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetDelegationPower() {
	ctx := suite.ctx
	suite.setupState()

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(1e18), nil).
		Times(1)
	expPower, _ := sdkmath.NewIntFromString("1000000000000000000")
	power, err := suite.restakeKeeper.GetDelegationPower(ctx, ValidAddress1)
	suite.Require().Equal(expPower, power)
	suite.Require().NoError(err)

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress2).
		Return(sdkmath.NewInt(1e18), nil).
		Times(1)
	expPower, _ = sdkmath.NewIntFromString("1000000000000000000")
	power, err = suite.restakeKeeper.GetDelegationPower(ctx, ValidAddress2)
	suite.Require().Equal(expPower, power)
	suite.Require().NoError(err)

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress3).
		Return(sdkmath.NewInt(10), nil).
		Times(1)
	expPower, _ = sdkmath.NewIntFromString("10")
	power, err = suite.restakeKeeper.GetDelegationPower(ctx, ValidAddress3)
	suite.Require().Equal(expPower, power)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestIsLiquidStaker() {
	// not liquid staker
	isLiquidStaker := suite.restakeKeeper.IsLiquidStaker(ValidAddress1)
	suite.Require().Equal(false, isLiquidStaker)

	// is liquid staker
	isLiquidStaker = suite.restakeKeeper.IsLiquidStaker(LiquidStakerAddress)
	suite.Require().Equal(true, isLiquidStaker)
}
