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
			name: "no vault",
			input: &types.MsgClaimRewards{
				StakerAddress: ValidAddress1.String(),
				Key:           InvalidVaultKey,
			},
			expErr:    true,
			expErrMsg: "vault not found",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "no lock",
			input: &types.MsgClaimRewards{
				StakerAddress: ValidAddress2.String(),
				Key:           VaultKeyWithoutRewards,
			},
			expErr:    true,
			expErrMsg: "lock not found",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "success - active vault",
			input: &types.MsgClaimRewards{
				StakerAddress: ValidAddress1.String(),
				Key:           VaultKeyWithRewards,
			},
			expErr:    false,
			expErrMsg: "",
			preCheck: func() {
				suite.bankKeeper.EXPECT().
					SendCoins(gomock.Any(), VaultWithRewardsAddress, ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("uband", sdk.NewInt(1)),
					)).
					Times(1)
			},
			postCheck: func() {
				lock, found := suite.restakeKeeper.GetLock(ctx, ValidAddress1, VaultKeyWithRewards)
				suite.Require().True(found)
				suite.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))), lock.PosRewardDebts)
				suite.Require().Equal(sdk.DecCoins(nil), lock.NegRewardDebts)
			},
		},
		{
			name: "success - inactive vault",
			input: &types.MsgClaimRewards{
				StakerAddress: ValidAddress1.String(),
				Key:           InactiveVaultKey,
			},
			expErr:    false,
			expErrMsg: "",
			preCheck:  func() {},
			postCheck: func() {
				_, found := suite.restakeKeeper.GetLock(ctx, ValidAddress1, InactiveVaultKey)
				suite.Require().False(found)
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
