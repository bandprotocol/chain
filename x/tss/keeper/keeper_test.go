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
	got, found := k.GetGroup(ctx, groupID)
	s.Require().True(found)
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
	got, found := k.GetGroup(ctx, groupID)

	// validate group size value
	s.Require().True(found)
	s.Require().Equal(group.Size_, got.Size_)
}

func (s *KeeperTestSuite) TestGetSetDKGContext() {
	ctx, k := s.ctx, s.app.TSSKeeper

	dkgContext := []byte("dkg-context sample")
	k.SetDKGContext(ctx, 1, dkgContext)

	got, found := k.GetDKGContext(ctx, 1)
	s.Require().True(found)
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

	got, found := k.GetMember(ctx, groupID, memberID)
	s.Require().True(found)
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

	got, found := k.GetMembers(ctx, groupID)
	s.Require().True(found)
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

	isMember1 := k.VerifyMember(ctx, groupID, 0, "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	s.Require().True(isMember1)
	isMember2 := k.VerifyMember(ctx, groupID, 1, "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	s.Require().True(isMember2)
}

func (s *KeeperTestSuite) TestGetSetRound1Commitments() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Commitments(ctx, groupID, memberID, round1Commitments)

	got, found := k.GetRound1Commitments(ctx, groupID, memberID)
	s.Require().True(found)
	s.Require().Equal(round1Commitments, got)
}

func (s *KeeperTestSuite) TestDeleteRound1Commitments() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Commitments(ctx, groupID, memberID, round1Commitments)

	got, found := k.GetRound1Commitments(ctx, groupID, memberID)
	s.Require().True(found)
	s.Require().Equal(round1Commitments, got)

	k.DeleteRound1Commitments(ctx, groupID, memberID)

	_, found = k.GetRound1Commitments(ctx, groupID, memberID)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestGetRound1CommitmentsCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, member0, member1 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round 1 commitments
	k.SetRound1Commitments(ctx, groupID, member0, round1Commitments)
	k.SetRound1Commitments(ctx, groupID, member1, round1Commitments)

	got := k.GetRound1CommitmentsCount(ctx, groupID)
	s.Require().Equal(uint64(2), got)
}

func (s *KeeperTestSuite) TestGetAllRound1Commitments() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, member0, member1 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	s.T().Log(types.Round1CommitmentsStoreKey(1))

	// Set round 1 commitments
	k.SetRound1Commitments(ctx, groupID, member0, round1Commitments)
	k.SetRound1Commitments(ctx, groupID, member1, round1Commitments)

	got, found := k.GetAllRound1Commitments(ctx, groupID)
	s.Require().True(found)

	s.Require().Equal(round1Commitments, got[1])
	s.Require().Equal(round1Commitments, got[2])
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
