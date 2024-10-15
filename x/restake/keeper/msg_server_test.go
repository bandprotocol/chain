package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
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
						sdk.NewCoin("uband", sdkmath.NewInt(1)),
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

func (suite *KeeperTestSuite) TestMsgStake() {
	ctx := suite.ctx
	suite.setupState()

	testCases := []struct {
		name      string
		input     *types.MsgStake
		expErr    bool
		expErrMsg string
		preCheck  func()
		postCheck func()
	}{
		{
			name: "not allowed denoms",
			input: &types.MsgStake{
				StakerAddress: ValidAddress1.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("usdt", sdkmath.NewInt(10)),
				),
			},
			expErr:    true,
			expErrMsg: "not allowed denom",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "mix both allow and unallow denom",
			input: &types.MsgStake{
				StakerAddress: ValidAddress1.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("usdt", sdkmath.NewInt(10)),
					sdk.NewCoin("uband", sdkmath.NewInt(10)),
				),
			},
			expErr:    true,
			expErrMsg: "not allowed denom",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "success - have previous stake",
			input: &types.MsgStake{
				StakerAddress: ValidAddress1.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("uband", sdkmath.NewInt(10)),
				),
			},
			expErr:    false,
			expErrMsg: "",
			preCheck: func() {
				suite.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), ValidAddress1, types.ModuleName, sdk.NewCoins(
						sdk.NewCoin("uband", sdkmath.NewInt(10)),
					)).
					Return(nil).
					Times(1)
			},
			postCheck: func() {
				stake := suite.restakeKeeper.GetStake(ctx, ValidAddress1)
				suite.Require().Equal(types.Stake{
					StakerAddress: ValidAddress1.String(),
					Coins: sdk.NewCoins(
						sdk.NewCoin("uband", sdkmath.NewInt(60)),
					),
				}, stake)
			},
		},
		{
			name: "success - no previous stake",
			input: &types.MsgStake{
				StakerAddress: ValidAddress2.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("uband", sdkmath.NewInt(10)),
				),
			},
			expErr:    false,
			expErrMsg: "",
			preCheck: func() {
				suite.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), ValidAddress2, types.ModuleName, sdk.NewCoins(
						sdk.NewCoin("uband", sdkmath.NewInt(10)),
					)).
					Return(nil).
					Times(1)
			},
			postCheck: func() {
				stake := suite.restakeKeeper.GetStake(ctx, ValidAddress2)
				suite.Require().Equal(types.Stake{
					StakerAddress: ValidAddress2.String(),
					Coins: sdk.NewCoins(
						sdk.NewCoin("uband", sdkmath.NewInt(10)),
					),
				}, stake)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preCheck()
			_, err := suite.msgServer.Stake(suite.ctx, tc.input)

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

func (suite *KeeperTestSuite) TestMsgUnstake() {
	ctx := suite.ctx
	suite.setupState()

	testCases := []struct {
		name      string
		input     *types.MsgUnstake
		expErr    bool
		expErrMsg string
		preCheck  func()
		postCheck func()
	}{
		{
			name: "unstake more than staked coins",
			input: &types.MsgUnstake{
				StakerAddress: ValidAddress1.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("uband", sdkmath.NewInt(2000)),
				),
			},
			expErr:    true,
			expErrMsg: "stake not enough",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "success",
			input: &types.MsgUnstake{
				StakerAddress: ValidAddress1.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin("uband", sdkmath.NewInt(10)),
				),
			},
			expErr:    false,
			expErrMsg: "",
			preCheck: func() {
				suite.stakingKeeper.EXPECT().
					GetDelegatorBonded(gomock.Any(), ValidAddress1).
					Return(sdkmath.NewInt(100), nil).
					Times(1)

				suite.bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, ValidAddress1, sdk.NewCoins(
						sdk.NewCoin("uband", sdkmath.NewInt(10)),
					)).
					Return(nil).
					Times(1)
			},
			postCheck: func() {
				stake := suite.restakeKeeper.GetStake(ctx, ValidAddress1)
				suite.Require().Equal(types.Stake{
					StakerAddress: ValidAddress1.String(),
					Coins: sdk.NewCoins(
						sdk.NewCoin("uband", sdkmath.NewInt(40)),
					),
				}, stake)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preCheck()
			_, err := suite.msgServer.Unstake(suite.ctx, tc.input)

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

func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	ctx := suite.ctx
	suite.setupState()

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
		preCheck  func()
		postCheck func()
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid authority",
				Params: types.Params{
					AllowedDenoms: []string{""},
				},
			},
			expErr:    true,
			expErrMsg: "invalid authority",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "invalid denom",
			input: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					AllowedDenoms: []string{""},
				},
			},
			expErr:    true,
			expErrMsg: "invalid denom",
			preCheck:  func() {},
			postCheck: func() {},
		},
		{
			name: "success",
			input: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params: types.Params{
					AllowedDenoms: []string{"ustBand"},
				},
			},
			expErr:    false,
			expErrMsg: "",
			preCheck:  func() {},
			postCheck: func() {
				params := suite.restakeKeeper.GetParams(ctx)
				suite.Require().Equal(types.Params{
					AllowedDenoms: []string{"ustBand"},
				}, params)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preCheck()
			_, err := suite.msgServer.UpdateParams(suite.ctx, tc.input)

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
