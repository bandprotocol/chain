package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

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
	suite.Require().Equal(expectedLocks[:2], locks)

	locks = suite.restakeKeeper.GetLocksByAddress(ctx, ValidAddress2)
	suite.Require().Equal(expectedLocks[2:3], locks)

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
