package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/mock/gomock"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestMsgCreateTunnel() {
	signalDeviations := []types.SignalDeviation{
		{
			SignalID:         "BTC",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
		{
			SignalID:         "ETH",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}

	cases := map[string]struct {
		preRun    func() (*types.MsgCreateTunnel, error)
		expErr    bool
		expErrMsg string
	}{
		"max signal exceed": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				params := types.DefaultParams()
				params.MaxSignals = 1
				s.Require().NoError(s.keeper.SetParams(s.ctx, params))

				return types.NewMsgCreateTunnel(
					signalDeviations,
					10,
					route,
					types.ENCODER_FIXED_POINT_ABI,
					sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100))),
					sdk.AccAddress([]byte("creator_address")),
				)
			},
			expErr:    true,
			expErrMsg: "max signals exceeded",
		},
		"interval too low": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				params := types.DefaultParams()
				params.MinInterval = 5
				s.Require().NoError(s.keeper.SetParams(s.ctx, params))

				return types.NewMsgCreateTunnel(
					signalDeviations,
					1,
					route,
					types.ENCODER_FIXED_POINT_ABI,
					sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100))),
					sdk.AccAddress([]byte("creator_address")),
				)
			},
			expErr:    true,
			expErrMsg: "interval too low",
		},
		"all good without initial deposit": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				s.accountKeeper.EXPECT().
					GetAccount(s.ctx, gomock.Any()).
					Return(nil).Times(1)
				s.accountKeeper.EXPECT().NewAccount(s.ctx, gomock.Any()).Times(1)
				s.accountKeeper.EXPECT().SetAccount(s.ctx, gomock.Any()).Times(1)

				return types.NewMsgCreateTunnel(
					signalDeviations,
					10,
					route,
					types.ENCODER_FIXED_POINT_ABI,
					sdk.NewCoins(),
					sdk.AccAddress([]byte("creator_address")),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
		"all good": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				depositor := sdk.AccAddress([]byte("creator_address"))
				depositAmount := sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100)))

				s.accountKeeper.EXPECT().
					GetAccount(s.ctx, gomock.Any()).
					Return(nil).Times(1)
				s.accountKeeper.EXPECT().NewAccount(s.ctx, gomock.Any()).Times(1)
				s.accountKeeper.EXPECT().SetAccount(s.ctx, gomock.Any()).Times(1)
				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(s.ctx, depositor, types.ModuleName, depositAmount).
					Return(nil).Times(1)

				return types.NewMsgCreateTunnel(
					signalDeviations,
					10,
					route,
					types.ENCODER_FIXED_POINT_ABI,
					depositAmount,
					depositor,
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg, err := tc.preRun()
			s.Require().NoError(err)

			res, err := s.msgServer.CreateTunnel(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(res.TunnelID)
			}

			s.reset()
		})
	}
}

func (s *KeeperTestSuite) TestMsgEditTunnel() {
	cases := map[string]struct {
		preRun    func() *types.MsgEditTunnel
		expErr    bool
		expErrMsg string
	}{
		"max signal exceed": {
			preRun: func() *types.MsgEditTunnel {
				params := types.DefaultParams()
				params.MaxSignals = 1
				err := s.keeper.SetParams(s.ctx, params)
				s.Require().NoError(err)

				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "BTC",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
					{
						SignalID:         "ETH",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgEditTunnel(
					1,
					editedSignalDeviations,
					10,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "max signals exceeded",
		},
		"interval too low": {
			preRun: func() *types.MsgEditTunnel {
				params := types.DefaultParams()
				params.MinInterval = 5
				err := s.keeper.SetParams(s.ctx, params)
				s.Require().NoError(err)

				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "BTC",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgEditTunnel(
					1,
					editedSignalDeviations,
					1,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "interval too low",
		},
		"tunnel not found": {
			preRun: func() *types.MsgEditTunnel {
				return types.NewMsgEditTunnel(
					1,
					[]types.SignalDeviation{},
					10,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid creator of the tunnel": {
			preRun: func() *types.MsgEditTunnel {
				s.AddSampleTunnel(false)

				return types.NewMsgEditTunnel(
					1,
					[]types.SignalDeviation{},
					10,
					sdk.AccAddress([]byte("wrong_creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "invalid creator of the tunnel",
		},
		"all good": {
			preRun: func() *types.MsgEditTunnel {
				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "BTC",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgEditTunnel(
					1,
					editedSignalDeviations,
					10,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg := tc.preRun()

			_, err := s.msgServer.EditTunnel(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}

			s.reset()
		})
	}
}

func (s *KeeperTestSuite) TestMsgActivate() {
	cases := map[string]struct {
		preRun    func() *types.MsgActivate
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() *types.MsgActivate {
				return types.NewMsgActivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid creator of the tunnel": {
			preRun: func() *types.MsgActivate {
				s.AddSampleTunnel(false)

				return types.NewMsgActivate(1, sdk.AccAddress([]byte("wrong_creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "invalid creator of the tunnel",
		},
		"already active": {
			preRun: func() *types.MsgActivate {
				s.AddSampleTunnel(true)

				return types.NewMsgActivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "already active",
		},
		"insufficient deposit": {
			preRun: func() *types.MsgActivate {
				params := types.DefaultParams()
				params.MinDeposit = sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(1000)))
				s.Require().NoError(s.keeper.SetParams(s.ctx, params))

				s.AddSampleTunnel(false)

				return types.NewMsgActivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "insufficient deposit",
		},
		"all good": {
			preRun: func() *types.MsgActivate {
				params := types.DefaultParams()
				params.MinDeposit = sdk.NewCoins()
				s.Require().NoError(s.keeper.SetParams(s.ctx, params))

				s.AddSampleTunnel(false)

				return types.NewMsgActivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg := tc.preRun()

			_, err := s.msgServer.Activate(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}

			s.reset()
		})
	}
}

func (s *KeeperTestSuite) TestMsgDeactivate() {
	cases := map[string]struct {
		preRun    func() *types.MsgDeactivate
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() *types.MsgDeactivate {
				return types.NewMsgDeactivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid creator of the tunnel": {
			preRun: func() *types.MsgDeactivate {
				s.AddSampleTunnel(true)

				return types.NewMsgDeactivate(1, sdk.AccAddress([]byte("wrong_creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "invalid creator of the tunnel",
		},
		"already inactive": {
			preRun: func() *types.MsgDeactivate {
				s.AddSampleTunnel(false)

				return types.NewMsgDeactivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "already inactive",
		},
		"all good": {
			preRun: func() *types.MsgDeactivate {
				s.AddSampleTunnel(true)

				return types.NewMsgDeactivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg := tc.preRun()

			_, err := s.msgServer.Deactivate(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}

			s.reset()
		})
	}
}

func (s *KeeperTestSuite) TestMsgTriggerTunnel() {
	cases := map[string]struct {
		preRun    func() *types.MsgTriggerTunnel
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() *types.MsgTriggerTunnel {
				return types.NewMsgTriggerTunnel(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid creator of the tunnel": {
			preRun: func() *types.MsgTriggerTunnel {
				s.AddSampleTunnel(true)

				return types.NewMsgTriggerTunnel(1, sdk.AccAddress([]byte("wrong_creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "invalid creator of the tunnel",
		},
		"inactive tunnel": {
			preRun: func() *types.MsgTriggerTunnel {
				s.AddSampleTunnel(false)

				return types.NewMsgTriggerTunnel(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    true,
			expErrMsg: "inactive tunnel",
		},
		"all good": {
			preRun: func() *types.MsgTriggerTunnel {
				s.AddSampleTunnel(true)

				feePayer := sdk.MustAccAddressFromBech32(
					"cosmos1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62qh49enj",
				)

				s.feedsKeeper.EXPECT().GetCurrentPrices(gomock.Any()).Return([]feedstypes.Price{
					{PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
				})
				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), feePayer, types.ModuleName, types.DefaultBasePacketFee).
					Return(nil)

				return types.NewMsgTriggerTunnel(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg := tc.preRun()

			_, err := s.msgServer.TriggerTunnel(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}

			s.reset()
		})
	}
}

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	params := types.DefaultParams()

	cases := map[string]struct {
		preRun    func() *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		"invalid authority": {
			preRun: func() *types.MsgUpdateParams {
				return types.NewMsgUpdateParams(
					"invalid authority",
					params,
				)
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		"all good": {
			preRun: func() *types.MsgUpdateParams {
				return types.NewMsgUpdateParams(
					s.authority.String(),
					params,
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.preRun())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
