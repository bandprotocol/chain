package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel"
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper      *keeper.Keeper
	queryServer types.QueryServer
	msgServer   types.MsgServer

	accountKeeper *testutil.MockAccountKeeper
	bankKeeper    *testutil.MockBankKeeper
	feedsKeeper   *testutil.MockFeedsKeeper
	bandtssKeeper *testutil.MockBandtssKeeper

	ctx       sdk.Context
	authority sdk.AccAddress
}

func (s *KeeperTestSuite) SetupTest() {
	s.reset()
}

func (s *KeeperTestSuite) reset() {
	ctrl := gomock.NewController(s.T())
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(tunnel.AppModuleBasic{})

	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	feedsKeeper := testutil.NewMockFeedsKeeper(ctrl)
	bandtssKeeper := testutil.NewMockBandtssKeeper(ctrl)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(authority).AnyTimes()

	s.keeper = keeper.NewKeeper(
		encCfg.Codec.(codec.BinaryCodec),
		key,
		accountKeeper,
		bankKeeper,
		feedsKeeper,
		bandtssKeeper,
		authority.String(),
	)
	s.queryServer = keeper.NewQueryServer(s.keeper)
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.accountKeeper = accountKeeper
	s.bankKeeper = bankKeeper
	s.feedsKeeper = feedsKeeper
	s.bandtssKeeper = bandtssKeeper
	s.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})
	s.authority = authority

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) AddSampleTunnel(isActive bool) {
	s.accountKeeper.EXPECT().
		GetAccount(s.ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(s.ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(s.ctx, gomock.Any()).Times(1)

	signalDeviations := []types.SignalDeviation{
		{
			SignalID:         "BTC",
			SoftDeviationBPS: 100,
			HardDeviationBPS: 100,
		},
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}
	routeAny, err := codectypes.NewAnyWithValue(route)
	s.Require().NoError(err)

	tunnel, err := s.keeper.AddTunnel(
		s.ctx,
		routeAny,
		types.ENCODER_FIXED_POINT_ABI,
		signalDeviations,
		10,
		sdk.AccAddress([]byte("creator_address")),
	)
	s.Require().NoError(err)

	if isActive {
		err := s.keeper.ActivateTunnel(s.ctx, tunnel.ID)
		s.Require().NoError(err)
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
