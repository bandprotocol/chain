package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestMsgClaimRewards() {
	ctx := suite.ctx

	// setup
	for _, key := range suite.validKeys {
		suite.restakeKeeper.SetKey(ctx, key)
	}
	for _, lock := range suite.validLocks {
		suite.restakeKeeper.SetLock(ctx, lock)
	}

	testCases := []struct {
		name      string
		input     *types.MsgClaimRewards
		expErr    bool
		expErrMsg string
		postCheck func()
	}{
		{
			name: "no key",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress1.String(),
				Key:           "nonKey",
			},
			expErr:    true,
			expErrMsg: "key not found",
			postCheck: func() {},
		},
		{
			name: "no lock",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress2.String(),
				Key:           "Key1",
			},
			expErr:    true,
			expErrMsg: "lock not found",
			postCheck: func() {},
		},
		{
			name: "valid request",
			input: &types.MsgClaimRewards{
				LockerAddress: ValidAddress1.String(),
				Key:           "Key0",
			},
			expErr:    false,
			expErrMsg: "",
			postCheck: func() {
				lock, err := suite.restakeKeeper.GetLock(ctx, ValidAddress1, "Key0")
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1))), lock.PosRewardDebts)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
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
