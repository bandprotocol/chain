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

	KeyWithRewards    = "0_key_with_rewards"
	KeyWithoutRewards = "1_key_without_rewards"
	KeyWithoutLocks   = "2_key_without_locks"
	InactiveKey       = "3_inactive_key"
	InvalidKey        = "invalid_key"
	ValidKey          = "valid_key"

	KeyWithRewardsPoolAddress    = sdk.AccAddress("2000000001")
	KeyWithoutRewardsPoolAddress = sdk.AccAddress("2000000002")
	KeyWithoutLocksPoolAddress   = sdk.AccAddress("2000000003")
	InactiveKeyPoolAddress       = sdk.AccAddress("2000000004")

	ValidKeyPoolAddress = "cosmos1jyea6nm7dhrfr8j4v2ethyfv4vvv2rcaz6a0sptfun7797g59vjsm53vhf"

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
			Name:            KeyWithRewards,
			PoolAddress:     KeyWithRewardsPoolAddress.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(20),
			RewardPerPowers: sdk.NewDecCoins(sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(1, 1))),
			Remainders:      nil,
		},
		{
			Name:            KeyWithoutRewards,
			PoolAddress:     KeyWithoutRewardsPoolAddress.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            KeyWithoutLocks,
			PoolAddress:     KeyWithoutLocksPoolAddress.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(0),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            InactiveKey,
			PoolAddress:     InactiveKeyPoolAddress.String(),
			IsActive:        false,
			TotalPower:      sdkmath.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
	}

	suite.validLocks = []types.Lock{
		{
			StakerAddress:  ValidAddress1.String(),
			Key:            KeyWithRewards,
			Power:          sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			StakerAddress:  ValidAddress1.String(),
			Key:            KeyWithoutRewards,
			Power:          sdkmath.NewInt(100),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			StakerAddress:  ValidAddress1.String(),
			Key:            InactiveKey,
			Power:          sdkmath.NewInt(50),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			StakerAddress:  ValidAddress2.String(),
			Key:            KeyWithRewards,
			Power:          sdkmath.NewInt(10),
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
		SendCoins(gomock.Any(), RewarderAddress, gomock.Any(), gomock.Any()).
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
			name: "1 account",
			check: func() {
				// pre check
				_, err := suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().Error(err)

				_, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().Error(err)

				// --------------------------
				// address1 locks on key1 333 powers
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey, sdkmath.NewInt(333))
				suite.Require().NoError(err)

				// post check
				// - total of key must be changed.
				// - lock of the user must be created.
				power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(333), power)

				key, err := suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:            ValidKey,
					PoolAddress:     ValidKeyPoolAddress,
					IsActive:        true,
					RewardPerPowers: nil,
					TotalPower:      sdk.NewInt(333),
					Remainders:      nil,
				}, key)

				lock, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidKey,
					Power:          sdk.NewInt(333),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// rewards in 1 aaaa, 1000 bbbb
				// --------------------------
				err = suite.restakeKeeper.AddRewards(ctx, RewarderAddress, ValidKey, sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1000)),
				))
				suite.Require().NoError(err)

				// post check
				// - reward per powers must be changed.
				// - remainders must be calculated.
				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdk.NewInt(333),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, key)

				// --------------------------
				// address1 locks on key1 100 powers (override)
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey, sdkmath.NewInt(100))
				suite.Require().NoError(err)

				// post check
				// - locked power must be changed.
				// - total power of key must be changed.
				// - neg reward debts must be changed.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(100), power)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdk.NewInt(100),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidKey,
					Power:          sdk.NewInt(100),
					PosRewardDebts: nil,
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.699699699699699699")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("699.699699699699699699")),
					),
				}, lock)

				// --------------------------
				// address1 locks on key1 2000 powers (override)
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey, sdkmath.NewInt(2000))
				suite.Require().NoError(err)

				// post check
				// - locked power must be changed.
				// - total power of key must be changed.
				// - pos reward debts must be changed.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(2000), power)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdk.NewInt(2000),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress: ValidAddress1.String(),
					Key:           ValidKey,
					Power:         sdk.NewInt(2000),
					PosRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("5.705705705705705700")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("5705.705705705705705700")),
					),
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.699699699699699699")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("699.699699699699699699")),
					),
				}, lock)

				// --------------------------
				// claim rewards
				// --------------------------

				// rewards needs to be transfer from pool address to user
				suite.bankKeeper.EXPECT().
					SendCoins(gomock.Any(), sdk.MustAccAddressFromBech32(ValidKeyPoolAddress), ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("bbbb", sdkmath.NewInt(999)),
					)).
					Times(1)

				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress1, ValidKey))
				suite.Require().NoError(err)

				// post check
				// - reward debts need to be updated.
				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress: ValidAddress1.String(),
					Key:           ValidKey,
					Power:         sdk.NewInt(2000),
					PosRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("6.006006006006006000")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("6006.006006006006006000")),
					),
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.999999999999999999")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.999999999999999999")),
					),
				}, lock)

				// --------------------------
				// deactivate keys
				// --------------------------
				err = suite.restakeKeeper.DeactivateKey(ctx, ValidKey)
				suite.Require().NoError(err)

				// post check
				// - status of key must be inactive
				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    false,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdk.NewInt(2000),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, key)

				// --------------------------
				// claim rewards after key is inactive
				// --------------------------
				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress1, ValidKey))
				suite.Require().NoError(err)

				// post check
				// - lock must be deleted
				// - remainders must be 1
				_, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().Error(err)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    false,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdk.NewInt(2000),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 0)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 0)),
					),
				}, key)
			},
		},
		{
			name: "2 accounts",
			check: func() {
				// pre check
				_, err := suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().Error(err)

				_, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().Error(err)

				_, err = suite.restakeKeeper.GetLock(ctx, ValidAddress2, ValidKey)
				suite.Require().Error(err)

				// --------------------------
				// address1 locks on key1 10^18 powers
				// --------------------------
				val18, ok := sdkmath.NewIntFromString("1_000_000_000_000_000_000")
				suite.Require().True(ok)

				err = suite.restakeKeeper.SetLockedPower(
					ctx,
					ValidAddress1,
					ValidKey,
					val18,
				)
				suite.Require().NoError(err)

				// post check
				// - total of key must be changed.
				// - lock of the user1 must be created.
				power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(val18, power)

				key, err := suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:            ValidKey,
					PoolAddress:     "cosmos1jyea6nm7dhrfr8j4v2ethyfv4vvv2rcaz6a0sptfun7797g59vjsm53vhf",
					IsActive:        true,
					RewardPerPowers: nil,
					TotalPower:      val18,
					Remainders:      nil,
				}, key)

				lock, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidKey,
					Power:          val18,
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// address2 locks on key1 1 powers
				// --------------------------
				val18Plus1 := val18.Add(sdkmath.NewInt(1))
				suite.Require().True(ok)

				err = suite.restakeKeeper.SetLockedPower(
					ctx,
					ValidAddress2,
					ValidKey,
					sdk.NewInt(1),
				)
				suite.Require().NoError(err)

				// post check
				// - total of key must be changed.
				// - lock of the user1 must be created.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress2, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewInt(1), power)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:            ValidKey,
					PoolAddress:     ValidKeyPoolAddress,
					IsActive:        true,
					RewardPerPowers: nil,
					TotalPower:      val18Plus1,
					Remainders:      nil,
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress2, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress2.String(),
					Key:            ValidKey,
					Power:          sdk.NewInt(1),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// rewards in 1 aaaa, 1e18 bbbb
				// --------------------------
				err = suite.restakeKeeper.AddRewards(ctx, RewarderAddress, ValidKey, sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", val18),
				))
				suite.Require().NoError(err)

				// post check
				// - reward per powers must be changed.
				// - remainders must have "aaaa" as too much power for 1aaaa
				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.999999999999999999")),
					),
					TotalPower: val18Plus1,
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("1")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
					),
				}, key)

				// --------------------------
				// address1 locks on key1 0 powers (override, remove all locked power)
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey, sdkmath.NewInt(0))
				suite.Require().NoError(err)

				// post check
				// - total power of key must be changed.
				// - locked power must be changed.
				// - neg reward debts must be changed.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(0), power)

				key, err = suite.restakeKeeper.GetKey(ctx, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Key{
					Name:        ValidKey,
					PoolAddress: ValidKeyPoolAddress,
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.999999999999999999")),
					),
					TotalPower: sdk.NewInt(1),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("1")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
					),
				}, key)

				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidKey,
					Power:          sdk.NewInt(0),
					PosRewardDebts: nil,
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("999999999999999999")),
					),
				}, lock)

				// --------------------------
				// address1 claim rewards
				// --------------------------

				// rewards needs to be transfer from pool address to address1
				suite.bankKeeper.EXPECT().
					SendCoins(gomock.Any(), sdk.MustAccAddressFromBech32(ValidKeyPoolAddress), ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("bbbb", sdk.NewInt(999999999999999999)),
					)).
					Times(1)

				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress1, ValidKey))
				suite.Require().NoError(err)

				// post check
				// - reward debts need to be updated.
				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidKey,
					Power:          sdk.NewInt(0),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// address2 claim rewards
				// --------------------------
				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress2, ValidKey))
				suite.Require().NoError(err)

				// post check
				// - nothing change as reward isn't enough
				lock, err = suite.restakeKeeper.GetLock(ctx, ValidAddress2, ValidKey)
				suite.Require().NoError(err)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress2.String(),
					Key:            ValidKey,
					Power:          sdk.NewInt(1),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
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
