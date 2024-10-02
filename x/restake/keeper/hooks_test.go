package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestHooksAfterDelegationModified() {
	ctx := suite.ctx
	suite.setupState()

	// validator1 locked max at 100 power

	// change delegation to 100 -> success
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(100)).
		Times(1)
	err := suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// change delegation to 101 -> success
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(101)).
		Times(1)
	err = suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// change delegation to 99 -> failed
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(99)).
		Times(1)
	err = suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestHooksBeforeDelegationRemoved() {
	ctx := suite.ctx
	suite.setupState()

	// validator1 locked max at 100 power

	// set current delegation power as 200
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(200)).
		AnyTimes()

	// remove delegation 100 -> success (200 - 100 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegation(gomock.Any(), ValidAddress1, ValAddress).
		Return(stakingtypes.Delegation{
			DelegatorAddress: ValidAddress1.String(),
			ValidatorAddress: ValAddress.String(),
			Shares:           sdkmath.LegacyNewDec(10),
		}, true).
		Times(1)
	suite.stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), ValAddress).
		Return(stakingtypes.Validator{
			Tokens:          sdkmath.NewInt(1),
			DelegatorShares: sdkmath.LegacyNewDec(1),
		}, true).
		Times(1)
	err := suite.stakingHooks.BeforeDelegationRemoved(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// remove delegation 101 -> failed (200 - 101 < 100)
	suite.stakingKeeper.EXPECT().
		GetDelegation(gomock.Any(), ValidAddress1, ValAddress).
		Return(stakingtypes.Delegation{
			DelegatorAddress: ValidAddress1.String(),
			ValidatorAddress: ValAddress.String(),
			Shares:           sdkmath.LegacyNewDec(101),
		}, true).
		Times(1)
	suite.stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), ValAddress).
		Return(stakingtypes.Validator{
			Tokens:          sdkmath.NewInt(1),
			DelegatorShares: sdkmath.LegacyNewDec(1),
		}, true).
		Times(1)
	err = suite.stakingHooks.BeforeDelegationRemoved(ctx, ValidAddress1, ValAddress)
	suite.Require().Error(err)
}
