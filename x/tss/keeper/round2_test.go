package keeper_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetRound2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	round2Data := types.Round2Data{
		MemberID: memberID,
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round2 secret share
	k.SetRound2Data(ctx, groupID, round2Data)

	got, err := k.GetRound2Data(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round2Data, got)
}

func (s *KeeperTestSuite) TestDeleteRound2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	round2Data := types.Round2Data{
		MemberID: memberID,
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 secret data
	k.SetRound2Data(ctx, groupID, round2Data)

	// Delete round 2 secret data
	k.DeleteRound2Data(ctx, groupID, memberID)

	_, err := k.GetRound2Data(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetAllRound2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	member2 := tss.MemberID(2)
	round2DataMember1 := types.Round2Data{
		MemberID: member1,
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2DataMember2 := types.Round2Data{
		MemberID: member2,
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 data
	k.SetRound2Data(ctx, groupID, round2DataMember1)
	k.SetRound2Data(ctx, groupID, round2DataMember2)

	got := k.GetAllRound2Data(ctx, groupID)
	// Member3 expected nil value because didn't submit round2Data
	s.Require().Equal([]types.Round2Data{round2DataMember1, round2DataMember2}, got)
}

func (s *KeeperTestSuite) TestGetSetRound2DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set round 2 data count
	k.SetRound2DataCount(ctx, groupID, count)

	got := k.GetRound2DataCount(ctx, groupID)
	s.Require().Equal(uint64(5), got)
}

func (s *KeeperTestSuite) TestDeleteRound2DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set round 2 data count
	k.SetRound2DataCount(ctx, groupID, count)

	// Delete round 2 data count
	k.DeleteRound2DataCount(ctx, groupID)

	got := k.GetRound2DataCount(ctx, groupID)
	s.Require().Empty(got)
}
