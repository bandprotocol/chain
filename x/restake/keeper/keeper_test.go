package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/restake/keeper"
	restaketestutil "github.com/bandprotocol/chain/v2/x/restake/testutil"
	"github.com/bandprotocol/chain/v2/x/restake/types"
)

var (
	// delegate power
	// - 1e18 -> address 1,2
	// - 10   -> address 3
	ValidAddress1 = sdk.AccAddress("1000000001")
	ValidAddress2 = sdk.AccAddress("1000000002")
	ValidAddress3 = sdk.AccAddress("1000000003")

	// status
	// - active   -> key 1,2,4
	// - inactive -> key 3
	// key 4 has total power as zero
	ValidKey1  = "key1"
	ValidKey2  = "key2"
	ValidKey3  = "key3"
	ValidKey4  = "key4"
	InvalidKey = "nonKey"

	ValidPoolAddress1 = sdk.AccAddress("2000000001")
	ValidPoolAddress2 = sdk.AccAddress("2000000002")
	ValidPoolAddress3 = sdk.AccAddress("2000000003")
	ValidPoolAddress4 = sdk.AccAddress("2000000004")

	RewarderAddress = sdk.AccAddress("3000000001")
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	storeKey      storetypes.StoreKey
	restakeKeeper keeper.Keeper
	accountKeeper *restaketestutil.MockAccountKeeper
	bankKeeper    *restaketestutil.MockBankKeeper
	stakingKeeper *restaketestutil.MockStakingKeeper

	queryClient types.QueryClient
	msgServer   types.MsgServer

	validKeys  []types.Key
	validLocks []types.Lock
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.validKeys = []types.Key{
		{
			Name:            ValidKey1,
			PoolAddress:     ValidPoolAddress1.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(20),
			RewardPerPowers: sdk.NewDecCoins(sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(1, 1))),
			Remainders:      nil,
		},
		{
			Name:            ValidKey2,
			PoolAddress:     ValidPoolAddress2.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            ValidKey3,
			PoolAddress:     ValidPoolAddress3.String(),
			IsActive:        false,
			TotalPower:      sdkmath.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            ValidKey4,
			PoolAddress:     ValidPoolAddress4.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(0),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
	}

	suite.validLocks = []types.Lock{
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            ValidKey1,
			Amount:         sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            ValidKey2,
			Amount:         sdkmath.NewInt(100),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            ValidKey3,
			Amount:         sdkmath.NewInt(50),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress2.String(),
			Key:            ValidKey1,
			Amount:         sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
	}

	suite.resetState()
}

func (suite *KeeperTestSuite) resetState() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := restaketestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().
		GetAccount(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	accountKeeper.EXPECT().
		NewAccount(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	accountKeeper.EXPECT().
		SetAccount(gomock.Any(), gomock.Any()).
		Return().
		AnyTimes()
	suite.accountKeeper = accountKeeper

	bankKeeper := restaketestutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().
		SendCoins(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	suite.bankKeeper = bankKeeper

	stakingKeeper := restaketestutil.NewMockStakingKeeper(ctrl)
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(1e18)).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress2).
		Return(sdkmath.NewInt(1e18)).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress3).
		Return(sdkmath.NewInt(10)).
		AnyTimes()
	suite.stakingKeeper = stakingKeeper

	suite.restakeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
	)
	suite.restakeKeeper.InitGenesis(suite.ctx, types.DefaultGenesisState())

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
	for _, key := range suite.validKeys {
		suite.restakeKeeper.SetKey(suite.ctx, key)
	}

	for _, lock := range suite.validLocks {
		suite.restakeKeeper.SetLock(suite.ctx, lock)
	}
}

func (suite *KeeperTestSuite) TestScenarios() {
	ctx := suite.ctx
	suite.setupState()

	testCases := []struct {
		name  string
		check func()
	}{
		{
			name: "1 locker",
			check: func() {
				// pre check
				_, err := suite.restakeKeeper.GetKey(ctx, ValidKey1)
				suite.Require().Error(err)

				_, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey1)
				suite.Require().Error(err)

				// address1 locks on key1 1000 powers
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey1, sdkmath.NewInt(1000))
				suite.Require().NoError(err)

				// post check
				power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(1000), power)

				key, err := suite.restakeKeeper.GetKey(ctx, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:            ValidKey1,
					PoolAddress:     "cosmos15p5q4vgsn4lxk7zefu86fep6htydvr3pfhkv5h0lt49tw8pjsekqtp8tpj",
					IsActive:        true,
					RewardPerPowers: nil,
					TotalPower:      sdk.NewInt(1000),
					Remainders:      nil,
				}, key)

				lock, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					LockerAddress:  ValidAddress1.String(),
					Key:            ValidKey1,
					Amount:         sdk.NewInt(1000),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// rewards in 1 aaaa, 1000 bbbb
				err = suite.restakeKeeper.AddRewards(ctx, RewarderAddress, ValidKey1, sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1000)),
				))
				suite.Require().NoError(err)

				// post check
				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey1,
					PoolAddress: "cosmos15p5q4vgsn4lxk7zefu86fep6htydvr3pfhkv5h0lt49tw8pjsekqtp8tpj",
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 0)),
					),
					TotalPower: sdk.NewInt(1000),
					Remainders: nil,
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					LockerAddress:  ValidAddress1.String(),
					Key:            ValidKey1,
					Amount:         sdk.NewInt(1000),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// address1 locks on key1 100 powers (override)
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey1, sdkmath.NewInt(100))
				suite.Require().NoError(err)

				// post check
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(100), power)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey1,
					PoolAddress: "cosmos15p5q4vgsn4lxk7zefu86fep6htydvr3pfhkv5h0lt49tw8pjsekqtp8tpj",
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 0)),
					),
					TotalPower: sdk.NewInt(100),
					Remainders: nil,
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					LockerAddress:  ValidAddress1.String(),
					Key:            ValidKey1,
					Amount:         sdk.NewInt(100),
					PosRewardDebts: nil,
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(9, 1)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(900, 0)),
					),
				}, lock)

				// address1 locks on key1 2000 powers (override)
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey1, sdkmath.NewInt(2000))
				suite.Require().NoError(err)

				// post check
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(2000), power)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey1,
					PoolAddress: "cosmos15p5q4vgsn4lxk7zefu86fep6htydvr3pfhkv5h0lt49tw8pjsekqtp8tpj",
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 0)),
					),
					TotalPower: sdk.NewInt(2000),
					Remainders: nil,
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey1)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					LockerAddress: ValidAddress1.String(),
					Key:           ValidKey1,
					Amount:        sdk.NewInt(2000),
					PosRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(19, 1)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1900, 0)),
					),
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(9, 1)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(900, 0)),
					),
				}, lock)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.resetState()
			ctx = suite.ctx
			tc.check()
		})
	}
}
