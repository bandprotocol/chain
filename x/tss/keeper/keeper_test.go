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

	// Commit genesis for test get LastCommitHash in msg create group
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

	// Initial group count
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

	// Update group size value
	group.Size_ = 6
	k.UpdateGroup(ctx, groupID, group)

	// Get group from chain state
	got, err := k.GetGroup(ctx, groupID)

	// Validate group size value
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
		Member: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey: tss.PublicKey(nil),
	}
	k.SetMember(ctx, groupID, memberID, member)

	got, err := k.GetMember(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(member, got)
}

func (s *KeeperTestSuite) TestGetMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			Member: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey: tss.PublicKey(nil),
		},
		{
			Member: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey: tss.PublicKey(nil),
		},
	}

	// set members
	for i, m := range members {
		k.SetMember(ctx, groupID, tss.MemberID(i+1), m)
	}

	got, err := k.GetMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(members, got)
}

func (s *KeeperTestSuite) TesGetMemberID() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			Member: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey: tss.PublicKey(nil),
		},
		{
			Member: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey: tss.PublicKey(nil),
		},
	}

	// set members
	for i, m := range members {
		k.SetMember(ctx, groupID, tss.MemberID(i+1), m)
	}

	memberID1, err := k.GetMemberID(ctx, groupID, "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), memberID1)
	memberID2, err := k.GetMemberID(ctx, groupID, "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), memberID2)
}

func (s *KeeperTestSuite) TestGetSetRound1Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	Round1Data := types.Round1Data{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Data(ctx, groupID, memberID, Round1Data)

	got, err := k.GetRound1Data(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(Round1Data, got)
}

func (s *KeeperTestSuite) TestDeleteRound1Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	Round1Data := types.Round1Data{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Data(ctx, groupID, memberID, Round1Data)

	got, err := k.GetRound1Data(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(Round1Data, got)

	k.DeleteRound1Data(ctx, groupID, memberID)

	_, err = k.GetRound1Data(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetRound1DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, member0, member1 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	Round1Data := types.Round1Data{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round1 data
	k.SetRound1Data(ctx, groupID, member0, Round1Data)
	k.SetRound1Data(ctx, groupID, member1, Round1Data)

	got := k.GetRound1DataCount(ctx, groupID)
	s.Require().Equal(uint64(2), got)
}

func (s *KeeperTestSuite) TestGetAllRound1Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, groupSize, member1, member2 := tss.GroupID(1), uint64(3), tss.MemberID(1), tss.MemberID(2)
	Round1Data := types.Round1Data{
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round1 data
	k.SetRound1Data(ctx, groupID, member1, Round1Data)
	k.SetRound1Data(ctx, groupID, member2, Round1Data)

	got := k.GetAllRound1Data(ctx, groupID, groupSize)

	// member3 expected nil value because didn't commit round1
	s.Require().Equal([]*types.Round1Data{&Round1Data, &Round1Data, nil}, got)
}

func (s *KeeperTestSuite) TestGetSetround2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round2Data := types.Round2Data{
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round2 secret share
	k.SetRound2Data(ctx, groupID, memberID, round2Data)

	got, err := k.GetRound2Data(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round2Data, got)
}

func (s *KeeperTestSuite) TestDeleteround2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	round2Data := types.Round2Data{
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round2secret data
	k.SetRound2Data(ctx, groupID, memberID, round2Data)

	// delete round2secret data
	k.DeleteRound2Data(ctx, groupID, memberID)

	_, err := k.GetRound2Data(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetround2DatasCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, member1, member2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	round2DataM1 := types.Round2Data{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2DataM2 := types.Round2Data{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// set round2secret share
	k.SetRound2Data(ctx, groupID, member1, round2DataM1)
	k.SetRound2Data(ctx, groupID, member2, round2DataM2)

	got := k.GetRound2DataCount(ctx, groupID)
	s.Require().Equal(uint64(2), got)
}

func (s *KeeperTestSuite) TestGetAllround2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, groupSize, member1, member2 := tss.GroupID(1), uint64(3), tss.MemberID(1), tss.MemberID(2)
	round2DataM1 := types.Round2Data{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2DataM2 := types.Round2Data{
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round2 data
	k.SetRound2Data(ctx, groupID, member1, round2DataM1)
	k.SetRound2Data(ctx, groupID, member2, round2DataM2)

	got := k.GetAllRound2Data(ctx, groupID, groupSize)
	// member3 expected nil value because didn't submit round2Data
	s.Require().Equal([]*types.Round2Data{&round2DataM1, &round2DataM2, nil}, got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
