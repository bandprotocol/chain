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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/restake/keeper"
	restaketestutil "github.com/bandprotocol/chain/v3/x/restake/testutil"
	"github.com/bandprotocol/chain/v3/x/restake/types"
)

var (
	ValidAddress1 = sdk.AccAddress("1000000001")
	ValidAddress2 = sdk.AccAddress("1000000002")
	ValidAddress3 = sdk.AccAddress("1000000003")

	VaultKeyWithRewards    = "0_key_with_rewards"
	VaultKeyWithoutRewards = "1_key_without_rewards"
	VaultKeyWithoutLocks   = "2_key_without_locks"
	InactiveVaultKey       = "3_inactive_key"
	InvalidVaultKey        = "invalid_key"
	ValidVaultKey          = "valid_key"

	VaultWithRewardsAddress    = sdk.AccAddress("2000000001")
	VaultWithoutRewardsAddress = sdk.AccAddress("2000000002")
	VaultWithoutLocksAddress   = sdk.AccAddress("2000000003")
	InactiveVaultAddress       = sdk.AccAddress("2000000004")

	ValidVaultAddress = "cosmos142hwqg2wwnverkcteaa5pn2lpkwp7e0ya2q84v4wdd7cffvfpeeq0zf2n6"

	RewarderAddress = sdk.AccAddress("3000000001")

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
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.validVaults = []types.Vault{
		{
			Key:             VaultKeyWithRewards,
			VaultAddress:    VaultWithRewardsAddress.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(20),
			RewardsPerPower: sdk.NewDecCoins(sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(1, 1))),
			Remainders:      nil,
		},
		{
			Key:             VaultKeyWithoutRewards,
			VaultAddress:    VaultWithoutRewardsAddress.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(100),
			RewardsPerPower: nil,
			Remainders:      nil,
		},
		{
			Key:             VaultKeyWithoutLocks,
			VaultAddress:    VaultWithoutLocksAddress.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(0),
			RewardsPerPower: nil,
			Remainders:      nil,
		},
		{
			Key:             InactiveVaultKey,
			VaultAddress:    InactiveVaultAddress.String(),
			IsActive:        false,
			TotalPower:      sdkmath.NewInt(100),
			RewardsPerPower: nil,
			Remainders:      nil,
		},
	}

	suite.validLocks = []types.Lock{
		{
			StakerAddress:  ValidAddress1.String(),
			Key:            VaultKeyWithRewards,
			Power:          sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			StakerAddress:  ValidAddress1.String(),
			Key:            VaultKeyWithoutRewards,
			Power:          sdkmath.NewInt(100),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			StakerAddress:  ValidAddress1.String(),
			Key:            InactiveVaultKey,
			Power:          sdkmath.NewInt(50),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			StakerAddress:  ValidAddress2.String(),
			Key:            VaultKeyWithRewards,
			Power:          sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
	}

	key := storetypes.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
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
	suite.stakingKeeper = stakingKeeper

	suite.restakeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
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
	for _, vault := range suite.validVaults {
		suite.restakeKeeper.SetVault(suite.ctx, vault)
	}

	for _, lock := range suite.validLocks {
		suite.restakeKeeper.SetLock(suite.ctx, lock)
	}
}

func (suite *KeeperTestSuite) TestScenarios() {
	testCases := []struct {
		name  string
		check func(sdk.Context)
	}{
		{
			name: "1 account",
			check: func(ctx sdk.Context) {
				// pre check
				_, found := suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().False(found)

				_, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().False(found)

				// --------------------------
				// address1 locks on key1 333 powers
				// --------------------------
				err := suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidVaultKey, sdkmath.NewInt(333))
				suite.Require().NoError(err)

				// post check
				// - total of key must be changed.
				// - lock of the user must be created.
				power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(333), power)

				vault, found := suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:             ValidVaultKey,
					VaultAddress:    ValidVaultAddress,
					IsActive:        true,
					RewardsPerPower: nil,
					TotalPower:      sdkmath.NewInt(333),
					Remainders:      nil,
				}, vault)

				lock, found := suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidVaultKey,
					Power:          sdkmath.NewInt(333),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// rewards in 1 aaaa, 1000 bbbb
				// --------------------------
				err = suite.restakeKeeper.AddRewards(ctx, RewarderAddress, ValidVaultKey, sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1000)),
				))
				suite.Require().NoError(err)

				// post check
				// - reward per powers must be changed.
				// - remainders must be calculated.
				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdkmath.NewInt(333),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, vault)

				// --------------------------
				// address1 locks on key1 100 powers (override)
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidVaultKey, sdkmath.NewInt(100))
				suite.Require().NoError(err)

				// post check
				// - locked power must be changed.
				// - total power of key must be changed.
				// - neg reward debts must be changed.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(100), power)

				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdkmath.NewInt(100),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, vault)

				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidVaultKey,
					Power:          sdkmath.NewInt(100),
					PosRewardDebts: nil,
					NegRewardDebts: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.699699699699699699")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("699.699699699699699699")),
					),
				}, lock)

				// --------------------------
				// address1 locks on key1 2000 powers (override)
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidVaultKey, sdkmath.NewInt(2000))
				suite.Require().NoError(err)

				// post check
				// - locked power must be changed.
				// - total power of key must be changed.
				// - pos reward debts must be changed.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(2000), power)

				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdkmath.NewInt(2000),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, vault)

				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress: ValidAddress1.String(),
					Key:           ValidVaultKey,
					Power:         sdkmath.NewInt(2000),
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
					SendCoins(gomock.Any(), sdk.MustAccAddressFromBech32(ValidVaultAddress), ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("bbbb", sdkmath.NewInt(999)),
					)).
					Times(1)

				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress1, ValidVaultKey))
				suite.Require().NoError(err)

				// post check
				// - reward debts need to be updated.
				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress: ValidAddress1.String(),
					Key:           ValidVaultKey,
					Power:         sdkmath.NewInt(2000),
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
				err = suite.restakeKeeper.DeactivateVault(ctx, ValidVaultKey)
				suite.Require().NoError(err)

				// post check
				// - status of key must be inactive
				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     false,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdkmath.NewInt(2000),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 18)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 18)),
					),
				}, vault)

				// --------------------------
				// claim rewards after vault is inactive
				// --------------------------
				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress1, ValidVaultKey))
				suite.Require().NoError(err)

				// post check
				// - lock must be deleted
				// - remainders must be 1
				_, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().False(found)

				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     false,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("0.003003003003003003")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("3.003003003003003003")),
					),
					TotalPower: sdkmath.NewInt(2000),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 0)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 0)),
					),
				}, vault)
			},
		},
		{
			name: "2 accounts",
			check: func(ctx sdk.Context) {
				// pre check
				_, found := suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().False(found)

				_, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().False(found)

				_, found = suite.restakeKeeper.GetLock(ctx, ValidAddress2, ValidVaultKey)
				suite.Require().False(found)

				// --------------------------
				// address1 locks on key1 10^18 powers
				// --------------------------
				val18, ok := sdkmath.NewIntFromString("1_000_000_000_000_000_000")
				suite.Require().True(ok)

				err := suite.restakeKeeper.SetLockedPower(
					ctx,
					ValidAddress1,
					ValidVaultKey,
					val18,
				)
				suite.Require().NoError(err)

				// post check
				// - total of key must be changed.
				// - lock of the user1 must be created.
				power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().NoError(err)
				suite.Require().Equal(val18, power)

				vault, found := suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:             ValidVaultKey,
					VaultAddress:    ValidVaultAddress,
					IsActive:        true,
					RewardsPerPower: nil,
					TotalPower:      val18,
					Remainders:      nil,
				}, vault)

				lock, found := suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidVaultKey,
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
					ValidVaultKey,
					sdkmath.NewInt(1),
				)
				suite.Require().NoError(err)

				// post check
				// - total of key must be changed.
				// - lock of the user1 must be created.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress2, ValidVaultKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(1), power)

				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:             ValidVaultKey,
					VaultAddress:    ValidVaultAddress,
					IsActive:        true,
					RewardsPerPower: nil,
					TotalPower:      val18Plus1,
					Remainders:      nil,
				}, vault)

				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress2, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress2.String(),
					Key:            ValidVaultKey,
					Power:          sdkmath.NewInt(1),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// rewards in 1 aaaa, 1e18 bbbb
				// --------------------------
				err = suite.restakeKeeper.AddRewards(ctx, RewarderAddress, ValidVaultKey, sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(1)),
					sdk.NewCoin("bbbb", val18),
				))
				suite.Require().NoError(err)

				// post check
				// - reward per powers must be changed.
				// - remainders must have "aaaa" as too much power for 1aaaa
				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.999999999999999999")),
					),
					TotalPower: val18Plus1,
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("1")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
					),
				}, vault)

				// --------------------------
				// address1 locks on key1 0 powers (override, remove all locked power)
				// --------------------------
				err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidVaultKey, sdkmath.NewInt(0))
				suite.Require().NoError(err)

				// post check
				// - total power of key must be changed.
				// - locked power must be changed.
				// - neg reward debts must be changed.
				power, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().NoError(err)
				suite.Require().Equal(sdkmath.NewInt(0), power)

				vault, found = suite.restakeKeeper.GetVault(ctx, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Vault{
					Key:          ValidVaultKey,
					VaultAddress: ValidVaultAddress,
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.999999999999999999")),
					),
					TotalPower: sdkmath.NewInt(1),
					Remainders: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyMustNewDecFromStr("1")),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
					),
				}, vault)

				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidVaultKey,
					Power:          sdkmath.NewInt(0),
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
					SendCoins(gomock.Any(), sdk.MustAccAddressFromBech32(ValidVaultAddress), ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("bbbb", sdkmath.NewInt(999999999999999999)),
					)).
					Times(1)

				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress1, ValidVaultKey))
				suite.Require().NoError(err)

				// post check
				// - reward debts need to be updated.
				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress1, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            ValidVaultKey,
					Power:          sdkmath.NewInt(0),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)

				// --------------------------
				// address2 claim rewards
				// --------------------------
				_, err = suite.msgServer.ClaimRewards(ctx, types.NewMsgClaimRewards(ValidAddress2, ValidVaultKey))
				suite.Require().NoError(err)

				// post check
				// - nothing change as reward isn't enough
				lock, found = suite.restakeKeeper.GetLock(ctx, ValidAddress2, ValidVaultKey)
				suite.Require().True(found)
				suite.Require().Equal(types.Lock{
					StakerAddress:  ValidAddress2.String(),
					Key:            ValidVaultKey,
					Power:          sdkmath.NewInt(1),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}, lock)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// setup delegator bond
			suite.stakingKeeper.EXPECT().
				GetDelegatorBonded(gomock.Any(), ValidAddress1).
				Return(sdkmath.NewInt(1e18), nil).
				AnyTimes()
			suite.stakingKeeper.EXPECT().
				GetDelegatorBonded(gomock.Any(), ValidAddress2).
				Return(sdkmath.NewInt(1e18), nil).
				AnyTimes()
			suite.stakingKeeper.EXPECT().
				GetDelegatorBonded(gomock.Any(), ValidAddress3).
				Return(sdkmath.NewInt(10), nil).
				AnyTimes()

			tc.check(suite.ctx)
		})
	}
}
