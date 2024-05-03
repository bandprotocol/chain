package keeper_test

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type KeeperTestSuite struct {
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

func (s *KeeperTestSuite) SetupTest() {
	app, ctx := bandtesting.CreateTestApp(s.T(), true)
	s.app = app
	s.ctx = ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.BandtssKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	s.queryClient = queryClient
	s.msgSrvr = keeper.NewMsgServerImpl(app.BandtssKeeper)
	s.tssMsgSrvr = tsskeeper.NewMsgServerImpl(app.TSSKeeper)
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
}

func (s *KeeperTestSuite) setupCreateGroup() {
	ctx, bandtssMsgSrvr, tssKeeper := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Initialize members
		var members []string
		for _, m := range tc.Group.Members {
			address := sdk.AccAddress(m.PubKey())
			members = append(members, address.String())
		}

		// Create group
		_, err := bandtssMsgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Authority: s.authority.String(),
		})
		s.Require().NoError(err)

		// Set DKG context
		tssKeeper.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)
	}
}

func (s *KeeperTestSuite) setupRound1() {
	s.setupCreateGroup()

	ctx, app, tssMsgSrvr := s.ctx, s.app, s.tssMsgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit Round 1 information for each member
			_, err := tssMsgSrvr.SubmitDKGRound1(ctx, &tsstypes.MsgSubmitDKGRound1{
				GroupID: tc.Group.ID,
				Round1Info: tsstypes.Round1Info{
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

	ctx, app, tssMsgSrvr := s.ctx, s.app, s.tssMsgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit Round 2 information for each member
			_, err := tssMsgSrvr.SubmitDKGRound2(ctx, &tsstypes.MsgSubmitDKGRound2{
				GroupID: tc.Group.ID,
				Round2Info: tsstypes.Round2Info{
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

	ctx, app, tssMsgSrvr := s.ctx, s.app, s.tssMsgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Confirm the group participation for each member
			_, err := tssMsgSrvr.Confirm(ctx, &tsstypes.MsgConfirm{
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
	ctx, tssMsgSrvr := s.ctx, s.tssMsgSrvr

	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit DEs for each member
			_, err := tssMsgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
				DEs: []tsstypes.DE{
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

func (s *KeeperTestSuite) SetupGroup(groupStatus tsstypes.GroupStatus) {
	switch groupStatus {
	case tsstypes.GROUP_STATUS_ROUND_1:
		s.setupCreateGroup()
	case tsstypes.GROUP_STATUS_ROUND_2:
		s.setupRound1()
	case tsstypes.GROUP_STATUS_ROUND_3:
		s.setupRound2()
	case tsstypes.GROUP_STATUS_ACTIVE:
		s.setupConfirm()
		s.setupDE()
	}
}

func (s *KeeperTestSuite) TestParams() {
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
				JailPenaltyDuration:     time.Duration(0),
				RewardPercentage:        0,
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
				JailPenaltyDuration:     types.DefaultJailPenaltyDuration,
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
