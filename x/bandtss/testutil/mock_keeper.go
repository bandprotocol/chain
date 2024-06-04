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

	"github.com/bandprotocol/chain/v2/x/bandtss"
	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// TestSuite is a struct that embeds a *testing.T and provides a setup for a mock keeper
type TestSuite struct {
	t           *testing.T
	Keeper      *keeper.Keeper
	QueryServer types.QueryServer
	Hook        keeper.Hooks

	MockAccountKeeper *MockAccountKeeper
	MockBankKeeper    *MockBankKeeper
	MockDistrKeeper   *MockDistrKeeper
	MockStakingKeeper *MockStakingKeeper
	MockTSSKeeper     *MockTSSKeeper

	Ctx       sdk.Context
	Authority sdk.AccAddress
}

// NewTestSuite returns a new TestSuite object
func NewTestSuite(t *testing.T) TestSuite {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(bandtss.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})

	authzKeeper := NewMockAuthzKeeper(ctrl)
	accountKeeper := NewMockAccountKeeper(ctrl)
	bankKeeper := NewMockBankKeeper(ctrl)
	distrKeeper := NewMockDistrKeeper(ctrl)
	stakingKeeper := NewMockStakingKeeper(ctrl)
	tssKeeper := NewMockTSSKeeper(ctrl)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(authority).AnyTimes()
	bandtssKeeper := keeper.NewKeeper(
		encCfg.Codec.(codec.BinaryCodec),
		key,
		authzKeeper,
		accountKeeper,
		bankKeeper,
		distrKeeper,
		stakingKeeper,
		tssKeeper,
		authority.String(),
		authtypes.FeeCollectorName,
	)

	queryServer := keeper.NewQueryServer(bandtssKeeper)

	return TestSuite{
		Keeper:            bandtssKeeper,
		MockAccountKeeper: accountKeeper,
		MockBankKeeper:    bankKeeper,
		MockDistrKeeper:   distrKeeper,
		MockStakingKeeper: stakingKeeper,
		MockTSSKeeper:     tssKeeper,
		Ctx:               ctx,
		Authority:         authority,
		QueryServer:       queryServer,
		Hook:              bandtssKeeper.Hooks(),
		t:                 t,
	}
}

func (s *TestSuite) T() *testing.T {
	return s.t
}
