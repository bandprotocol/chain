package testutil

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel"
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// TestSuite is a struct that embeds a *testing.T and provides a setup for a mock keeper
type TestSuite struct {
	t           *testing.T
	Keeper      *keeper.Keeper
	QueryServer types.QueryServer

	MockAccountKeeper *MockAccountKeeper
	MockBankKeeper    *MockBankKeeper
	MockFeedsKeeper   *MockFeedsKeeper
	MockBandtssKeeper *MockBandtssKeeper

	Ctx       sdk.Context
	Authority sdk.AccAddress
}

// NewTestSuite returns a new TestSuite object
func NewTestSuite(t *testing.T) TestSuite {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(tunnel.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})

	accountKeeper := NewMockAccountKeeper(ctrl)
	bankKeeper := NewMockBankKeeper(ctrl)
	feedsKeeper := NewMockFeedsKeeper(ctrl)
	bandtssKeeper := NewMockBandtssKeeper(ctrl)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(authority).AnyTimes()
	tunnelKeeper := keeper.NewKeeper(
		encCfg.Codec.(codec.BinaryCodec),
		key,
		accountKeeper,
		bankKeeper,
		feedsKeeper,
		bandtssKeeper,
		authority.String(),
	)
	queryServer := keeper.NewQueryServer(tunnelKeeper)

	return TestSuite{
		t:                 t,
		Keeper:            tunnelKeeper,
		QueryServer:       queryServer,
		MockAccountKeeper: accountKeeper,
		MockBankKeeper:    bankKeeper,
		MockFeedsKeeper:   feedsKeeper,
		MockBandtssKeeper: bandtssKeeper,
		Ctx:               ctx,
		Authority:         authority,
	}
}

func (s *TestSuite) T() *testing.T {
	return s.t
}
