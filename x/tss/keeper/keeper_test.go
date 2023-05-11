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
		AppHash: app.LastCommitID().Hash,
	}, Hash: app.LastCommitID().Hash})

	ctx := app.NewContext(false, tmproto.Header{Height: app.LastBlockHeight(), LastCommitHash: app.LastCommitID().Hash})

	s.app = app
	s.ctx = ctx
	s.querier = keeper.Querier{
		app.TSSKeeper,
	}
	s.msgSrvr = app.TSSKeeper
}

func (s *KeeperTestSuite) TestGetSetGroupCount() {
	k := s.app.TSSKeeper
	k.SetGroupCount(s.ctx, 1)

	groupCount := k.GetGroupCount(s.ctx)
	s.Require().Equal(uint64(1), groupCount)
}

func (s *KeeperTestSuite) TestGetNextGroupID() {
	k := s.app.TSSKeeper

	// initial group count
	k.SetGroupCount(s.ctx, 0)

	groupID1 := k.GetNextGroupID(s.ctx)
	s.Require().Equal(uint64(1), groupID1)
	groupID2 := k.GetNextGroupID(s.ctx)
	s.Require().Equal(uint64(2), groupID2)
}

func (s *KeeperTestSuite) TestIsGrantee() {
	k := s.app.TSSKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.MsgGrants {
		s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
	}

	isGrantee := k.IsGrantee(s.ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func (s *KeeperTestSuite) TestCreateNewGroup() {
	k := s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.ROUND_1,
	}

	// Create new group
	groupID := k.CreateNewGroup(s.ctx, group)

	// Get group by id
	got, found := k.GetGroup(s.ctx, groupID)
	s.Require().True(found)
	s.Require().Equal(group, got)
}

func (s *KeeperTestSuite) TestUpdateGroup() {
	k := s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.ROUND_1,
	}

	// Create new group
	groupID := k.CreateNewGroup(s.ctx, group)

	// update group size value
	group.Size_ = 6
	k.UpdateGroup(s.ctx, groupID, group)

	// get group from chain state
	got, found := k.GetGroup(s.ctx, groupID)

	// validate group size value
	s.Require().True(found)
	s.Require().Equal(group.Size_, got.Size_)
}

func (s *KeeperTestSuite) TestGetSetDKGContext() {
	k := s.app.TSSKeeper

	dkgContext := []byte("dkg-context sample")
	k.SetDKGContext(s.ctx, 1, dkgContext)

	got, found := k.GetDKGContext(s.ctx, 1)
	s.Require().True(found)
	s.Require().Equal(dkgContext, got)
}

func (s *KeeperTestSuite) TestGetSetMember() {
	k := s.app.TSSKeeper
	groupID, memberID := uint64(1), uint64(1)
	member := types.Member{
		Signer: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey: "",
	}
	k.SetMember(s.ctx, groupID, memberID, member)

	got, found := k.GetMember(s.ctx, groupID, memberID)
	s.Require().True(found)
	s.Require().Equal(member, got)
}

func (s *KeeperTestSuite) TestGetMembers() {
	k := s.app.TSSKeeper
	groupID := uint64(1)
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
	for i, m := range members {
		k.SetMember(s.ctx, groupID, uint64(i), m)
	}

	got, found := k.GetMembers(s.ctx, groupID)
	s.Require().True(found)
	s.Require().Equal(members, got)
}

func (s *KeeperTestSuite) TestGetMemberID() {
	k := s.app.TSSKeeper
	groupID := uint64(1)
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
	for i, m := range members {
		k.SetMember(s.ctx, groupID, uint64(i), m)
	}

	memberID, found := k.GetMemberID(s.ctx, groupID, "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	s.Require().True(found)
	s.Require().Equal(uint64(1), memberID)
}

func (s *KeeperTestSuite) TestGetSetRound1Commitments() {
	k := s.app.TSSKeeper
	groupID, memberID := uint64(1), uint64(1)
	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: types.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Commitments(s.ctx, groupID, memberID, round1Commitments)

	got, found := k.GetRound1Commitments(s.ctx, groupID, memberID)
	s.Require().True(found)
	s.Require().Equal(round1Commitments, got)
}

func (s *KeeperTestSuite) TestGetRound1CommitmentsCount() {
	k := s.app.TSSKeeper
	groupID, member0, member1 := uint64(1), uint64(0), uint64(1)
	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: types.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round 1 commitments
	k.SetRound1Commitments(s.ctx, groupID, member0, round1Commitments)
	k.SetRound1Commitments(s.ctx, groupID, member1, round1Commitments)

	got := k.GetRound1CommitmentsCount(s.ctx, groupID)
	s.Require().Equal(uint64(2), got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
