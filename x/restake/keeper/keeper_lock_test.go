package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestSetLockedPower() {
	ctx := suite.ctx

	// error case -  amount is not uint64
	err := suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ValidKey1, sdkmath.NewInt(-5))
	suite.Require().Error(err)

	// error case - lock more than delegation
	err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress3, ValidKey1, sdkmath.NewInt(30))
	suite.Require().Error(err)

	// error case - key is deactivated
	_, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress2, ValidKey3)
	suite.Require().Error(err)

	// success cases
	var (
		preKey  types.Key
		preLock *types.Lock
		amount  sdkmath.Int
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
				preKey = types.Key{
					Name:            ValidKey1,
					PoolAddress:     ValidPoolAddress1.String(),
					IsActive:        true,
					RewardPerPowers: nil,
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      nil,
				}

				preLock = nil

				amount = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(200),
			types.Lock{
				LockerAddress:  ValidAddress1.String(),
				Key:            ValidKey1,
				Amount:         sdkmath.NewInt(100),
				PosRewardDebts: nil,
				NegRewardDebts: nil,
			},
		},
		{
			"success case - have previous lock with empty rewards",
			func() {
				preKey = types.Key{
					Name:            ValidKey1,
					PoolAddress:     ValidPoolAddress1.String(),
					IsActive:        true,
					RewardPerPowers: nil,
					TotalPower:      sdkmath.NewInt(100),
					Remainders:      nil,
				}

				preLock = &types.Lock{
					LockerAddress:  ValidAddress1.String(),
					Key:            ValidKey1,
					Amount:         sdkmath.NewInt(10),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}

				amount = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(190),
			types.Lock{
				LockerAddress:  ValidAddress1.String(),
				Key:            ValidKey1,
				Amount:         sdkmath.NewInt(100),
				PosRewardDebts: nil,
				NegRewardDebts: nil,
			},
		},
		{
			"success case - no previous lock with rewards",
			func() {
				preKey = types.Key{
					Name:        ValidKey1,
					PoolAddress: ValidPoolAddress1.String(),
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
					),
					TotalPower: sdkmath.NewInt(100),
					Remainders: nil,
				}

				preLock = nil

				amount = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(200),
			types.Lock{
				LockerAddress: ValidAddress1.String(),
				Key:           ValidKey1,
				Amount:        sdkmath.NewInt(100),
				PosRewardDebts: sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(0)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(1)),
				),
				NegRewardDebts: nil,
			},
		},
		{
			"success case - have previous lock with rewards - lock more",
			func() {
				preKey = types.Key{
					Name:        ValidKey1,
					PoolAddress: ValidPoolAddress1.String(),
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
					),
					TotalPower: sdkmath.NewInt(100),
					Remainders: nil,
				}

				preLock = &types.Lock{
					LockerAddress:  ValidAddress1.String(),
					Key:            ValidKey1,
					Amount:         sdkmath.NewInt(100),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}

				amount = sdkmath.NewInt(1000)
			},
			sdkmath.NewInt(1000),
			types.Lock{
				LockerAddress: ValidAddress1.String(),
				Key:           ValidKey1,
				Amount:        sdkmath.NewInt(1000),
				PosRewardDebts: sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(0)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(9)),
				),
				NegRewardDebts: nil,
			},
		},
		{
			"success case - have previous lock with rewards - lock less",
			func() {
				preKey = types.Key{
					Name:        ValidKey1,
					PoolAddress: ValidPoolAddress1.String(),
					IsActive:    true,
					RewardPerPowers: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("aaaa", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("bbbb", sdkmath.LegacyNewDecWithPrec(1, 2)),
					),
					TotalPower: sdkmath.NewInt(1000),
					Remainders: nil,
				}

				preLock = &types.Lock{
					LockerAddress:  ValidAddress1.String(),
					Key:            ValidKey1,
					Amount:         sdkmath.NewInt(1000),
					PosRewardDebts: nil,
					NegRewardDebts: nil,
				}

				amount = sdkmath.NewInt(100)
			},
			sdkmath.NewInt(100),
			types.Lock{
				LockerAddress:  ValidAddress1.String(),
				Key:            ValidKey1,
				Amount:         sdkmath.NewInt(100),
				PosRewardDebts: nil,
				NegRewardDebts: sdk.NewCoins(
					sdk.NewCoin("aaaa", sdkmath.NewInt(0)),
					sdk.NewCoin("bbbb", sdkmath.NewInt(9)),
				),
			},
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.name), func() {
			suite.resetState()
			ctx = suite.ctx
			testCase.malleate()

			suite.restakeKeeper.SetKey(ctx, preKey)

			if preLock != nil {
				suite.restakeKeeper.SetLock(ctx, *preLock)
			}

			err = suite.restakeKeeper.SetLockedPower(
				ctx,
				ValidAddress1,
				preKey.Name,
				amount,
			)
			suite.Require().NoError(err)

			key, err := suite.restakeKeeper.GetKey(ctx, preKey.Name)
			suite.Require().NoError(err)
			suite.Require().Equal(testCase.expTotalPower, key.TotalPower)

			lock, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, preKey.Name)
			suite.Require().NoError(err)
			suite.Require().Equal(testCase.expLock, lock)
		})
	}
}

func (suite *KeeperTestSuite) TestGetLockedPower() {
	ctx := suite.ctx
	suite.setupState()

	// error case -  no key
	_, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, InvalidKey)
	suite.Require().Error(err)

	// error case - key is deactivated
	_, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey3)
	suite.Require().Error(err)

	// error case - no lock
	_, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress2, ValidKey2)
	suite.Require().Error(err)

	// success case
	power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ValidKey1)
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(10), power)
}

func (suite *KeeperTestSuite) TestGetSetLock() {
	ctx := suite.ctx

	// set
	expectedLocks := suite.validLocks
	for _, expLock := range expectedLocks {
		acc := sdk.MustAccAddressFromBech32(expLock.LockerAddress)
		suite.restakeKeeper.SetLock(ctx, expLock)

		// has
		has := suite.restakeKeeper.HasLock(ctx, acc, expLock.Key)
		suite.Require().True(has)

		// get
		lock, err := suite.restakeKeeper.GetLock(ctx, acc, expLock.Key)
		suite.Require().NoError(err)
		suite.Require().Equal(expLock, lock)

		// get lock by amount
		keyName := ctx.KVStore(suite.storeKey).Get(types.LockByAmountIndexKey(lock))
		suite.Require().Equal(expLock.Key, string(keyName))
	}

	// has
	has := suite.restakeKeeper.HasLock(ctx, ValidAddress1, "nonKey")
	suite.Require().False(has)

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
		acc := sdk.MustAccAddressFromBech32(expLock.LockerAddress)
		suite.restakeKeeper.DeleteLock(ctx, acc, expLock.Key)

		// has
		has := suite.restakeKeeper.HasLock(ctx, acc, expLock.Key)
		suite.Require().False(has)

		// get
		_, err := suite.restakeKeeper.GetLock(ctx, acc, expLock.Key)
		suite.Require().Error(err)

		// get lock by amount
		has = ctx.KVStore(suite.storeKey).Has(types.LockByAmountIndexKey(expLock))
		suite.Require().False(has)
	}
}
