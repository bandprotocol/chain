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
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
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
	for _, m := range types.GetMsgGrants() {
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
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Create new group
	groupID := k.CreateNewGroup(ctx, group)

	// init group ID
	group.GroupID = tss.GroupID(1)

	// Get group by id
	got, err := k.GetGroup(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(group, got)
}

func (s *KeeperTestSuite) TestSetGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Set new group
	groupID := k.CreateNewGroup(ctx, group)

	// Update group size value
	group.Size_ = 6

	// Add group ID
	group.GroupID = 1

	k.SetGroup(ctx, group)

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
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      tss.PublicKey(nil),
		IsMalicious: false,
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
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
		},
		{
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
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

func (s *KeeperTestSuite) TestVerifyMember() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
		},
		{
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
		},
	}

	// set members
	for i, m := range members {
		k.SetMember(ctx, groupID, tss.MemberID(i+1), m)
	}

	isMember1 := k.VerifyMember(ctx, groupID, tss.MemberID(1), "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	s.Require().True(isMember1)
	isMember2 := k.VerifyMember(ctx, groupID, tss.MemberID(2), "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	s.Require().True(isMember2)
}

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

	// member3 expected nil value because didn't commit round 1
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

	// set round2 secret share
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

	// set round 2 secret data
	k.SetRound2Data(ctx, groupID, round2Data)

	// delete round 2 secret data
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
	// member3 expected nil value because didn't submit round2Data
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

func (s *KeeperTestSuite) TestGetMaliciousMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID1 := tss.MemberID(1)
	memberID2 := tss.MemberID(2)
	member1 := types.Member{
		Address:     "member_address_1",
		PubKey:      []byte("pub_key"),
		IsMalicious: true,
	}
	member2 := types.Member{
		Address:     "member_address_2",
		PubKey:      []byte("pub_key"),
		IsMalicious: true,
	}

	// Set member
	k.SetMember(ctx, groupID, memberID1, member1)
	k.SetMember(ctx, groupID, memberID2, member2)

	// Get malicious members
	got, err := k.GetMaliciousMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal([]types.Member{member1, member2}, got)
}

func (s *KeeperTestSuite) TestHandleVerifyComplainSig() {
	ctx, k := s.ctx, s.app.TSSKeeper

	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Set member
			k.SetMember(ctx, tc.Group.ID, m.ID, types.Member{
				Address:     "member_address",
				PubKey:      m.PubKey(),
				IsMalicious: false,
			})
		}

		slot := testutil.GetSlot(tc.Group.Members[0].ID, tc.Group.Members[1].ID)

		err := tss.VerifyComplainSig(
			tc.Group.Members[0].OneTimePubKey(),
			tc.Group.Members[1].OneTimePubKey(),
			tc.Group.Members[0].KeySyms[slot],
			tc.Group.Members[0].ComplainSigs[slot],
		)
		s.Require().NoError(err)
	}
}

func (s *KeeperTestSuite) TestHandleVerifyOwnPubKeySig() {
	ctx, k := s.ctx, s.app.TSSKeeper

	for _, tc := range testutil.TestCases {
		// Set dkg context
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)

		for _, m := range tc.Group.Members {
			// Set member
			k.SetMember(ctx, tc.Group.ID, m.ID, types.Member{
				Address:     "member_address",
				PubKey:      m.PubKey(),
				IsMalicious: false,
			})

			// Sign
			sig, err := tss.SignOwnPubkey(m.ID, tc.Group.DKGContext, m.PubKey(), m.PrivKey)
			s.Require().NoError(err)

			// Verify own public key signature
			err = k.HandleVerifyOwnPubKeySig(ctx, tc.Group.ID, m.ID, sig)
			s.Require().NoError(err)
		}

	}
}

func (s *KeeperTestSuite) TestGetSetComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	complainWithStatus := types.ComplainsWithStatus{
		MemberID: memberID,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:      1,
					J:      2,
					KeySym: []byte("key_sym"),
					Sig:    []byte("signature"),
				},
				ComplainStatus: types.COMPLAIN_STATUS_SUCCESS,
			},
		},
	}

	// Set complains with status
	k.SetComplainsWithStatus(ctx, groupID, complainWithStatus)

	got, err := k.GetComplainsWithStatus(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(complainWithStatus, got)
}

func (s *KeeperTestSuite) TestDeleteComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	complainWithStatus := types.ComplainsWithStatus{
		MemberID: memberID,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:      1,
					J:      2,
					KeySym: []byte("key_sym"),
					Sig:    []byte("signature"),
				},
				ComplainStatus: types.COMPLAIN_STATUS_SUCCESS,
			},
		},
	}

	// Set complains with status
	k.SetComplainsWithStatus(ctx, groupID, complainWithStatus)
	// Delete complains with status
	k.DeleteComplainsWithStatus(ctx, groupID, memberID)

	_, err := k.GetComplainsWithStatus(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetAllComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID1 := tss.MemberID(1)
	memberID2 := tss.MemberID(2)
	complainWithStatus1 := types.ComplainsWithStatus{
		MemberID: memberID1,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:      1,
					J:      2,
					KeySym: []byte("key_sym"),
					Sig:    []byte("signature"),
				},
				ComplainStatus: types.COMPLAIN_STATUS_SUCCESS,
			},
		},
	}
	complainWithStatus2 := types.ComplainsWithStatus{
		MemberID: memberID2,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:      1,
					J:      2,
					KeySym: []byte("key_sym"),
					Sig:    []byte("signature"),
				},
				ComplainStatus: types.COMPLAIN_STATUS_SUCCESS,
			},
		},
	}

	// Set complains with status
	k.SetComplainsWithStatus(ctx, groupID, complainWithStatus1)
	k.SetComplainsWithStatus(ctx, groupID, complainWithStatus2)

	got := k.GetAllComplainsWithStatus(ctx, groupID)
	s.Require().Equal([]types.ComplainsWithStatus{complainWithStatus1, complainWithStatus2}, got)
}

func (s *KeeperTestSuite) TestGetSetConfirm() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	confirm := types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, confirm)

	got, err := k.GetConfirm(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(confirm, got)

	// Get confirm or complain count
	count := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(1), count)
}

func (s *KeeperTestSuite) TestDeleteConfirm() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	confirm := types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, confirm)

	// Delete confirm
	k.DeleteConfirm(ctx, groupID, memberID)

	_, err := k.GetConfirm(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetConfirms() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID1 := tss.MemberID(1)
	memberID2 := tss.MemberID(2)
	confirm1 := types.Confirm{
		MemberID:     memberID1,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}
	confirm2 := types.Confirm{
		MemberID:     memberID2,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, confirm1)
	k.SetConfirm(ctx, groupID, confirm2)

	got := k.GetConfirms(ctx, groupID)
	s.Require().Equal([]types.Confirm{confirm1, confirm2}, got)
}

func (s *KeeperTestSuite) TestGetSetConfirmComplainCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(1)

	// Get confirm complain count before assign
	got1 := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(0), got1)

	// Set confirm complain count
	k.SetConfirmComplainCount(ctx, groupID, count)

	// Get confirm complain count
	got2 := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(count, got2)
}

func (s *KeeperTestSuite) TestDeleteConfirmComplainCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set confirm complain count
	k.SetConfirmComplainCount(ctx, groupID, count)

	// Delete confirm complain count
	k.DeleteConfirmComplainCount(ctx, groupID)

	got := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Empty(got)
}

func (s *KeeperTestSuite) TestMarkMalicious() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)

	// Set member
	k.SetMember(ctx, groupID, memberID, types.Member{
		Address:     "member_address",
		PubKey:      []byte("pub_key"),
		IsMalicious: false,
	})

	// Mark malicious
	err := k.MarkMalicious(ctx, groupID, memberID)
	s.Require().NoError(err)

	got, err := k.GetMaliciousMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal([]types.Member{
		{
			Address:     "member_address",
			PubKey:      []byte("pub_key"),
			IsMalicious: true,
		},
	}, got)
}

func (s *KeeperTestSuite) TestDeleteAllDKGInterimData() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	groupSize := uint64(5)
	groupThreshold := uint64(3)
	dkgContext := []byte("dkg-context")

	// Assuming you have corresponding Set methods for each Delete method
	// Setting up initial state
	k.SetDKGContext(ctx, groupID, dkgContext)

	for i := uint64(1); i <= groupSize; i++ {
		memberID := tss.MemberID(i)
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
		round2Data := types.Round2Data{
			MemberID: memberID,
			EncryptedSecretShares: tss.Scalars{
				[]byte("e_12"),
				[]byte("e_13"),
				[]byte("e_14"),
			},
		}
		complainWithStatus := types.ComplainsWithStatus{
			MemberID: memberID,
			ComplainsWithStatus: []types.ComplainWithStatus{
				{
					Complain: types.Complain{
						I:      1,
						J:      2,
						KeySym: []byte("key_sym"),
						Sig:    []byte("signature"),
					},
					ComplainStatus: types.COMPLAIN_STATUS_SUCCESS,
				},
			},
		}
		confirm := types.Confirm{
			MemberID:     memberID,
			OwnPubKeySig: []byte("own_pub_key_sig"),
		}

		k.SetRound1Data(ctx, groupID, round1Data)
		k.SetRound2Data(ctx, groupID, round2Data)
		k.SetComplainsWithStatus(ctx, groupID, complainWithStatus)
		k.SetConfirm(ctx, groupID, confirm)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		k.SetAccumulatedCommit(ctx, groupID, i, []byte("point1"))
	}

	k.SetRound1DataCount(ctx, groupID, 1)
	k.SetRound2DataCount(ctx, groupID, 1)
	k.SetConfirmComplainCount(ctx, groupID, 1)

	// Delete all interim data
	k.DeleteAllDKGInterimData(ctx, groupID, groupSize, groupThreshold)

	// Check if all data is deleted
	s.Require().Nil(k.GetDKGContext(ctx, groupID))

	for i := uint64(1); i <= groupSize; i++ {
		memberID := tss.MemberID(i)

		gotRound1Data, err := k.GetRound1Data(ctx, groupID, memberID)
		s.Require().ErrorIs(types.ErrRound1DataNotFound, err)
		s.Require().Empty(types.Round1Data{}, gotRound1Data)

		gotRound2Data, err := k.GetRound2Data(ctx, groupID, memberID)
		s.Require().ErrorIs(types.ErrRound2DataNotFound, err)
		s.Require().Empty(types.Round2Data{}, gotRound2Data)

		gotComplainWithStatus, err := k.GetComplainsWithStatus(ctx, groupID, memberID)
		s.Require().ErrorIs(types.ErrComplainsWithStatusNotFound, err)
		s.Require().Empty(types.ComplainWithStatus{}, gotComplainWithStatus)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		s.Require().Empty(tss.Point{}, k.GetAccumulatedCommit(ctx, groupID, i))
	}

	s.Require().Equal(uint64(0), k.GetRound1DataCount(ctx, groupID))
	s.Require().Equal(uint64(0), k.GetRound2DataCount(ctx, groupID))
	s.Require().Equal(uint64(0), k.GetConfirmComplainCount(ctx, groupID))
}

func (s *KeeperTestSuite) TestGetSetDEQueue() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	deQueue := types.DEQueue{
		Head: 1,
		Tail: 2,
	}

	// Set DEQueue
	k.SetDEQueue(ctx, address, deQueue)

	// Get DEQueue
	got := k.GetDEQueue(ctx, address)

	s.Require().Equal(deQueue, got)
}

func (s *KeeperTestSuite) TestGetSetDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	index := uint64(1)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, index, de)

	// Get DE
	got, err := k.GetDE(ctx, address, index)

	s.Require().NoError(err)
	s.Require().Equal(de, got)
}

func (s *KeeperTestSuite) TestDeleteDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	index := uint64(1)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, index, de)

	// Get DE
	k.DeleteDE(ctx, address, index)

	// Try to get the deleted DE
	got, err := k.GetDE(ctx, address, index)

	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, got)
}

func (s *KeeperTestSuite) TestHandleSetDEs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}

	// Handle setting DEs
	k.HandleSetDEs(ctx, address, des)

	// Get DEQueue
	deQueue := k.GetDEQueue(ctx, address)

	// Check that all DEs have been stored correctly
	s.Require().Equal(uint64(len(des)), deQueue.Tail)
	for i := uint64(0); i < deQueue.Tail; i++ {
		gotDE, err := k.GetDE(ctx, address, i)
		s.Require().NoError(err)
		s.Require().Equal(des[i], gotDE)
	}
}

func (s *KeeperTestSuite) TestPollDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}
	index := uint64(1)

	// Set DE and DEQueue
	k.HandleSetDEs(ctx, address, des)

	// Poll DE
	polledDE, err := k.PollDE(ctx, address)
	s.Require().NoError(err)

	// Ensure polled DE is equal to original DE
	s.Require().Equal(des[0], polledDE)

	// Attempt to get deleted DE
	deletedDE, err := k.GetDE(ctx, address, index)

	// Should return error
	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, deletedDE)
}

func (s *KeeperTestSuite) TestGetSetSigningCount() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Set signing count
	count := uint64(42)
	k.SetSigningCount(ctx, count)

	// Get signing count
	got := k.GetSigningCount(ctx)

	// Assert equality
	s.Require().Equal(count, got)
}

func (s *KeeperTestSuite) TestGetNextSigningID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Get initial signing count
	initialCount := k.GetSigningCount(ctx)

	// Get next signing ID
	signingID := k.GetNextSigningID(ctx)

	// Get updated signing count
	updatedCount := k.GetSigningCount(ctx)

	// Assert that the signing ID is incremented and the signing count is updated
	s.Require().Equal(tss.SigningID(initialCount+1), signingID)
	s.Require().Equal(initialCount+1, updatedCount)
}

func (s *KeeperTestSuite) TestGetSetSigning() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample signing object
	signingID := tss.SigningID(1)
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	signing := types.Signing{
		SigningID: signingID,
		GroupID:   groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: member1,
				Member:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
		Sig:           nil,
	}

	// Set signing
	k.SetSigning(ctx, signing)

	// Get signing
	got, err := k.GetSigning(ctx, signingID)

	// Assert no error and equality
	s.Require().NoError(err)
	s.Require().Equal(signing, got)
}

func (s *KeeperTestSuite) TestAddSigning() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample signing object
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	signing := types.Signing{
		GroupID: groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: member1,
				Member:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
		Sig:           nil,
	}

	// Add signing
	signingID := k.AddSigning(ctx, signing)

	// Get added signing
	got, err := k.GetSigning(ctx, signingID)

	// Assert no error and equality
	s.Require().NoError(err)
	s.Require().Equal(signingID, got.SigningID)
}

func (s *KeeperTestSuite) TestDeleteSigning() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample signing object
	signingID := tss.SigningID(1)
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	signing := types.Signing{
		SigningID: signingID,
		GroupID:   groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: member1,
				Member:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
		Sig:           nil,
	}

	// Set signing
	k.SetSigning(ctx, signing)

	// Delete the signing
	k.DeleteSigning(ctx, signingID)

	// Verify that the signing data is deleted
	_, err := k.GetSigning(ctx, signingID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetSetPendingSign() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	signingID := tss.SigningID(1)

	// Set PendingSign
	k.SetPendingSign(ctx, address, signingID)

	// Get PendingSign
	got := k.GetPendingSign(ctx, address, signingID)

	s.Require().True(got)
}

func (s *KeeperTestSuite) TestDeletePendingSign() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	signingID := tss.SigningID(1)

	// Set PendingSign
	k.SetPendingSign(ctx, address, signingID)

	// Confirm PendingSign was set
	got := k.GetPendingSign(ctx, address, signingID)
	s.Require().True(got)

	// Delete PendingSign
	k.DeletePendingSign(ctx, address, signingID)

	// Confirm PendingSign was deleted
	got = k.GetPendingSign(ctx, address, signingID)
	s.Require().False(got)
}

func (s *KeeperTestSuite) TestGetPendingSignIDs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	signingIDs := []tss.SigningID{1, 2, 3}

	// Set PendingSign for multiple SigningIDs
	for _, id := range signingIDs {
		k.SetPendingSign(ctx, address, id)
	}

	// Get all PendingSignIDs
	got := k.GetPendingSignIDs(ctx, address)

	// Convert got (which is []uint64) to []tss.SigningID for comparison
	var gotConverted []tss.SigningID
	for _, id := range got {
		gotConverted = append(gotConverted, tss.SigningID(id))
	}

	// Check if the returned IDs are equal to the ones we set
	s.Require().Equal(signingIDs, gotConverted)
}

func (s *KeeperTestSuite) TestSetGetSigCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)

	// Set initial SigCount
	initialCount := uint64(5)
	k.SetSigCount(ctx, signingID, initialCount)

	// Get and check SigCount
	gotCount := k.GetSigCount(ctx, signingID)
	s.Require().Equal(initialCount, gotCount)
}

func (s *KeeperTestSuite) TestAddSigCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)

	// Set initial SigCount
	initialCount := uint64(5)
	k.SetSigCount(ctx, signingID, initialCount)

	// Add to SigCount
	k.AddSigCount(ctx, signingID)

	// Get and check incremented SigCount
	gotCount := k.GetSigCount(ctx, signingID)
	s.Require().Equal(initialCount+1, gotCount)
}

func (s *KeeperTestSuite) TestDeleteSigCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)

	// Set initial SigCount
	initialCount := uint64(5)
	k.SetSigCount(ctx, signingID, initialCount)

	// Delete SigCount
	k.DeleteSigCount(ctx, signingID)

	// Get and check SigCount after deletion
	gotCount := k.GetSigCount(ctx, signingID)
	s.Require().Equal(uint64(0), gotCount) // usually, Get on a non-existing key will return the zero value of the type
}

func (s *KeeperTestSuite) TestGetSetPartialSig() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberID := tss.MemberID(1)
	sig := tss.Signature("sample-signature")

	// Set PartialSig
	k.SetPartialSig(ctx, signingID, memberID, sig)

	// Get and check PartialSig
	gotSig, err := k.GetPartialSig(ctx, signingID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(sig, gotSig)
}

func (s *KeeperTestSuite) TestDeletePartialSig() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberID := tss.MemberID(1)
	sig := tss.Signature("sample-signature")

	// Set PartialSig
	k.SetPartialSig(ctx, signingID, memberID, sig)

	// Delete PartialSig
	k.DeletePartialSig(ctx, signingID, memberID)

	// Try to get the deleted PartialSig, expecting an error
	_, err := k.GetPartialSig(ctx, signingID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetPartialSigs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberIDs := []tss.MemberID{1, 2, 3}
	sigs := tss.Signatures{
		tss.Signature("sample-signature-1"),
		tss.Signature("sample-signature-2"),
		tss.Signature("sample-signature-3"),
	}

	// Set PartialSigs
	for i, memberID := range memberIDs {
		k.SetPartialSig(ctx, signingID, memberID, sigs[i])
	}

	// Get all PartialSigs
	got := k.GetPartialSigs(ctx, signingID)

	// Check if the returned signatures are equal to the ones we set
	s.Require().ElementsMatch(sigs, got)
}

func (s *KeeperTestSuite) TestGetPartialSigsWithKey() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberIDs := []tss.MemberID{1, 2, 3}
	sigs := tss.Signatures{
		tss.Signature("sample-signature-1"),
		tss.Signature("sample-signature-2"),
		tss.Signature("sample-signature-3"),
	}

	// Set PartialSigs
	for i, memberID := range memberIDs {
		k.SetPartialSig(ctx, signingID, memberID, sigs[i])
	}

	// Get all PartialSigs with keys
	got := k.GetPartialSigsWithKey(ctx, signingID)

	// Construct expected result
	expected := []types.PartialSig{}
	for i, memberID := range memberIDs {
		expected = append(expected, types.PartialSig{
			MemberID: memberID,
			Sig:      sigs[i],
		})
	}

	// Check if the returned signatures with keys are equal to the ones we set
	s.Require().ElementsMatch(expected, got)
}

func (s *KeeperTestSuite) TestGetSetRollingSeed() {
	ctx, k := s.ctx, s.app.TSSKeeper
	rollingSeed := []byte("sample-rolling-seed")

	// Set RollingSeed
	k.SetRollingSeed(ctx, rollingSeed)

	// Get and check RollingSeed
	gotSeed := k.GetRollingSeed(ctx)
	s.Require().Equal(rollingSeed, gotSeed)
}

func (s *KeeperTestSuite) TestGetRandomAssigningParticipants() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := uint64(1)
	size := uint64(10)
	t := uint64(5)

	// Set RollingSeed
	k.SetRollingSeed(ctx, []byte("sample-rolling-seed"))

	// Generate random participants
	participants, err := k.GetRandomAssigningParticipants(ctx, signingID, size, t)
	s.Require().NoError(err)

	// Check that the number of participants is correct
	s.Require().Len(participants, int(t))

	// Check that there are no duplicate participants
	participantSet := make(map[tss.MemberID]struct{})
	for _, participant := range participants {
		_, exists := participantSet[participant]
		s.Require().False(exists)
		participantSet[participant] = struct{}{}
	}

	// Check that if use same block and rolling seed will got same answer
	s.Require().Equal([]tss.MemberID{4, 8, 6, 2, 5}, participants)

	// Test that it returns an error if t > size
	_, err = k.GetRandomAssigningParticipants(ctx, signingID, t-1, t)
	s.Require().Error(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
