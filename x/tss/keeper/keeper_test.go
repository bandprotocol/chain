package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	tssapp "github.com/bandprotocol/chain/v3/x/tss"
	"github.com/bandprotocol/chain/v3/x/tss/keeper"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
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
	app := bandtesting.SetupWithCustomHome(false, dir)

	s.app = app
	s.ctx = s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.TSSKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	s.queryClient = queryClient
	s.msgSrvr = keeper.NewMsgServerImpl(app.TSSKeeper)
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
}

func (s *AppTestSuite) setupCreateGroup() {
	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Initialize members
		var members []sdk.AccAddress
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()))
		}

		// Create group
		_, err := s.app.TSSKeeper.CreateGroup(
			s.ctx,
			members,
			tc.Group.Threshold,
			"test",
		)
		s.Require().NoError(err)

		// Set DKG context
		s.app.TSSKeeper.SetDKGContext(s.ctx, tc.Group.ID, tc.Group.DKGContext)
	}
}

func (s *AppTestSuite) setupRound1() {
	s.setupCreateGroup()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit Round 1 information for each member
			_, err := msgSrvr.SubmitDKGRound1(ctx, &types.MsgSubmitDKGRound1{
				GroupID: tc.Group.ID,
				Round1Info: types.Round1Info{
					MemberID:           m.ID,
					CoefficientCommits: m.CoefficientCommits,
					OneTimePubKey:      m.OneTimePubKey(),
					A0Signature:        m.A0Signature,
					OneTimeSignature:   m.OneTimeSignature,
				},
				Sender: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	_, err := app.EndBlocker(ctx.WithBlockHeight(ctx.BlockHeight() + 1))
	s.Require().NoError(err)
}

func (s *AppTestSuite) setupRound2() {
	s.setupRound1()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit Round 2 information for each member
			_, err := msgSrvr.SubmitDKGRound2(ctx, &types.MsgSubmitDKGRound2{
				GroupID: tc.Group.ID,
				Round2Info: types.Round2Info{
					MemberID:              m.ID,
					EncryptedSecretShares: m.EncSecretShares,
				},
				Sender: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	_, err := app.EndBlocker(ctx.WithBlockHeight(ctx.BlockHeight() + 1))
	s.Require().NoError(err)
}

func (s *AppTestSuite) setupConfirm() {
	s.setupRound2()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Confirm the group participation for each member
			_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
				GroupID:      tc.Group.ID,
				MemberID:     m.ID,
				OwnPubKeySig: m.PubKeySignature,
				Sender:       sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	_, err := app.EndBlocker(ctx.WithBlockHeight(ctx.BlockHeight() + 1))
	s.Require().NoError(err)
}

func (s *AppTestSuite) setupDE() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit DEs for each member
			_, err := msgSrvr.SubmitDEs(ctx, &types.MsgSubmitDEs{
				DEs: []types.DE{
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
				},
				Sender: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}
}

func (s *AppTestSuite) SetupGroup(groupStatus types.GroupStatus) {
	switch groupStatus {
	case types.GROUP_STATUS_ROUND_1:
		s.setupCreateGroup()
	case types.GROUP_STATUS_ROUND_2:
		s.setupRound1()
	case types.GROUP_STATUS_ROUND_3:
		s.setupRound2()
	case types.GROUP_STATUS_ACTIVE:
		s.setupConfirm()
		s.setupDE()
	}
}

func (s *AppTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.app.TSSKeeper
	expTime := time.Unix(0, 0)

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

func (s *AppTestSuite) TestParams() {
	k := s.app.TSSKeeper

	testCases := []struct {
		name         string
		input        types.Params
		expectErr    bool
		expectErrStr string
	}{
		{
			name: "set invalid params",
			input: types.Params{
				MaxGroupSize:      0,
				MaxDESize:         0,
				CreationPeriod:    1,
				SigningPeriod:     1,
				MaxSigningAttempt: 1,
				MaxMemoLength:     1,
				MaxMessageLength:  1,
			},
			expectErr:    true,
			expectErrStr: "must be positive:",
		},
		{
			name: "set full valid params",
			input: types.Params{
				MaxGroupSize:      types.DefaultMaxGroupSize,
				MaxDESize:         types.DefaultMaxDESize,
				CreationPeriod:    types.DefaultCreationPeriod,
				SigningPeriod:     types.DefaultSigningPeriod,
				MaxSigningAttempt: types.DefaultMaxSigningAttempt,
				MaxMemoLength:     types.DefaultMaxMemoLength,
				MaxMessageLength:  types.DefaultMaxMessageLength,
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

	ctx           sdk.Context
	keeper        *keeper.Keeper
	msgServer     types.MsgServer
	queryServer   types.QueryServer
	contentRouter *types.ContentRouter
	cbRouter      *types.CallbackRouter

	authzKeeper       *tsstestutil.MockAuthzKeeper
	rollingseedKeeper *tsstestutil.MockRollingseedKeeper

	authority sdk.AccAddress
}

// SetupTest initializes the mock keeper and the context
func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// create a new controller for mocking object
	ctrl := gomock.NewController(s.T())
	s.authzKeeper = tsstestutil.NewMockAuthzKeeper(ctrl)
	s.rollingseedKeeper = tsstestutil.NewMockRollingseedKeeper(ctrl)

	// declare tss components
	s.contentRouter = types.NewContentRouter()
	s.cbRouter = types.NewCallbackRouter()
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	s.keeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		s.authzKeeper,
		s.rollingseedKeeper,
		s.contentRouter,
		s.cbRouter,
		s.authority.String(),
	)
	s.keeper.InitGenesis(s.ctx, *types.DefaultGenesisState())

	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, encCfg.InterfaceRegistry)
	s.queryServer = keeper.NewQueryServer(s.keeper)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.queryServer)

	// add route
	s.contentRouter.AddRoute(types.RouterKey, tssapp.NewSignatureOrderHandler(*s.keeper))
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

// GetExampleSigning returns an example of a signing object.
func GetExampleSigning() types.Signing {
	return types.Signing{
		ID:               1,
		CurrentAttempt:   1,
		GroupID:          1,
		Originator:       []byte("originator"),
		Message:          []byte("data"),
		GroupPubNonce:    testutil.HexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Signature:        nil,
		Status:           types.SIGNING_STATUS_SUCCESS,
		CreatedHeight:    1000,
		CreatedTimestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// GetExampleSigningAttempt returns an example of a signing attempt object.
func GetExampleSigningAttempt() types.SigningAttempt {
	return types.SigningAttempt{
		SigningID:     1,
		Attempt:       1,
		ExpiredHeight: 1050,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: 1,
				Address:  "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     testutil.HexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     testutil.HexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: testutil.HexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
			{
				MemberID: 2,
				Address:  "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
				PubD:     testutil.HexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     testutil.HexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: testutil.HexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
	}
}

// GetExampleGroup returns an example of a group object.
func GetExampleGroup() types.Group {
	return types.Group{
		ID:            1,
		Size_:         3,
		Threshold:     2,
		PubKey:        []byte("test_pubkey"),
		Status:        types.GROUP_STATUS_ACTIVE,
		CreatedHeight: 900,
		ModuleOwner:   "test",
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
