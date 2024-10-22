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
	err := suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, ActiveVaultKey, sdkmath.NewInt(-5))
	suite.Require().ErrorIs(err, types.ErrInvalidPower)

	// error case - lock more than delegation
	err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress3, ActiveVaultKey, sdkmath.NewInt(30))
	suite.Require().ErrorIs(err, types.ErrPowerNotEnough)

	// error case - vault is deactivated
	err = suite.restakeKeeper.SetLockedPower(ctx, ValidAddress1, InactiveVaultKey, sdkmath.NewInt(10))
	suite.Require().ErrorIs(err, types.ErrVaultNotActive)

	// error case - staker is liquid staker
	err = suite.restakeKeeper.SetLockedPower(ctx, LiquidStakerAddress, ActiveVaultKey, sdkmath.NewInt(10))
	suite.Require().ErrorIs(err, types.ErrLiquidStakerNotAllowed)

	// success cases
	var (
		preVault types.Vault
		preLock  *types.Lock
		power    sdkmath.Int
	)

	testCases := []struct {
		name     string
		malleate func()
		expLock  types.Lock
	}{
		{
			"success case - no previous lock",
			func() {
				preVault = types.Vault{
					Key:      ActiveVaultKey,
					IsActive: true,
				}

				preLock = nil

				power = sdkmath.NewInt(100)
			},
			types.Lock{
				StakerAddress: ValidAddress1.String(),
				Key:           ActiveVaultKey,
				Power:         sdkmath.NewInt(100),
			},
		},
		{
			"success case - have previous lock",
			func() {
				preVault = types.Vault{
					Key:      ActiveVaultKey,
					IsActive: true,
				}

				preLock = &types.Lock{
					StakerAddress: ValidAddress1.String(),
					Key:           ActiveVaultKey,
					Power:         sdkmath.NewInt(10),
				}

				power = sdkmath.NewInt(100)
			},
			types.Lock{
				StakerAddress: ValidAddress1.String(),
				Key:           ActiveVaultKey,
				Power:         sdkmath.NewInt(100),
			},
		},
		{
			"success case - no vault",
			func() {
				preVault = types.Vault{
					Key:      "newVault",
					IsActive: true,
				}

				preLock = nil

				power = sdkmath.NewInt(100)
			},
			types.Lock{
				StakerAddress: ValidAddress1.String(),
				Key:           "newVault",
				Power:         sdkmath.NewInt(100),
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

			_, found := suite.restakeKeeper.GetVault(ctx, preVault.Key)
			suite.Require().True(found)

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
	_, err = suite.restakeKeeper.GetLockedPower(ctx, ValidAddress3, ActiveVaultKey)
	suite.Require().Error(err)

	// error case - staker is liquid staker
	_, err = suite.restakeKeeper.GetLockedPower(ctx, LiquidStakerAddress, ActiveVaultKey)
	suite.Require().Error(err)

	// success case
	power, err := suite.restakeKeeper.GetLockedPower(ctx, ValidAddress1, ActiveVaultKey)
	suite.Require().NoError(err)
	suite.Require().Equal(sdkmath.NewInt(100), power)
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
	suite.Require().Equal(expectedLocks[:2], locks)

	locks = suite.restakeKeeper.GetLocksByAddress(ctx, ValidAddress2)
	suite.Require().Equal(expectedLocks[2:3], locks)

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
