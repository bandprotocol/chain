package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestMsgClaimRewards() {
	ctx := suite.ctx
	suite.setupState()

	testCases := []struct {
		name      string
		input     *types.MsgClaimRewards
		expErr    bool
		expErrMsg string
		preCheck  func()
		postCheck func()
	}{
		{
			name: "no key",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress1.String(),
				Key:           InvalidKey,
			},
			expErr:    true,
			expErrMsg: "key not found",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "no lock",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress2.String(),
				Key:           KeyWithoutRewards,
			},
			expErr:    true,
			expErrMsg: "lock not found",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "success - active key",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress1.String(),
				Key:           KeyWithRewards,
			},
			expErr:    false,
			expErrMsg: "",
			preCheck: func() {
				suite.bankKeeper.EXPECT().
					SendCoins(gomock.Any(), KeyWithRewardsPoolAddress, ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("uband", sdk.NewInt(1)),
					)).
					Times(1)
			},
			postCheck: func() {
				lock, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, KeyWithRewards)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))), lock.PosRewardDebts)
				suite.Require().Equal(sdk.DecCoins(nil), lock.NegRewardDebts)
			},
		},
		{
			name: "success - inactive key",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress1.String(),
				Key:           InactiveKey,
			},
			expErr:    false,
			expErrMsg: "",
			preCheck:  func() {},
			postCheck: func() {
				_, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, InactiveKey)
				suite.Require().Error(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preCheck()
			_, err := suite.msgServer.ClaimRewards(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}

			tc.postCheck()
		})
	}
}
