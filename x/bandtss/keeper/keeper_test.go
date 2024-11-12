package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/bandtss"
	"github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	bandtsstestutil "github.com/bandprotocol/chain/v3/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
	tsstestutils "github.com/bandprotocol/chain/v3/x/tss/testutil"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

type AppTestSuite struct {
	suite.Suite

	app         *band.BandApp
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
	dir := sdktestutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)
	s.ctx = s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(s.app.BandtssKeeper))
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient

	s.msgSrvr = keeper.NewMsgServerImpl(s.app.BandtssKeeper)
	s.tssMsgSrvr = tsskeeper.NewMsgServerImpl(s.app.TSSKeeper)
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
	if _, err := s.app.EndBlocker(s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)); err != nil {
		return err
	}

	if transition.IncomingGroupID != s.app.BandtssKeeper.GetCurrentGroup(s.ctx).GroupID {
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

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

// KeeperTestSuite is a struct that embeds a *testing.T and provides a setup for a mock keeper
type KeeperTestSuite struct {
	suite.Suite

	key  *storetypes.KVStoreKey
	tkey *storetypes.TransientStoreKey

	keeper      keeper.Keeper
	queryServer types.QueryServer
	tssCallback *keeper.TSSCallback

	accountKeeper *bandtsstestutil.MockAccountKeeper
	bankKeeper    *bandtsstestutil.MockBankKeeper
	distrKeeper   *bandtsstestutil.MockDistrKeeper
	tssKeeper     *bandtsstestutil.MockTSSKeeper

	moduleAcc sdk.ModuleAccountI
	ctx       sdk.Context
	authority sdk.AccAddress
}

// SetupTest initializes the mock keeper and the context
func (s *KeeperTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.key = storetypes.NewKVStoreKey(types.StoreKey)
	s.tkey = storetypes.NewTransientStoreKey("transient_test")

	testCtx := sdktestutil.DefaultContextWithDB(s.T(), s.key, s.tkey)
	encCfg := moduletestutil.MakeTestEncodingConfig(bandtss.AppModuleBasic{})
	s.ctx = testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)})

	s.accountKeeper = bandtsstestutil.NewMockAccountKeeper(ctrl)
	s.bankKeeper = bandtsstestutil.NewMockBankKeeper(ctrl)
	s.distrKeeper = bandtsstestutil.NewMockDistrKeeper(ctrl)
	s.tssKeeper = bandtsstestutil.NewMockTSSKeeper(ctrl)

	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	s.accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(s.authority).AnyTimes()
	s.keeper = keeper.NewKeeper(
		encCfg.Codec.(codec.BinaryCodec),
		s.key,
		s.accountKeeper,
		s.bankKeeper,
		s.distrKeeper,
		s.tssKeeper,
		s.authority.String(),
		authtypes.FeeCollectorName,
	)

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	tssCallback := keeper.NewTSSCallback(s.keeper)
	s.tssCallback = &tssCallback
	s.queryServer = keeper.NewQueryServer(s.keeper)

	s.moduleAcc = authtypes.NewModuleAccount(&authtypes.BaseAccount{}, types.ModuleName, "auth")
}

func (s *KeeperTestSuite) SetupSubTest() {
	// clear the context state and set params
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), s.key, s.tkey)
	s.ctx = testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)})

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
