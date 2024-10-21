package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	"github.com/bandprotocol/chain/v3/x/restake/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestHooksAfterDelegationModified() {
	ctx := suite.ctx
	suite.setupState()

	// validator1 locked max at 100 power
	// staked power = 50

	// change delegation to 51 -> success (51 + 50 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(51), nil).
		Times(1)
	err := suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// change delegation to 50 -> success (50 + 50 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(50), nil).
		Times(1)
	err = suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().NoError(err)

	// change delegation to 49 -> failed (50 + 49 < 100)
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(49), nil).
		Times(1)
	err = suite.stakingHooks.AfterDelegationModified(ctx, ValidAddress1, ValAddress)
	suite.Require().ErrorIs(err, types.ErrUnableToUndelegate)
}

func (suite *KeeperTestSuite) TestHooksBeforeDelegationRemoved() {
	ctx := suite.ctx
	suite.setupState()

	// validator1 locked max at 100 power
	// staked power = 50

	// set current delegation power as 100
	suite.stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(100), nil).
		AnyTimes()

	// remove delegation 50 -> success (150 - 50 >= 100)
	suite.stakingKeeper.EXPECT().
		GetDelegation(gomock.Any(), ValidAddress1, ValAddress).
		Return(stakingtypes.Delegation{
			DelegatorAddress: ValidAddress1.String(),
			ValidatorAddress: ValAddress.String(),
			Shares:           sdkmath.LegacyNewDec(50),
		}, nil).
		Times(1)
	suite.stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), ValAddress).
		Return(stakingtypes.Validator{
			Tokens:          sdkmath.NewInt(1),
			DelegatorShares: sdkmath.LegacyNewDec(1),
		}, nil).
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
		}, nil).
		Times(1)
	suite.stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), ValAddress).
		Return(stakingtypes.Validator{
			Tokens:          sdkmath.NewInt(1),
			DelegatorShares: sdkmath.LegacyNewDec(1),
		}, nil).
		Times(1)
	err = suite.stakingHooks.BeforeDelegationRemoved(ctx, ValidAddress1, ValAddress)
	suite.Require().ErrorIs(err, types.ErrUnableToUndelegate)
}
