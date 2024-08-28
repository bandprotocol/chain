package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/mock/gomock"
)

func (suite *KeeperTestSuite) TestHooksAfterDelegationModified() {
	ctx := suite.ctx
	suite.setupState()

	// validator1 locked max at 100 power
	// staked power = 50

	// change delegation to 51 -> success (51 + 50 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(51)).
		Times(1)
	err := suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// change delegation to 50 -> success (50 + 50 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(50)).
		Times(1)
	err = suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// change delegation to 49 -> failed (50 + 49 < 100)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(49)).
		Times(1)
	err = suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestHooksBeforeDelegationRemoved() {
	ctx := suite.ctx
	suite.setupState()

	// validator1 locked max at 100 power
	// staked power = 50

	// set current delegation power as 100
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(100)).
		AnyTimes()

	// remove delegation 50 -> success (150 - 50 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegation(gomock.Any(), ValidAddress1, ValAddress).
		Return(stakingtypes.Delegation{
			DelegatorAddress: ValidAddress1.String(),
			ValidatorAddress: ValAddress.String(),
			Shares:           sdkmath.LegacyNewDec(50),
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

	// remove delegation 51 -> failed (150 - 51 < 100)
	suite.stakingKeeper.EXPECT().
		GetDelegation(gomock.Any(), ValidAddress1, ValAddress).
		Return(stakingtypes.Delegation{
			DelegatorAddress: ValidAddress1.String(),
			ValidatorAddress: ValAddress.String(),
			Shares:           sdkmath.LegacyNewDec(51),
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
