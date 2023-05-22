package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app     *testapp.TestingApp
	ctx     sdk.Context
	querier keeper.Querier
	msgSrvr types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	app := testapp.NewTestApp("BANDCHAIN", log.NewNopLogger())

	// commit genesis for test get LastCommitHash in msg create group
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:  app.LastBlockHeight() + 1,
		AppHash: []byte("app-hash sample"),
	}, Hash: []byte("app-hash sample")})

	ctx := app.NewContext(
		false,
		tmproto.Header{Height: app.LastBlockHeight(), LastCommitHash: []byte("app-hash sample")},
	)

	s.app = app
	s.ctx = ctx
	s.querier = keeper.Querier{
		app.TSSKeeper,
	}
	s.msgSrvr = app.TSSKeeper
}

func (s *KeeperTestSuite) TestGetSetGroupCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	k.SetGroupCount(ctx, 1)

	groupCount := k.GetGroupCount(ctx)
	s.Require().Equal(uint64(1), groupCount)
}

func (s *KeeperTestSuite) TestGetNextGroupID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// initial group count
	k.SetGroupCount(ctx, 0)

	groupID1 := k.GetNextGroupID(ctx)
	s.Require().Equal(tss.GroupID(1), groupID1)
	groupID2 := k.GetNextGroupID(ctx)
	s.Require().Equal(tss.GroupID(2), groupID2)
}

func (s *KeeperTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.app.TSSKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.MsgGrants {
		s.app.AuthzKeeper.SaveGrant(ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
	}

	isGrantee := k.IsGrantee(ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func (s *KeeperTestSuite) TestCreateNewGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.ROUND_1,
	}

	// Create new group
	groupID := k.CreateNewGroup(ctx, group)

	// Get group by id
	got, err := k.GetGroup(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(group, got)
}

func (s *KeeperTestSuite) TestUpdateGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.ROUND_1,
	}

	// Create new group
	groupID := k.CreateNewGroup(ctx, group)

	// update group size value
	group.Size_ = 6
	k.UpdateGroup(ctx, groupID, group)

	// get group from chain state
	got, err := k.GetGroup(ctx, groupID)

	// validate group size value
	s.Require().NoError(err)
	s.Require().Equal(group.Size_, got.Size_)
}

func (s *KeeperTestSuite) TestGetSetDKGContext() {
	ctx, k := s.ctx, s.app.TSSKeeper

	dkgContext := []byte("dkg-context sample")
	k.SetDKGContext(ctx, 1, dkgContext)

	got, err := k.GetDKGContext(ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(dkgContext, got)
}

func (s *KeeperTestSuite) TestGetSetMember() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		Signer: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey: "",
	}
	k.SetMember(ctx, groupID, memberID, member)

	got, err := k.GetMember(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(member, got)
}

func (s *KeeperTestSuite) TestGetSetMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			Signer: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey: "",
		},
		{
			Signer: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey: "",
		},
	}

	// set members
	k.SetMembers(ctx, groupID, members)

	got, err := k.GetMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(members, got)
}

func (s *KeeperTestSuite) TesVerifyMember() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			Signer: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey: "",
		},
		{
			Signer: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey: "",
		},
	}

	// set members
	k.SetMembers(ctx, groupID, members)

	memberID1, err := k.VerifyMember(ctx, groupID, "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), memberID1)
	memberID2, err := k.VerifyMember(ctx, groupID, "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), memberID2)
}

func (s *KeeperTestSuite) TestGetSetRound1Commitments() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round1Commitment := types.Round1Commitment{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Commitment(ctx, groupID, memberID, round1Commitment)

	got, err := k.GetRound1Commitment(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round1Commitment, got)
}

func (s *KeeperTestSuite) TestDeleteRound1Commitments() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round1Commitment := types.Round1Commitment{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Commitment(ctx, groupID, memberID, round1Commitment)

	got, err := k.GetRound1Commitment(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round1Commitment, got)

	k.DeleteRound1Commitment(ctx, groupID, memberID)

	_, err = k.GetRound1Commitment(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetRound1CommitmentsCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, member0, member1 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	round1Commitment := types.Round1Commitment{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round 1 commitments
	k.SetRound1Commitment(ctx, groupID, member0, round1Commitment)
	k.SetRound1Commitment(ctx, groupID, member1, round1Commitment)

	got := k.GetRound1CommitmentsCount(ctx, groupID)
	s.Require().Equal(uint64(2), got)
}

func (s *KeeperTestSuite) TestGetAllRound1Commitments() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, groupSize, member1, member2 := tss.GroupID(1), uint64(3), tss.MemberID(1), tss.MemberID(2)
	round1Commitment := types.Round1Commitment{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round 1 commitments
	k.SetRound1Commitment(ctx, groupID, member1, round1Commitment)
	k.SetRound1Commitment(ctx, groupID, member2, round1Commitment)

	got := k.GetAllRound1Commitments(ctx, groupID, groupSize)

	// member3 expected nil value because didn't commit round 1
	s.Require().Equal([]*types.Round1Commitment{&round1Commitment, &round1Commitment, nil}, got)
}

func (s *KeeperTestSuite) TestGetSetRound2Share() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round2Share := types.Round2Share{
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round 2 secret share
	k.SetRound2Share(ctx, groupID, memberID, round2Share)

	got, err := k.GetRound2Share(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round2Share, got)
}

func (s *KeeperTestSuite) TestDeleteRound2Share() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round2Share := types.Round2Share{
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round 2 secret share
	k.SetRound2Share(ctx, groupID, memberID, round2Share)

	// delete round 2 secret share
	k.DeleteRound2share(ctx, groupID, memberID)

	_, err := k.GetRound2Share(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetRound2SharesCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, member1, member2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	round2ShareM1 := types.Round2Share{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2ShareM2 := types.Round2Share{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round 2 secret share
	k.SetRound2Share(ctx, groupID, member1, round2ShareM1)
	k.SetRound2Share(ctx, groupID, member2, round2ShareM2)

	got := k.GetRound2SharesCount(ctx, groupID)
	s.Require().Equal(uint64(2), got)
}

func (s *KeeperTestSuite) TestGetAllRound2Shares() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, groupSize, member1, member2 := tss.GroupID(1), uint64(3), tss.MemberID(1), tss.MemberID(2)
	round2ShareM1 := types.Round2Share{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2ShareM2 := types.Round2Share{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round 2 secret share
	k.SetRound2Share(ctx, groupID, member1, round2ShareM1)
	k.SetRound2Share(ctx, groupID, member2, round2ShareM2)

	got := k.GetAllRound2Shares(ctx, groupID, groupSize)
	// member3 expected nil value because didn't submit round 2 share
	s.Require().Equal([]*types.Round2Share{&round2ShareM1, &round2ShareM2, nil}, got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
