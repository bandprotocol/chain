package keeper_test

import (
	"fmt"

	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

func (suite *KeeperTestSuite) TestSetLockedPower() {
	ctx := suite.ctx
	suite.setupState()

	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(1e18), nil).
		Times(1)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress3).
		Return(sdkmath.NewInt(10), nil).
		Times(1)

	// error case -  power is not uint64
	err := suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, VaultKeyWithRewards, sdkmath.NewInt(-5))
	suite.Require().Error(err)

	// error case - lock more than delegation
	err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress3, VaultKeyWithRewards, sdkmath.NewInt(30))
	suite.Require().Error(err)

	// error case - vault is deactivated
	err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, InactiveVaultKey, sdkmath.NewInt(10))
	suite.Require().Error(err)

	// error case - staker is liquid staker
	err = suite.restakeKeeper.SetLockedPower(ctx, LiquidStakerAddress, VaultKeyWithRewards, sdkmath.NewInt(10))
	suite.Require().Error(err)

	// success cases
	var (
		preVault types.Vault
		preLock  *types.Lock
		power    sdkmath.Int
	)

	testCases := []struct {
		name          string
		malleate      func()
		expTotalPower sdkmath.Int
		expLock       types.Lock
	}{
		{
			"success case - no previous lock with empty rewards",
			func() {
				preVault = types.Vault{
					Key:             VaultKeyWithoutRewards,
					VaultAddress:    VaultWithoutRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: nil,
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      nil,
				}

				preLock = nil

				power = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(200),
			types.Lock{
				StakerAddress:  ValidAddress1.String(),
				Key:            VaultKeyWithoutRewards,
				Power:          sdkmath.NewInt(100),
				PosRewardDebts: nil,
				NegRewardDebts: nil,
			},
		},
		{
			"success case - have previous lock with empty rewards",
			func() {
				preVault = types.Vault{
					Key:             VaultKeyWithoutRewards,
					VaultAddress:    VaultWithoutRewardsAddress.String(),
					IsActive:        true,
					RewardsPerPower: nil,
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      nil,
				}

				preLock = &types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            VaultKeyWithoutRewards,
					Power:          sdkmath.NewInt(10),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}

				power = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(190),
			types.Lock{
				StakerAddress:  ValidAddress1.String(),
				Key:            VaultKeyWithoutRewards,
				Power:          sdkmath.NewInt(100),
				PosRewardDebts: nil,
				NegRewardDebts: nil,
			},
		},
		{
			"success case - no previous lock with rewards",
			func() {
				preVault = types.Vault{
					Key:          VaultKeyWithRewards,
					VaultAddress: VaultWithRewardsAddress.String(),
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
					),
					TotalPower: sdkmath.NewInt(100),
					Remainders: nil,
				}

				preLock = nil

				power = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(200),
			types.Lock{
				StakerAddress: ValidAddress1.String(),
				Key:           VaultKeyWithRewards,
				Power:         sdkmath.NewInt(100),
				PosRewardDebts: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 1)),
					sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 0)),
				),
				NegRewardDebts: nil,
			},
		},
		{
			"success case - have previous lock with rewards - lock more",
			func() {
				preVault = types.Vault{
					Key:          VaultKeyWithRewards,
					VaultAddress: VaultWithRewardsAddress.String(),
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
					),
					TotalPower: sdkmath.NewInt(100),
					Remainders: nil,
				}

				preLock = &types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            VaultKeyWithRewards,
					Power:          sdkmath.NewInt(100),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}

				power = sdkmath.NewInt(1000)
			},
			sdkmath.NewInt(1000),
			types.Lock{
				StakerAddress: ValidAddress1.String(),
				Key:           VaultKeyWithRewards,
				Power:         sdkmath.NewInt(1000),
				PosRewardDebts: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(9, 1)),
					sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(9, 0)),
				),
				NegRewardDebts: nil,
			},
		},
		{
			"success case - have previous lock with rewards - lock less",
			func() {
				preVault = types.Vault{
					Key:          VaultKeyWithRewards,
					VaultAddress: VaultWithRewardsAddress.String(),
					IsActive:     true,
					RewardsPerPower: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
					),
					TotalPower: sdkmath.NewInt(1000),
					Remainders: nil,
				}

				preLock = &types.Lock{
					StakerAddress:  ValidAddress1.String(),
					Key:            VaultKeyWithRewards,
					Power:          sdkmath.NewInt(1000),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}

				power = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(100),
			types.Lock{
				StakerAddress:  ValidAddress1.String(),
				Key:            VaultKeyWithRewards,
				Power:          sdkmath.NewInt(100),
				PosRewardDebts: nil,
				NegRewardDebts: sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(9, 1)),
					sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(9, 0)),
				),
			},
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.name), func() {
			suite.SetupTest()
			ctx = suite.ctx

			suite.stakingKeeper.EXPECT().
				GetDelegatorBonded(gomock.Any(), ValidAddress1).
				Return(sdkmath.NewInt(1e18), nil).
				Times(1)

			testCase.malleate()

			suite.restakeKeeper.SetVault(ctx, preVault)

			if preLock != nil {
				suite.restakeKeeper.SetLock(ctx, *preLock)
			}

			err = suite.restakeKeeper.SetLockedPower(
				ctx,
				ValidAddress1,
				preVault.Key,
				power,
			)
			suite.Require().NoError(err)

			vault, found := suite.restakeKeeper.GetVault(ctx, preVault.Key)
			suite.Require().True(found)
			suite.Require().Equal(testCase.expTotalPower, vault.TotalPower)

			lock, found := suite.restakeKeeper.GetLock(ctx, ValidAddress1, preVault.Key)
			suite.Require().True(found)
			suite.Require().Equal(testCase.expLock, lock)
		})
	}
}

func (suite *KeeperTestSuite) TestGetLockedPower() {
	ctx := suite.ctx
	suite.setupState()

	// error case -  no vault
	_, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, InvalidVaultKey)
	suite.Require().Error(err)

	// error case - vault is deactivated
	_, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, InactiveVaultKey)
	suite.Require().Error(err)

	// error case - no lock
	_, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress2, VaultKeyWithoutRewards)
	suite.Require().Error(err)

	// error case - staker is liquid staker
	_, err = suite.restakeKeeper.GetLockedPower(ctx, LiquidStakerAddress, VaultKeyWithRewards)
	suite.Require().Error(err)

	// success case
	power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, VaultKeyWithRewards)
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(10), power)
}

func (suite *KeeperTestSuite) TestGetSetLock() {
	ctx := suite.ctx

	// set
	expectedLocks := suite.validLocks
	for _, expLock := range expectedLocks {
		acc := sdk.MustAccAddressFromBech32(expLock.StakerAddress)
		suite.restakeKeeper.SetLock(ctx, expLock)

		// get
		lock, found := suite.restakeKeeper.GetLock(ctx, acc, expLock.Key)
		suite.Require().True(found)
		suite.Require().Equal(expLock, lock)

		// get lock by power
		key := ctx.KVStore(suite.storeKey).Get(types.LockByPowerIndexKey(lock))
		suite.Require().Equal(expLock.Key, string(key))
	}

	// get
	locks := suite.restakeKeeper.GetLocks(ctx)
	suite.Require().Equal(expectedLocks, locks)

	locks = suite.restakeKeeper.GetLocksByAddress(ctx, ValidAddress1)
	suite.Require().Equal(expectedLocks[:3], locks)

	locks = suite.restakeKeeper.GetLocksByAddress(ctx, ValidAddress2)
	suite.Require().Equal(expectedLocks[3:4], locks)

	locks = suite.restakeKeeper.GetLocksByAddress(ctx, ValidAddress3)
	suite.Require().Equal([]types.Lock(nil), locks)

	// delete
	for _, expLock := range expectedLocks {
		acc := sdk.MustAccAddressFromBech32(expLock.StakerAddress)
		suite.restakeKeeper.DeleteLock(ctx, acc, expLock.Key)

		// get
		_, found := suite.restakeKeeper.GetLock(ctx, acc, expLock.Key)
		suite.Require().False(found)

		// get lock by Power
		has := ctx.KVStore(suite.storeKey).Has(types.LockByPowerIndexKey(expLock))
		suite.Require().False(has)
	}
}
