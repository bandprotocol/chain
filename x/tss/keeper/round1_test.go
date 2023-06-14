package keeper_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetRound1Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	round1Data := types.Round1Data{
		MemberID: memberID,
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Data(ctx, groupID, round1Data)

	got, err := k.GetRound1Data(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round1Data, got)
}

func (s *KeeperTestSuite) TestDeleteRound1Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	round1Data := types.Round1Data{
		MemberID: memberID,
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	k.SetRound1Data(ctx, groupID, round1Data)

	got, err := k.GetRound1Data(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round1Data, got)

	k.DeleteRound1Data(ctx, groupID, memberID)

	_, err = k.GetRound1Data(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetAllRound1Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	member2 := tss.MemberID(2)
	round1DataMember1 := types.Round1Data{
		MemberID: member1,
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}
	round1DataMember2 := types.Round1Data{
		MemberID: member2,
		CoefficientsCommit: tss.Points{
			[]byte("point1"),
			[]byte("point2"),
		},
		OneTimePubKey: []byte("OneTimePubKeySimple"),
		A0Sig:         []byte("A0SigSimple"),
		OneTimeSig:    []byte("OneTimeSigSimple"),
	}

	// Set round 1 data
	k.SetRound1Data(ctx, groupID, round1DataMember1)
	k.SetRound1Data(ctx, groupID, round1DataMember2)

	got := k.GetAllRound1Data(ctx, groupID)

	s.Require().Equal([]types.Round1Data{round1DataMember1, round1DataMember2}, got)
}

func (s *KeeperTestSuite) TestGetSetRound1DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set round 1 data count
	k.SetRound1DataCount(ctx, groupID, count)

	got := k.GetRound1DataCount(ctx, groupID)
	s.Require().Equal(uint64(5), got)
}

func (s *KeeperTestSuite) TestDeleteRound1DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set round 1 data count
	k.SetRound1DataCount(ctx, groupID, count)

	// Delete round 1 data count
	k.DeleteRound1DataCount(ctx, groupID)

	got := k.GetRound1DataCount(ctx, groupID)
	s.Require().Empty(got)
}

func (s *KeeperTestSuite) TestGetSetAccumulatedCommit() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	index := uint64(1)
	commit := tss.Point([]byte("point"))

	// Set Accumulated Commit
	k.SetAccumulatedCommit(ctx, groupID, index, commit)

	// Get Accumulated Commit
	got := k.GetAccumulatedCommit(ctx, groupID, index)

	s.Require().Equal(commit, got)
}
