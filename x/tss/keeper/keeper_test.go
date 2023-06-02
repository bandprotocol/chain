package keeper_test

import (
	"encoding/hex"
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
		Member:      "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
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
			Member:      "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
		},
		{
			Member:      "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
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
			Member:      "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
		},
		{
			Member:      "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
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
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
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
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
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
	groupID, member1, member2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
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
	groupID, count := tss.GroupID(1), uint64(5)

	// Set round 1 data count
	k.SetRound1DataCount(ctx, groupID, count)

	got := k.GetRound1DataCount(ctx, groupID)
	s.Require().Equal(uint64(5), got)
}

func (s *KeeperTestSuite) TestDeleteRound1DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, count := tss.GroupID(1), uint64(5)

	// Set round 1 data count
	k.SetRound1DataCount(ctx, groupID, count)

	// Delete round 1 data count
	k.DeleteRound1DataCount(ctx, groupID)

	got := k.GetRound1DataCount(ctx, groupID)
	s.Require().Empty(got)
}

func (s *KeeperTestSuite) TestGetSetRound2Data() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
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
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
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
	groupID, member1, member2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
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
	groupID, count := tss.GroupID(1), uint64(5)

	// Set round 2 data count
	k.SetRound2DataCount(ctx, groupID, count)

	got := k.GetRound2DataCount(ctx, groupID)
	s.Require().Equal(uint64(5), got)
}

func (s *KeeperTestSuite) TestDeleteRound2DataCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, count := tss.GroupID(1), uint64(5)

	// Set round 2 data count
	k.SetRound2DataCount(ctx, groupID, count)

	// Delete round 2 data count
	k.DeleteRound2DataCount(ctx, groupID)

	got := k.GetRound2DataCount(ctx, groupID)
	s.Require().Empty(got)
}

func (s *KeeperTestSuite) TestGetMaliciousMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	member1 := types.Member{
		Member:      "member_address_1",
		PubKey:      []byte("pub_key"),
		IsMalicious: true,
	}
	member2 := types.Member{
		Member:      "member_address_2",
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
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	privKeyI, _ := hex.DecodeString("7fc4175e7eb9661496cc38526f0eb4abccfd89d15f3371c3729e11c3ba1d6a14")
	pubKeyI, _ := hex.DecodeString("03936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa")
	pubKeyJ, _ := hex.DecodeString("03f70e80bac0b32b2599fa54d83b5471e90fac27bb09528f0337b49d464d64426f")
	member1 := types.Member{
		Member:      "member_address_1",
		PubKey:      pubKeyI,
		IsMalicious: false,
	}
	member2 := types.Member{
		Member:      "member_address_2",
		PubKey:      pubKeyJ,
		IsMalicious: false,
	}

	// Set member
	k.SetMember(ctx, groupID, memberID1, member1)
	k.SetMember(ctx, groupID, memberID2, member2)

	// Sign
	sig, keySym, nonceSym, err := tss.SignComplain(pubKeyI, pubKeyJ, privKeyI)
	s.Require().NoError(err)

	err = k.HandleVerifyComplainSig(ctx, groupID, types.Complain{
		I:         memberID1,
		J:         memberID2,
		KeySym:    keySym,
		Signature: sig,
		NonceSym:  nonceSym,
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleVerifyOwnPubKeySig() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	dkgContext, _ := hex.DecodeString("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c")
	pubKey, _ := hex.DecodeString("03936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa")
	privKey, _ := hex.DecodeString("7fc4175e7eb9661496cc38526f0eb4abccfd89d15f3371c3729e11c3ba1d6a14")
	member := types.Member{
		Member:      "member_address",
		PubKey:      pubKey,
		IsMalicious: false,
	}

	// Set member
	k.SetMember(ctx, groupID, memberID, member)

	// Set dkg context
	k.SetDKGContext(ctx, groupID, dkgContext)

	// Sign
	sig, err := tss.SignOwnPublickey(memberID, dkgContext, pubKey, privKey)
	s.Require().NoError(err)

	err = k.HandleVerifyOwnPubKeySig(ctx, groupID, memberID, sig)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestHandleComputeGroupPublicKey() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	point1, _ := hex.DecodeString("023487463ba3c7dbf9de9dc5bc73393f99ba0d86270ce2e4218d60e4a01d8cd11c")
	point2, _ := hex.DecodeString("03bd0e1b7c880ce80d4340540240972522b44bba2afcf50bfbe30e0352f225eba9")

	// Set round 1 data
	k.SetRound1Data(ctx, groupID, types.Round1Data{
		MemberID:           memberID1,
		CoefficientsCommit: tss.Points{point1},
	})
	k.SetRound1Data(ctx, groupID, types.Round1Data{
		MemberID:           memberID2,
		CoefficientsCommit: tss.Points{point2},
	})

	pubKey, err := k.HandleComputeGroupPublicKey(ctx, groupID)
	s.Require().NoError(err)
	s.Require().
		Equal("023704dcdb774ed4fd0841ded5757211fe5a6f7637c4f9a1346b5b20e2524d12e5", hex.EncodeToString(pubKey))
}

func (s *KeeperTestSuite) TestGetSetComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	complainWithStatus := types.ComplainsWithStatus{
		MemberID: memberID,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:         1,
					J:         2,
					KeySym:    []byte("key_sym"),
					Signature: []byte("signature"),
					NonceSym:  []byte("nonce_sym"),
				},
				ComplainStatus: types.SUCCESS,
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
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	complainWithStatus := types.ComplainsWithStatus{
		MemberID: memberID,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:         1,
					J:         2,
					KeySym:    []byte("key_sym"),
					Signature: []byte("signature"),
					NonceSym:  []byte("nonce_sym"),
				},
				ComplainStatus: types.SUCCESS,
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
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	complainWithStatus1 := types.ComplainsWithStatus{
		MemberID: memberID1,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:         1,
					J:         2,
					KeySym:    []byte("key_sym"),
					Signature: []byte("signature"),
					NonceSym:  []byte("nonce_sym"),
				},
				ComplainStatus: types.SUCCESS,
			},
		},
	}
	complainWithStatus2 := types.ComplainsWithStatus{
		MemberID: memberID2,
		ComplainsWithStatus: []types.ComplainWithStatus{
			{
				Complain: types.Complain{
					I:         1,
					J:         2,
					KeySym:    []byte("key_sym"),
					Signature: []byte("signature"),
					NonceSym:  []byte("nonce_sym"),
				},
				ComplainStatus: types.SUCCESS,
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
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
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
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
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
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
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
	groupID, count := tss.GroupID(1), uint64(1)

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
	groupID, count := tss.GroupID(1), uint64(5)

	// Set confirm complain count
	k.SetConfirmComplainCount(ctx, groupID, count)

	// Delete confirm complain count
	k.DeleteConfirmComplainCount(ctx, groupID)

	got := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Empty(got)
}

func (s *KeeperTestSuite) TestMarkMalicious() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)

	// Set member
	k.SetMember(ctx, groupID, memberID, types.Member{
		Member:      "member_address",
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
			Member:      "member_address",
			PubKey:      []byte("pub_key"),
			IsMalicious: true,
		},
	}, got)
}

func (s *KeeperTestSuite) TestGetRandomAssigningParticipants() {
	ctx, k := s.ctx, s.app.TSSKeeper

	got, err := k.GetRandomAssigningParticipants(ctx, 1, 5, 3)
	s.Require().NoError(err)
	s.Require().Equal([]tss.MemberID{4, 3, 5}, got)
}

func (s *KeeperTestSuite) TestGetPendingSignIDs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	member, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	k.SetPendingSign(ctx, member, 1)
	k.SetPendingSign(ctx, member, 2)
	k.SetPendingSign(ctx, member, 5)

	k.DeletePendingSign(ctx, member, 5)

	got := k.GetPendingSignIDs(ctx, member)
	s.Require().Equal([]uint64{1, 2}, got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
