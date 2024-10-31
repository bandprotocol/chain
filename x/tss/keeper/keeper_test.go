package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

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
	tssapp "github.com/bandprotocol/chain/v3/x/tss"
	"github.com/bandprotocol/chain/v3/x/tss/keeper"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

var (
	PrivD = testutil.HexDecode("de6aedbe8ba688dd6d342881eb1e67c3476e825106477360148e2858a5eb565c")
	PrivE = testutil.HexDecode("3ff4fb2beac0cee0ab230829a5ae0881310046282e79c978ca22f44897ea434a")
	PubD  = tss.Scalar(PrivD).Point()
	PubE  = tss.Scalar(PrivE).Point()
)

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

// newMockRound1Info generates a mock object of round1Info.
func newMockRound1Info(memberID tss.MemberID) types.Round1Info {
	return types.Round1Info{
		MemberID: memberID,
		CoefficientCommits: []tss.Point{
			[]byte("point1"),
			[]byte("point2"),
			[]byte("point3"),
		},
		OneTimePubKey:    []byte("OneTimePubKeySample"),
		A0Signature:      []byte("A0SignatureSample"),
		OneTimeSignature: []byte("OneTimeSignatureSample"),
	}
}

// newMockRound2Info generates a mock object of round2Info.
func newMockRound2Info(memberID tss.MemberID) types.Round2Info {
	return types.Round2Info{
		MemberID:              memberID,
		EncryptedSecretShares: tss.EncSecretShares{[]byte("secret1"), []byte("secret2")},
	}
}

// newMockComplaintsWithStatus generates a mock object of ComplaintsWithStatus
func newMockComplaintsWithStatus(complainant, respondent tss.MemberID) types.ComplaintsWithStatus {
	return types.ComplaintsWithStatus{
		MemberID: complainant,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainant: complainant,
					Respondent:  respondent,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}
}

// SetupWithPreparedTestCase sets up the group to the given status with a given prepared test case number.
func (s *KeeperTestSuite) SetupWithPreparedTestCase(testCaseNo int, groupStatus types.GroupStatus) {
	testCase := testutil.TestCases[testCaseNo].Group

	var members []sdk.AccAddress
	for _, m := range testCase.Members {
		members = append(members, sdk.AccAddress(m.PubKey()))
	}

	// Setup group
	_, err := s.keeper.CreateGroup(s.ctx, members, testCase.Threshold, "test")
	s.Require().NoError(err)
	s.keeper.SetDKGContext(s.ctx, testCase.ID, testCase.DKGContext)

	group, err := s.keeper.GetGroup(s.ctx, testCase.ID)
	s.Require().NoError(err)
	s.Require().Equal(types.GROUP_STATUS_ROUND_1, group.Status)

	if groupStatus == types.GROUP_STATUS_ROUND_1 {
		return
	}

	// Member submit round1 info.
	for _, m := range testCase.Members {
		_, err := s.msgServer.SubmitDKGRound1(s.ctx, &types.MsgSubmitDKGRound1{
			GroupID: testCase.ID,
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

	// Execute the EndBlocker to process groups
	err = tssapp.EndBlocker(s.ctx.WithBlockHeight(s.ctx.BlockHeight()+1), s.keeper)
	s.Require().NoError(err)

	group, err = s.keeper.GetGroup(s.ctx, testCase.ID)
	s.Require().NoError(err)
	s.Require().Equal(types.GROUP_STATUS_ROUND_2, group.Status)

	if groupStatus == types.GROUP_STATUS_ROUND_2 {
		return
	}

	// Member submit round2 info.
	for _, m := range testCase.Members {
		_, err := s.msgServer.SubmitDKGRound2(s.ctx, &types.MsgSubmitDKGRound2{
			GroupID: testCase.ID,
			Round2Info: types.Round2Info{
				MemberID:              m.ID,
				EncryptedSecretShares: m.EncSecretShares,
			},
			Sender: sdk.AccAddress(m.PubKey()).String(),
		})
		s.Require().NoError(err)
	}

	// Execute the EndBlocker to process groups
	err = tssapp.EndBlocker(s.ctx.WithBlockHeight(s.ctx.BlockHeight()+1), s.keeper)
	s.Require().NoError(err)

	group, err = s.keeper.GetGroup(s.ctx, testCase.ID)
	s.Require().NoError(err)
	s.Require().Equal(types.GROUP_STATUS_ROUND_3, group.Status)

	if groupStatus == types.GROUP_STATUS_ROUND_3 {
		return
	}

	// Confirm the participation of each member in the group
	for _, m := range testCase.Members {
		_, err := s.msgServer.Confirm(s.ctx, &types.MsgConfirm{
			GroupID:      testCase.ID,
			MemberID:     m.ID,
			OwnPubKeySig: m.PubKeySignature,
			Sender:       sdk.AccAddress(m.PubKey()).String(),
		})
		s.Require().NoError(err)
	}

	// Execute the EndBlocker to process groups
	err = tssapp.EndBlocker(s.ctx.WithBlockHeight(s.ctx.BlockHeight()+1), s.keeper)
	s.Require().NoError(err)

	// Check the group's status and expiration time after confirmation
	got, err := s.keeper.GetGroup(s.ctx, testCase.ID)
	s.Require().NoError(err)
	s.Require().Equal(types.GROUP_STATUS_ACTIVE, got.Status)

	// submit DEs for each member
	for _, m := range testCase.Members {
		_, err := s.msgServer.SubmitDEs(s.ctx, &types.MsgSubmitDEs{
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

func (s *KeeperTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.keeper
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	genericAuthz := authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgSubmitDKGRound1{}))
	s.authzKeeper.EXPECT().
		GetAuthorization(gomock.Any(), grantee, granter, gomock.Any()).
		Return(genericAuthz, nil).
		AnyTimes()

	isGrantee := k.CheckIsGrantee(ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
