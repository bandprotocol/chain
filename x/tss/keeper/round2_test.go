package keeper_test

import (
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *AppTestSuite) TestGetSetRound2Info() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	round2Info := types.Round2Info{
		MemberID: memberID,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 info
	k.SetRound2Info(ctx, groupID, round2Info)

	got, err := k.GetRound2Info(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round2Info, got)
}

func (s *AppTestSuite) TestAddRound2Info() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	round2Info := types.Round2Info{
		MemberID: memberID,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Add round 2 info
	k.AddRound2Info(ctx, groupID, round2Info)

	gotR2, err := k.GetRound2Info(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(round2Info, gotR2)
	gotR2Count := k.GetRound2InfoCount(ctx, groupID)
	s.Require().Equal(uint64(1), gotR2Count)
}

func (s *AppTestSuite) TestDeleteRound2Infos() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	Round2Info := types.Round2Info{
		MemberID: memberID,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Set round 2 info
	k.SetRound2Info(ctx, groupID, Round2Info)

	// Delete round 2 info
	k.DeleteRound2Infos(ctx, groupID)

	_, err := k.GetRound2Info(ctx, groupID, memberID)
	s.Require().Error(err)

	cnt := k.GetRound2InfoCount(ctx, groupID)
	s.Require().Equal(uint64(0), cnt)
}

func (s *AppTestSuite) TestGetRound2Infos() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	member2 := tss.MemberID(2)
	round2InfoMember1 := types.Round2Info{
		MemberID: member1,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("e_12"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}
	round2InfoMember2 := types.Round2Info{
		MemberID: member2,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("e_11"),
			[]byte("e_13"),
			[]byte("e_14"),
		},
	}

	// Add round 2 info
	k.AddRound2Info(ctx, groupID, round2InfoMember1)
	k.AddRound2Info(ctx, groupID, round2InfoMember2)

	got := k.GetRound2Infos(ctx, groupID)
	s.Require().Equal([]types.Round2Info{round2InfoMember1, round2InfoMember2}, got)
}

func (s *AppTestSuite) TestGetSetRound2InfoCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)

	// Set round 2 info count
	k.AddRound2Info(ctx, groupID, types.Round2Info{MemberID: tss.MemberID(1)})
	k.AddRound2Info(ctx, groupID, types.Round2Info{MemberID: tss.MemberID(2)})

	got := k.GetRound2InfoCount(ctx, groupID)
	s.Require().Equal(uint64(2), got)
}
