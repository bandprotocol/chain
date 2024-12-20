package keeper_test

import (
	"go.uber.org/mock/gomock"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestMsgCreateTunnel() {
	signalDeviations := []types.SignalDeviation{
		{
			SignalID:         "CS:BAND-USD",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
		{
			SignalID:         "CS:ETH-USD",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
		Encoder:                    feedstypes.ENCODER_FIXED_POINT_ABI,
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
					60,
					route,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "max signals exceeded",
		},
		"deviation out of range": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				params := types.DefaultParams()
				params.MinDeviationBPS = 1000
				params.MaxDeviationBPS = 10000
				s.Require().NoError(s.keeper.SetParams(s.ctx, params))

				return types.NewMsgCreateTunnel(
					signalDeviations,
					60,
					route,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "deviation out of range",
		},
		"interval out of range": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				params := types.DefaultParams()
				params.MinInterval = 5
				s.Require().NoError(s.keeper.SetParams(s.ctx, params))

				return types.NewMsgCreateTunnel(
					signalDeviations,
					1,
					route,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "interval out of range",
		},
		"channel id should be set after create tunnel": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				depositor := sdk.AccAddress([]byte("creator_address"))
				depositAmount := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))

				return types.NewMsgCreateTunnel(
					signalDeviations,
					60,
					types.NewIBCRoute("channel-0"),
					depositAmount,
					depositor.String(),
				)
			},
			expErr:    true,
			expErrMsg: "channel id should be set after create tunnel",
		},
		"all good (ibc route)": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				s.accountKeeper.EXPECT().
					GetAccount(s.ctx, gomock.Any()).
					Return(nil).Times(1)
				s.accountKeeper.EXPECT().NewAccount(s.ctx, gomock.Any()).Times(1)
				s.accountKeeper.EXPECT().SetAccount(s.ctx, gomock.Any()).Times(1)
				s.scopedKeeper.EXPECT().
					GetCapability(s.ctx, "ports/tunnel.1").
					Return(&capabilitytypes.Capability{}, true)

				return types.NewMsgCreateTunnel(
					signalDeviations,
					60,
					types.NewIBCRoute(""),
					sdk.NewCoins(),
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
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
					60,
					route,
					sdk.NewCoins(),
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
		"all good": {
			preRun: func() (*types.MsgCreateTunnel, error) {
				depositor := sdk.AccAddress([]byte("creator_address"))
				depositAmount := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))

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
					60,
					route,
					depositAmount,
					depositor.String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
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
		})
	}
}

func (s *KeeperTestSuite) TestMsgUpdateRoute() {
	cases := map[string]struct {
		preRun    func() (*types.MsgUpdateRoute, error)
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() (*types.MsgUpdateRoute, error) {
				return types.NewMsgUpdateIBCRoute(
					1,
					"channel-0",
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"mismatch route type": {
			preRun: func() (*types.MsgUpdateRoute, error) {
				s.AddSampleTunnel(false)

				return types.NewMsgUpdateIBCRoute(
					1,
					"channel-0",
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "cannot change route type",
		},
		"invalid creator of the tunnel": {
			preRun: func() (*types.MsgUpdateRoute, error) {
				s.AddSampleTunnel(false)

				return types.NewMsgUpdateIBCRoute(
					1,
					"channel-0",
					sdk.AccAddress([]byte("wrong_creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "invalid creator of the tunnel",
		},
		"all good": {
			preRun: func() (*types.MsgUpdateRoute, error) {
				s.channelKeeper.EXPECT().
					GetChannel(gomock.Any(), "tunnel.1", "channel-0").
					Return(channeltypes.Channel{}, true)

				s.AddSampleIBCTunnel(false)

				return types.NewMsgUpdateIBCRoute(
					1,
					"channel-0",
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg, err := tc.preRun()
			s.Require().NoError(err)

			_, err = s.msgServer.UpdateRoute(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgUpdateSignalsAndInterval() {
	cases := map[string]struct {
		preRun    func() *types.MsgUpdateSignalsAndInterval
		expErr    bool
		expErrMsg string
	}{
		"max signal exceed": {
			preRun: func() *types.MsgUpdateSignalsAndInterval {
				params := types.DefaultParams()
				params.MaxSignals = 1
				err := s.keeper.SetParams(s.ctx, params)
				s.Require().NoError(err)

				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "CS:BAND-USD",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
					{
						SignalID:         "CS:ETH-USD",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgUpdateSignalsAndInterval(
					1,
					editedSignalDeviations,
					60,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "max signals exceeded",
		},
		"deviation out of range": {
			preRun: func() *types.MsgUpdateSignalsAndInterval {
				params := types.DefaultParams()
				params.MinDeviationBPS = 1000
				params.MaxDeviationBPS = 10000
				err := s.keeper.SetParams(s.ctx, params)
				s.Require().NoError(err)

				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "CS:BAND-USD",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgUpdateSignalsAndInterval(
					1,
					editedSignalDeviations,
					60,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "deviation out of range",
		},
		"interval out of range": {
			preRun: func() *types.MsgUpdateSignalsAndInterval {
				params := types.DefaultParams()
				params.MinInterval = 5
				err := s.keeper.SetParams(s.ctx, params)
				s.Require().NoError(err)

				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "CS:BAND-USD",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgUpdateSignalsAndInterval(
					1,
					editedSignalDeviations,
					1,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "interval out of range",
		},
		"tunnel not found": {
			preRun: func() *types.MsgUpdateSignalsAndInterval {
				return types.NewMsgUpdateSignalsAndInterval(
					1,
					[]types.SignalDeviation{},
					60,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid creator of the tunnel": {
			preRun: func() *types.MsgUpdateSignalsAndInterval {
				s.AddSampleTunnel(false)

				return types.NewMsgUpdateSignalsAndInterval(
					1,
					[]types.SignalDeviation{},
					60,
					sdk.AccAddress([]byte("wrong_creator_address")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "invalid creator of the tunnel",
		},
		"all good": {
			preRun: func() *types.MsgUpdateSignalsAndInterval {
				s.AddSampleTunnel(false)

				editedSignalDeviations := []types.SignalDeviation{
					{
						SignalID:         "CS:BAND-USD",
						SoftDeviationBPS: 200,
						HardDeviationBPS: 200,
					},
				}

				return types.NewMsgUpdateSignalsAndInterval(
					1,
					editedSignalDeviations,
					60,
					sdk.AccAddress([]byte("creator_address")).String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.UpdateSignalsAndInterval(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestWithdrawFeePayerFunds() {
	creator := sdk.AccAddress([]byte("creator_address")).String()

	cases := map[string]struct {
		preRun    func() *types.MsgWithdrawFeePayerFunds
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() *types.MsgWithdrawFeePayerFunds {
				return &types.MsgWithdrawFeePayerFunds{
					Creator:  creator,
					TunnelID: 1,
					Amount:   sdk.NewCoins(sdk.NewInt64Coin("uband", 500)),
				}
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid creator": {
			preRun: func() *types.MsgWithdrawFeePayerFunds {
				s.AddSampleTunnel(false)

				return &types.MsgWithdrawFeePayerFunds{
					Creator:  "invalid_creator",
					TunnelID: 1,
					Amount:   sdk.NewCoins(sdk.NewInt64Coin("uband", 500)),
				}
			},
			expErr:    true,
			expErrMsg: "creator invalid_creator, tunnelID 1",
		},
		"insufficient funds": {
			preRun: func() *types.MsgWithdrawFeePayerFunds {
				tunnel := s.AddSampleTunnel(false)

				amount := sdk.NewCoins(sdk.NewInt64Coin("uband", 9999999999))

				s.bankKeeper.EXPECT().
					SendCoins(s.ctx, sdk.MustAccAddressFromBech32(tunnel.FeePayer), sdk.MustAccAddressFromBech32(creator), amount).
					Return(sdkerrors.ErrInsufficientFunds).Times(1)

				return &types.MsgWithdrawFeePayerFunds{
					Creator:  creator,
					TunnelID: tunnel.ID,
					Amount:   amount,
				}
			},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
		"all good": {
			preRun: func() *types.MsgWithdrawFeePayerFunds {
				tunnel := s.AddSampleTunnel(false)

				amount := sdk.NewCoins(sdk.NewInt64Coin("uband", 500))

				s.bankKeeper.EXPECT().
					SendCoins(s.ctx, sdk.MustAccAddressFromBech32(tunnel.FeePayer), sdk.MustAccAddressFromBech32(creator), amount).
					Return(nil).Times(1)

				return &types.MsgWithdrawFeePayerFunds{
					Creator:  creator,
					TunnelID: 1,
					Amount:   amount,
				}
			},
			expErr: false,
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.WithdrawFeePayerFunds(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
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
				params.MinDeposit = sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1000)))
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

				s.bandtssKeeper.EXPECT().IsReady(gomock.Any()).Return(true)

				return types.NewMsgActivate(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.Activate(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
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
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.Deactivate(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
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

				latestTunnelID := s.keeper.GetTunnelCount(s.ctx)
				tunnel, err := s.keeper.GetTunnel(s.ctx, latestTunnelID)
				feePayer := sdk.MustAccAddressFromBech32(tunnel.FeePayer)
				s.Require().NoError(err)

				s.bandtssKeeper.EXPECT().IsReady(gomock.Any()).Return(true)
				s.bandtssKeeper.EXPECT().GetSigningFee(gomock.Any()).Return(
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), nil,
				).Times(2)

				s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
					gomock.Any(),
					uint64(1),
					"chain-1",
					"0x1234567890abcdef",
					gomock.Any(),
					feePayer,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
				).Return(bandtsstypes.SigningID(1), nil)

				s.feedsKeeper.EXPECT().
					GetPrices(gomock.Any(), []string{"CS:BAND-USD"}).
					Return([]feedstypes.Price{
						{
							Status:    feedstypes.PRICE_STATUS_AVAILABLE,
							SignalID:  "CS:BAND-USD",
							Price:     50000,
							Timestamp: 0,
						},
					})
				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), feePayer, types.ModuleName, types.DefaultBasePacketFee).
					Return(nil)

				spendableCoins := types.DefaultBasePacketFee.Add(sdk.NewCoin("uband", sdkmath.NewInt(20)))
				s.bankKeeper.EXPECT().SpendableCoins(gomock.Any(), feePayer).Return(spendableCoins)

				return types.NewMsgTriggerTunnel(1, sdk.AccAddress([]byte("creator_address")).String())
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.TriggerTunnel(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgDepositToTunnel() {
	cases := map[string]struct {
		preRun    func() *types.MsgDepositToTunnel
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() *types.MsgDepositToTunnel {
				return types.NewMsgDepositToTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("depositor")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"invalid deposit denom": {
			preRun: func() *types.MsgDepositToTunnel {
				s.AddSampleTunnel(true)

				return types.NewMsgDepositToTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("invalid_denom", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("depositor")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "invalid deposit denom",
		},
		"insufficient fund": {
			preRun: func() *types.MsgDepositToTunnel {
				s.AddSampleTunnel(true)

				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), types.ModuleName, gomock.Any()).
					Return(sdkerrors.ErrInsufficientFunds)

				return types.NewMsgDepositToTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("depositor")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "insufficient fund",
		},
		"all good": {
			preRun: func() *types.MsgDepositToTunnel {
				s.AddSampleTunnel(true)

				s.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), types.ModuleName, gomock.Any()).
					Return(nil)

				return types.NewMsgDepositToTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("depositor")).String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.DepositToTunnel(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgWithdrawFromTunnel() {
	cases := map[string]struct {
		preRun    func() *types.MsgWithdrawFromTunnel
		expErr    bool
		expErrMsg string
	}{
		"tunnel not found": {
			preRun: func() *types.MsgWithdrawFromTunnel {
				return types.NewMsgWithdrawFromTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("depositor")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "tunnel not found",
		},
		"deposit not found": {
			preRun: func() *types.MsgWithdrawFromTunnel {
				s.AddSampleTunnel(true)

				return types.NewMsgWithdrawFromTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					sdk.AccAddress([]byte("depositor")).String(),
				)
			},
			expErr:    true,
			expErrMsg: "deposit not found",
		},
		"insufficient deposit": {
			preRun: func() *types.MsgWithdrawFromTunnel {
				s.AddSampleTunnel(true)

				depositor := sdk.AccAddress([]byte("depositor"))
				deposit := types.Deposit{
					TunnelID:  1,
					Depositor: depositor.String(),
					Amount:    sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
				}

				s.keeper.SetDeposit(
					s.ctx,
					deposit,
				)

				return types.NewMsgWithdrawFromTunnel(
					1,
					sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1000))),
					depositor.String(),
				)
			},
			expErr:    true,
			expErrMsg: "insufficient deposit",
		},
		"all good": {
			preRun: func() *types.MsgWithdrawFromTunnel {
				s.AddSampleTunnel(true)

				amount := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))

				depositor := sdk.AccAddress([]byte("depositor"))
				deposit := types.Deposit{
					TunnelID:  1,
					Depositor: depositor.String(),
					Amount:    amount,
				}

				s.keeper.SetDeposit(
					s.ctx,
					deposit,
				)

				s.bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, depositor, amount).
					Return(nil)

				return types.NewMsgWithdrawFromTunnel(
					1,
					amount,
					depositor.String(),
				)
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			s.reset()
			msg := tc.preRun()

			_, err := s.msgServer.WithdrawFromTunnel(s.ctx, msg)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
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
