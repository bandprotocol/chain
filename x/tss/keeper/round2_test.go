package keeper_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetRound2Info() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	Round2Info := types.Round2Info{
		MemberID: memberID,
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 info
	k.SetRound2Info(ctx, groupID, Round2Info)

	got, err := k.GetRound2Info(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(Round2Info, got)
}

func (s *KeeperTestSuite) TestDeleteRound2Info() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	Round2Info := types.Round2Info{
		MemberID: memberID,
		EncryptedSecretShares: tss.Scalars{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 info
	k.SetRound2Info(ctx, groupID, Round2Info)

	// Delete round 2 info
	k.DeleteRound2Info(ctx, groupID, memberID)

	_, err := k.GetRound2Info(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetRound2Infos() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	member2 := tss.MemberID(2)
	round2InfoMember1 := types.Round2Info{
		MemberID: member1,
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2InfoMember2 := types.Round2Info{
		MemberID: member2,
		EncryptedSecretShares: []tss.Scalar{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 info
	k.SetRound2Info(ctx, groupID, round2InfoMember1)
	k.SetRound2Info(ctx, groupID, round2InfoMember2)

	got := k.GetRound2Infos(ctx, groupID)
	s.Require().Equal([]types.Round2Info{round2InfoMember1, round2InfoMember2}, got)
}

func (s *KeeperTestSuite) TestGetSetRound2InfoCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set round 2 info count
	k.SetRound2InfoCount(ctx, groupID, count)

	got := k.GetRound2InfoCount(ctx, groupID)
	s.Require().Equal(uint64(5), got)
}

func (s *KeeperTestSuite) TestDeleteRound2InfoCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set round 2 info count
	k.SetRound2InfoCount(ctx, groupID, count)

	// Delete round 2 info count
	k.DeleteRound2InfoCount(ctx, groupID)

	got := k.GetRound2InfoCount(ctx, groupID)
	s.Require().Empty(got)
}
