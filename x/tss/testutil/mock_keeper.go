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
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tss"
	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// TestSuite is a struct that embeds a *testing.T and provides a setup for a mock keeper
type TestSuite struct {
	t           *testing.T
	Keeper      *keeper.Keeper
	QueryServer types.QueryServer
	Hook        types.TSSHooks

	MockAuthzKeeper       *MockAuthzKeeper
	MockRollingseedKeeper *MockRollingseedKeeper

	Authority sdk.AccAddress
	Ctx       sdk.Context
}

// NewTestSuite returns a new TestSuite object
func NewTestSuite(t *testing.T) TestSuite {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(tss.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})

	authzKeeper := NewMockAuthzKeeper(ctrl)
	rollingseedKeeper := NewMockRollingseedKeeper(ctrl)
	tssRouter := types.NewRouter()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	tssKeeper := keeper.NewKeeper(
		encCfg.Codec.(codec.BinaryCodec),
		key,
		paramtypes.Subspace{},
		authzKeeper,
		rollingseedKeeper,
		tssRouter,
		authority.String(),
	)

	queryServer := keeper.NewQueryServer(tssKeeper)

	return TestSuite{
		Keeper:                tssKeeper,
		MockAuthzKeeper:       authzKeeper,
		MockRollingseedKeeper: rollingseedKeeper,
		Ctx:                   ctx,
		Authority:             authority,
		QueryServer:           queryServer,
		Hook:                  tssKeeper.Hooks(),
		t:                     t,
	}
}

func (s *TestSuite) T() *testing.T {
	return s.t
}
