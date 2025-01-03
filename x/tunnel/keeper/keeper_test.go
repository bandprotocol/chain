package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	band "github.com/bandprotocol/chain/v3/app"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

type KeeperTestSuite struct {
	suite.Suite

	keeper      keeper.Keeper
	queryServer types.QueryServer
	msgServer   types.MsgServer
	storeKey    storetypes.StoreKey

	accountKeeper  *testutil.MockAccountKeeper
	bankKeeper     *testutil.MockBankKeeper
	feedsKeeper    *testutil.MockFeedsKeeper
	bandtssKeeper  *testutil.MockBandtssKeeper
	channelKeeper  *testutil.MockChannelKeeper
	icsWrapper     *testutil.MockICS4Wrapper
	portKeeper     *testutil.MockPortKeeper
	scopedKeeper   *testutil.MockScopedKeeper
	transferKeeper *testutil.MockTransferKeeper

	ctx       sdk.Context
	authority sdk.AccAddress
}

func (s *KeeperTestSuite) SetupTest() {
	s.reset()
}

func (s *KeeperTestSuite) reset() {
	ctrl := gomock.NewController(s.T())
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(tunnel.AppModuleBasic{})

	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	feedsKeeper := testutil.NewMockFeedsKeeper(ctrl)
	bandtssKeeper := testutil.NewMockBandtssKeeper(ctrl)
	channelKeeper := testutil.NewMockChannelKeeper(ctrl)
	icsWrapper := testutil.NewMockICS4Wrapper(ctrl)
	portKeeper := testutil.NewMockPortKeeper(ctrl)
	scopedKeeper := testutil.NewMockScopedKeeper(ctrl)
	transferKeeper := testutil.NewMockTransferKeeper(ctrl)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(authority).AnyTimes()

	s.storeKey = key
	s.keeper = keeper.NewKeeper(
		encCfg.Codec.(codec.BinaryCodec),
		key,
		accountKeeper,
		bankKeeper,
		feedsKeeper,
		bandtssKeeper,
		channelKeeper,
		icsWrapper,
		portKeeper,
		scopedKeeper,
		transferKeeper,
		authority.String(),
	)
	s.queryServer = keeper.NewQueryServer(s.keeper)
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.accountKeeper = accountKeeper
	s.bankKeeper = bankKeeper
	s.feedsKeeper = feedsKeeper
	s.bandtssKeeper = bandtssKeeper
	s.channelKeeper = channelKeeper
	s.icsWrapper = icsWrapper
	s.portKeeper = portKeeper
	s.scopedKeeper = scopedKeeper
	s.transferKeeper = transferKeeper

	s.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})
	s.authority = authority

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) AddSampleTunnel(isActive bool) *types.Tunnel {
	ctx, k := s.ctx, s.keeper

	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	signalDeviations := []types.SignalDeviation{
		{
			SignalID:         "CS:BAND-USD",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
		Encoder:                    feedstypes.ENCODER_FIXED_POINT_ABI,
	}

	tunnel, err := k.AddTunnel(
		ctx,
		route,
		signalDeviations,
		10,
		sdk.AccAddress([]byte("creator_address")),
	)
	s.Require().NoError(err)

	if isActive {
		tunnel, err := k.GetTunnel(ctx, tunnel.ID)
		s.Require().NoError(err)

		// set deposit to the tunnel to be able to activate
		tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)
		k.SetTunnel(ctx, tunnel)

		err = k.ActivateTunnel(ctx, tunnel.ID)
		s.Require().NoError(err)
	}

	return tunnel
}

func (s *KeeperTestSuite) AddSampleIBCTunnel(isActive bool) {
	ctx, k := s.ctx, s.keeper

	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	signalDeviations := []types.SignalDeviation{
		{
			SignalID:         "CS:BAND-USD",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
	}

	route := &types.IBCRoute{
		ChannelID: "",
	}

	tunnel, err := k.AddTunnel(
		ctx,
		route,
		signalDeviations,
		10,
		sdk.AccAddress([]byte("creator_address")),
	)
	s.Require().NoError(err)

	if isActive {
		tunnel, err := k.GetTunnel(ctx, tunnel.ID)
		s.Require().NoError(err)

		// set deposit to the tunnel to be able to activate
		tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)
		k.SetTunnel(ctx, tunnel)

		err = k.ActivateTunnel(ctx, tunnel.ID)
		s.Require().NoError(err)
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
