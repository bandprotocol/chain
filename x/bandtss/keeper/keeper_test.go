package keeper_test

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss"
	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	bandtsstestutil "github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	tsstestutils "github.com/bandprotocol/chain/v2/x/tss/testutil"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type AppTestSuite struct {
	suite.Suite

	app         *bandtesting.TestingApp
	ctx         sdk.Context
	queryClient types.QueryClient
	msgSrvr     types.MsgServer
	tssMsgSrvr  tsstypes.MsgServer
	authority   sdk.AccAddress
}

var (
	PrivD = testutil.HexDecode("de6aedbe8ba688dd6d342881eb1e67c3476e825106477360148e2858a5eb565c")
	PrivE = testutil.HexDecode("3ff4fb2beac0cee0ab230829a5ae0881310046282e79c978ca22f44897ea434a")
	PubD  = tss.Scalar(PrivD).Point()
	PubE  = tss.Scalar(PrivE).Point()
)

func (s *AppTestSuite) SetupTest() {
	app, ctx := bandtesting.CreateTestApp(s.T(), true)
	s.app = app
	s.ctx = ctx.WithBlockTime(time.Now().UTC())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.BandtssKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	s.queryClient = queryClient
	s.msgSrvr = keeper.NewMsgServerImpl(app.BandtssKeeper)
	s.tssMsgSrvr = tsskeeper.NewMsgServerImpl(app.TSSKeeper)
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
}

func (s *AppTestSuite) CreateNewGroup(
	n uint64,
	threshold uint64,
	execTime time.Time,
) (*tsstestutils.GroupContext, error) {
	accounts := tsstestutils.GenerateAccounts(n)
	members := make([]sdk.AccAddress, n)
	memberStrs := make([]string, n)
	for i := 0; i < len(accounts); i++ {
		members[i] = accounts[i].Address
		memberStrs[i] = members[i].String()
	}

	des := make([][]tsstestutils.DEWithPrivateNonce, n)
	for i := 0; i < len(des); i++ {
		des[i] = make([]tsstestutils.DEWithPrivateNonce, 0, 10)
	}

	secrets := make([]tss.Scalar, n)
	for i := 0; i < len(secrets); i++ {
		secret, err := tss.RandomScalar()
		if err != nil {
			return nil, err
		}
		secrets[i] = secret
	}

	if _, err := s.msgSrvr.TransitionGroup(s.ctx, types.NewMsgTransitionGroup(
		memberStrs, threshold, execTime, s.authority.String(),
	)); err != nil {
		return nil, err
	}

	groupID := tss.GroupID(s.app.TSSKeeper.GetGroupCount(s.ctx))
	groupCtx := &tsstestutils.GroupContext{
		GroupID:  groupID,
		Accounts: accounts,
		DEs:      des,
		Secrets:  secrets,
	}

	if err := groupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper); err != nil {
		return nil, err
	}

	if err := groupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper); err != nil {
		return nil, err
	}

	if err := groupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper); err != nil {
		return nil, err
	}

	if err := groupCtx.GenerateDE(s.ctx, s.app.TSSKeeper); err != nil {
		return nil, err
	}

	return groupCtx, nil
}

func (s *AppTestSuite) ExecuteReplaceGroup() error {
	transition, found := s.app.BandtssKeeper.GetGroupTransition(s.ctx)
	if !found {
		return fmt.Errorf("group transition not found")
	}

	if transition.Status != types.TRANSITION_STATUS_WAITING_EXECUTION {
		return fmt.Errorf("unexpected transition status: %s", transition.Status.String())
	}

	s.ctx = s.ctx.WithBlockTime(transition.ExecTime)
	s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight() + 1})

	if transition.IncomingGroupID != s.app.BandtssKeeper.GetCurrentGroupID(s.ctx) {
		return fmt.Errorf("unexpected current group id: %d", transition.IncomingGroupID)
	}
	return nil
}

func (s *AppTestSuite) SignTransition(groupCtx *tsstestutils.GroupContext) error {
	transition, found := s.app.BandtssKeeper.GetGroupTransition(s.ctx)
	if !found {
		return fmt.Errorf("group transition not found")
	}
	if transition.Status != types.TRANSITION_STATUS_WAITING_SIGN {
		return fmt.Errorf("unexpected transition status: %s", transition.Status.String())
	}

	err := groupCtx.SubmitSignature(s.ctx, s.app.TSSKeeper, s.tssMsgSrvr, transition.SigningID)
	if err != nil {
		return err
	}

	// Execute the EndBlocker to process signings
	s.app.TSSKeeper.HandleSigningEndBlock(s.ctx)
	transition, found = s.app.BandtssKeeper.GetGroupTransition(s.ctx)
	if !found {
		return fmt.Errorf("group transition not found")
	}
	if transition.Status != types.TRANSITION_STATUS_WAITING_EXECUTION {
		return fmt.Errorf("unexpected transition status: %s", transition.Status.String())
	}
	return nil
}

func (s *AppTestSuite) SetupNewGroup(n uint64, threshold uint64) *tsstestutils.GroupContext {
	execTime := s.ctx.BlockTime().Add(10 * time.Minute)
	groupCtx, err := s.CreateNewGroup(n, threshold, execTime)
	s.Require().NoError(err)

	transition, found := s.app.BandtssKeeper.GetGroupTransition(s.ctx)
	if found && transition.Status == types.TRANSITION_STATUS_WAITING_SIGN {
		err := s.SignTransition(groupCtx)
		s.Require().NoError(err)
	}

	err = s.ExecuteReplaceGroup()
	s.Require().NoError(err)

	return groupCtx
}

func (s *AppTestSuite) TestParams() {
	k := s.app.BandtssKeeper

	testCases := []struct {
		name         string
		input        types.Params
		expectErr    bool
		expectErrStr string
	}{
		{
			name: "set invalid params",
			input: types.Params{
				ActiveDuration:          time.Duration(0),
				InactivePenaltyDuration: time.Duration(0),
				RewardPercentage:        0,
				Fee:                     sdk.NewCoins(),
			},
			expectErr:    true,
			expectErrStr: "must be positive:",
		},
		{
			name: "set full valid params",
			input: types.Params{
				ActiveDuration:          types.DefaultActiveDuration,
				RewardPercentage:        types.DefaultRewardPercentage,
				InactivePenaltyDuration: types.DefaultInactivePenaltyDuration,
				MaxTransitionDuration:   types.DefaultMaxTransitionDuration,
				Fee:                     types.DefaultFee,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expected := k.GetParams(s.ctx)
			err := k.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErrStr)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := k.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}

func (s *AppTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	expTime := s.ctx.BlockTime().Add(time.Hour)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.GetGrantMsgTypes() {
		err := s.app.AuthzKeeper.SaveGrant(ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
		s.Require().NoError(err)
	}

	isGrantee := k.CheckIsGrantee(ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

// KeeperTestSuite is a struct that embeds a *testing.T and provides a setup for a mock keeper
type KeeperTestSuite struct {
	t           *testing.T
	Keeper      *keeper.Keeper
	QueryServer types.QueryServer
	TssCallback *keeper.TSSCallback

	MockAccountKeeper *bandtsstestutil.MockAccountKeeper
	MockBankKeeper    *bandtsstestutil.MockBankKeeper
	MockDistrKeeper   *bandtsstestutil.MockDistrKeeper
	MockStakingKeeper *bandtsstestutil.MockStakingKeeper
	MockTSSKeeper     *bandtsstestutil.MockTSSKeeper

	ModuleAcc authtypes.ModuleAccountI
	Ctx       sdk.Context
	Authority sdk.AccAddress
}

// NewKeeperTestSuite returns a new KeeperTestSuite object
func NewKeeperTestSuite(t *testing.T) KeeperTestSuite {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(bandtss.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})

	authzKeeper := bandtsstestutil.NewMockAuthzKeeper(ctrl)
	accountKeeper := bandtsstestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := bandtsstestutil.NewMockBankKeeper(ctrl)
	distrKeeper := bandtsstestutil.NewMockDistrKeeper(ctrl)
	stakingKeeper := bandtsstestutil.NewMockStakingKeeper(ctrl)
	tssKeeper := bandtsstestutil.NewMockTSSKeeper(ctrl)

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
	err := bandtssKeeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	tssCallback := keeper.NewTSSCallback(bandtssKeeper)
	queryServer := keeper.NewQueryServer(bandtssKeeper)

	mAcc := authtypes.NewModuleAccount(&authtypes.BaseAccount{}, types.ModuleName, "auth")

	return KeeperTestSuite{
		Keeper:            bandtssKeeper,
		MockAccountKeeper: accountKeeper,
		MockBankKeeper:    bankKeeper,
		MockDistrKeeper:   distrKeeper,
		MockStakingKeeper: stakingKeeper,
		MockTSSKeeper:     tssKeeper,
		Ctx:               ctx,
		Authority:         authority,
		QueryServer:       queryServer,
		TssCallback:       &tssCallback,
		t:                 t,
		ModuleAcc:         mAcc,
	}
}

func (s *KeeperTestSuite) T() *testing.T {
	return s.t
}
