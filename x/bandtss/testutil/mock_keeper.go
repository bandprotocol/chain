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

	"github.com/bandprotocol/chain/v2/x/bandtss"
	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// NOTE: cannot put suite.Suite inside this struct, or else the test get timeout.
type TestSuite struct {
	Keeper *keeper.Keeper

	MockAccountKeeper *MockAccountKeeper
	MockBankKeeper    *MockBankKeeper
	MockDistrKeeper   *MockDistrKeeper
	MockStakingKeeper *MockStakingKeeper
	MockTSSKeeper     *MockTSSKeeper

	Ctx       sdk.Context
	Authority sdk.AccAddress
}

func NewTestSuite(t *testing.T) TestSuite {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(bandtss.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})

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
		paramtypes.Subspace{},
		accountKeeper,
		bankKeeper,
		distrKeeper,
		stakingKeeper,
		tssKeeper,
		authority.String(),
		authtypes.FeeCollectorName,
	)

	return TestSuite{
		Keeper:            bandtssKeeper,
		MockAccountKeeper: accountKeeper,
		MockBankKeeper:    bankKeeper,
		MockDistrKeeper:   distrKeeper,
		MockStakingKeeper: stakingKeeper,
		MockTSSKeeper:     tssKeeper,
		Ctx:               ctx,
		Authority:         authority,
	}
}
