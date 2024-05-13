package keeper_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	bandtsskeeper "github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *bandtesting.TestingApp
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

func (s *KeeperTestSuite) SetupTest() {
	app, ctx := bandtesting.CreateTestApp(s.T(), true)
	s.app = app
	s.ctx = ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.TSSKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	s.queryClient = queryClient
	s.msgSrvr = keeper.NewMsgServerImpl(app.TSSKeeper)
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
}

func (s *KeeperTestSuite) setupCreateGroup() {
	ctx, bandtssKeeper := s.ctx, s.app.BandtssKeeper
	bandtssMsgSrvr := bandtsskeeper.NewMsgServerImpl(bandtssKeeper)

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Initialize members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		_, err := bandtssMsgSrvr.CreateGroup(ctx, &bandtsstypes.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Authority: s.authority.String(),
		})
		s.Require().NoError(err)

		// Set DKG context
		s.app.TSSKeeper.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)
	}
}

func (s *KeeperTestSuite) setupRound1() {
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
				Address: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})
}

func (s *KeeperTestSuite) setupRound2() {
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
				Address: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})
}

func (s *KeeperTestSuite) setupConfirm() {
	s.setupRound2()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Confirm the group participation for each member
			_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
				GroupID:      tc.Group.ID,
				MemberID:     m.ID,
				OwnPubKeySig: m.PubKeySignature,
				Address:      sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})
}

func (s *KeeperTestSuite) setupDE() {
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
				Address: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}
}

func (s *KeeperTestSuite) SetupGroup(groupStatus types.GroupStatus) {
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

func (s *KeeperTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.app.TSSKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.TSSGrantMsgTypes {
		err := s.app.AuthzKeeper.SaveGrant(ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
		s.Require().NoError(err)
	}

	isGrantee := k.CheckIsGrantee(ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func (s *KeeperTestSuite) TestParams() {
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
				MaxGroupSize:   0,
				MaxDESize:      0,
				CreatingPeriod: 1,
				SigningPeriod:  1,
			},
			expectErr:    true,
			expectErrStr: "must be positive:",
		},
		{
			name: "set full valid params",
			input: types.Params{
				MaxGroupSize:   types.DefaultMaxGroupSize,
				MaxDESize:      types.DefaultMaxDESize,
				CreatingPeriod: types.DefaultCreatingPeriod,
				SigningPeriod:  types.DefaultSigningPeriod,
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
