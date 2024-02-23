package keeper_test

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtsskeeper "github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGRPCQueryCounts() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper

	res, err := q.Counts(ctx, &types.QueryCountsRequest{})
	s.Require().Nil(err)
	s.Require().Equal(k.GetGroupCount(ctx), res.GroupCount)
	s.Require().Equal(k.GetSigningCount(ctx), res.SigningCount)
	s.Require().Equal(k.GetReplacementCount(ctx), res.ReplacementCount)
}

func (s *KeeperTestSuite) TestGRPCQueryGroup() {
	ctx, q, k, bandTSSKeeper := s.ctx, s.queryClient, s.app.TSSKeeper, s.app.BandTSSKeeper
	bandTSSMsgSrvr := bandtsskeeper.NewMsgServerImpl(s.app.BandTSSKeeper)
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)

	members := []string{
		"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
		"band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9",
		"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	}

	for _, m := range members {
		address := sdk.MustAccAddressFromBech32(m)
		err := bandTSSKeeper.SetActiveStatus(ctx, sdk.MustAccAddressFromBech32(m))
		s.Require().NoError(err)

		err = k.HandleSetDEs(ctx, address, []types.DE{
			{
				PubD: testutil.HexDecode("dddd"),
				PubE: testutil.HexDecode("eeee"),
			},
		})
		s.Require().NoError(err)
	}

	round1Info1 := types.Round1Info{
		MemberID: memberID1,
		CoefficientCommits: []tss.Point{
			[]byte("point1"),
			[]byte("point2"),
			[]byte("point3"),
		},
		OneTimePubKey:    []byte("OneTimePubKeySample"),
		A0Signature:      []byte("A0SignatureSample"),
		OneTimeSignature: []byte("OneTimeSignatureSample"),
	}
	round1Info2 := types.Round1Info{
		MemberID: memberID2,
		CoefficientCommits: []tss.Point{
			[]byte("point1"),
			[]byte("point2"),
			[]byte("point3"),
		},
		OneTimePubKey:    []byte("OneTimePubKeySample"),
		A0Signature:      []byte("A0SignatureSample"),
		OneTimeSignature: []byte("OneTimeSignatureSample"),
	}
	round2Info1 := types.Round2Info{
		MemberID: memberID1,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("secret1"),
			[]byte("secret2"),
		},
	}
	round2Info2 := types.Round2Info{
		MemberID: memberID2,
		EncryptedSecretShares: tss.EncSecretShares{
			[]byte("secret1"),
			[]byte("secret2"),
		},
	}
	complaintWithStatus1 := types.ComplaintsWithStatus{
		MemberID: memberID1,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainant: 1,
					Respondent:  2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}
	complaintWithStatus2 := types.ComplaintsWithStatus{
		MemberID: memberID2,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainant: 1,
					Respondent:  2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}
	confirm1 := types.Confirm{
		MemberID:     memberID1,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}
	confirm2 := types.Confirm{
		MemberID:     memberID2,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	_, err := bandTSSMsgSrvr.CreateGroup(ctx, &bandtsstypes.MsgCreateGroup{
		Members:   members,
		Threshold: 3,
		Authority: s.authority.String(),
	})
	s.Require().NoError(err)

	// Add round 1 infos
	k.AddRound1Info(ctx, groupID, round1Info1)
	k.AddRound1Info(ctx, groupID, round1Info2)

	// Add round 2 infos
	k.AddRound2Info(ctx, groupID, round2Info1)
	k.AddRound2Info(ctx, groupID, round2Info2)

	// Add complains
	k.AddComplaintsWithStatus(ctx, groupID, complaintWithStatus1)
	k.AddComplaintsWithStatus(ctx, groupID, complaintWithStatus2)

	// Add confirms
	k.AddConfirm(ctx, groupID, confirm1)
	k.AddConfirm(ctx, groupID, confirm2)

	var req types.QueryGroupRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryGroupResponse)
	}{
		{
			"non existing group",
			func() {
				req = types.QueryGroupRequest{
					GroupId: 2,
				}
			},
			false,
			func(res *types.QueryGroupResponse) {},
		},
		{
			"success",
			func() {
				req = types.QueryGroupRequest{
					GroupId: uint64(groupID),
				}
			},
			true,
			func(res *types.QueryGroupResponse) {
				dkgContextB, _ := hex.DecodeString("6c31fc15422ebad28aaf9089c306702f67540b53c7eea8b7d2941044b027100f")

				expectedMemberStatuses := []bandtsstypes.Status{
					{
						Address: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
						Status:  bandtsstypes.MEMBER_STATUS_ACTIVE,
						Since:   ctx.BlockTime(),
					},
					{
						Address: "band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
						Status:  bandtsstypes.MEMBER_STATUS_ACTIVE,
						Since:   ctx.BlockTime(),
					},
					{
						Address: "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
						Status:  bandtsstypes.MEMBER_STATUS_ACTIVE,
						Since:   ctx.BlockTime(),
					},
					{
						Address: "band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9",
						Status:  bandtsstypes.MEMBER_STATUS_ACTIVE,
						Since:   ctx.BlockTime(),
					},
					{
						Address: "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
						Status:  bandtsstypes.MEMBER_STATUS_ACTIVE,
						Since:   ctx.BlockTime(),
					},
				}

				for _, expectedStatus := range expectedMemberStatuses {
					status := bandTSSKeeper.GetStatus(ctx, sdk.MustAccAddressFromBech32(expectedStatus.Address))
					s.Require().Equal(status, expectedStatus)
				}

				s.Require().Equal(&types.QueryGroupResponse{
					Group: types.Group{
						ID:        1,
						Size_:     5,
						Threshold: 3,
						PubKey:    nil,
						Status:    types.GROUP_STATUS_ROUND_1,
					},
					DKGContext: dkgContextB,
					Members: []types.Member{
						{
							ID:          1,
							GroupID:     1,
							Address:     "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
							PubKey:      nil,
							IsMalicious: false,
						},
						{
							ID:          2,
							GroupID:     1,
							Address:     "band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
							PubKey:      nil,
							IsMalicious: false,
						},
						{
							ID:          3,
							GroupID:     1,
							Address:     "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
							PubKey:      nil,
							IsMalicious: false,
						},
						{
							ID:          4,
							GroupID:     1,
							Address:     "band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9",
							PubKey:      nil,
							IsMalicious: false,
						},
						{
							ID:          5,
							GroupID:     1,
							Address:     "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
							PubKey:      nil,
							IsMalicious: false,
						},
					},
					IsActives: []bool{true, true, true, true, true},
					Round1Infos: []types.Round1Info{
						round1Info1,
						round1Info2,
					},
					Round2Infos: []types.Round2Info{
						round2Info1,
						round2Info2,
					},
					ComplaintsWithStatus: []types.ComplaintsWithStatus{
						complaintWithStatus1,
						complaintWithStatus2,
					},
					Confirms: []types.Confirm{
						confirm1,
						confirm2,
					},
				}, res)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.Group(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res)
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryMembers() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper
	members := []types.Member{
		{
			ID:          1,
			GroupID:     1,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			ID:          2,
			GroupID:     1,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}

	// Set members
	for _, m := range members {
		k.SetMember(ctx, m)
	}

	var req types.QueryMembersRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryMembersResponse)
	}{
		{
			"non existing member",
			func() {
				req = types.QueryMembersRequest{
					GroupId: 2,
				}
			},
			false,
			func(res *types.QueryMembersResponse) {},
		},
		{
			"success",
			func() {
				req = types.QueryMembersRequest{
					GroupId: 1,
				}
			},
			true,
			func(res *types.QueryMembersResponse) {
				s.Require().Equal(members, res.Members)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := q.Members(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryIsGrantee() {
	ctx, q, authzKeeper := s.ctx, s.queryClient, s.app.AuthzKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.GetTSSGrantMsgTypes() {
		err := authzKeeper.SaveGrant(s.ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
		s.Require().NoError(err)
	}

	var req types.QueryIsGranteeRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryIsGranteeResponse)
	}{
		{
			"address is not bech32",
			func() {
				req = types.QueryIsGranteeRequest{
					Granter: "asdsd1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					Grantee: "padads40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
				}
			},
			false,
			func(res *types.QueryIsGranteeResponse) {},
		},
		{
			"address is empty",
			func() {
				req = types.QueryIsGranteeRequest{
					Granter: "",
					Grantee: "",
				}
			},
			false,
			func(res *types.QueryIsGranteeResponse) {},
		},
		{
			"grantee address is not grant by granter",
			func() {
				req = types.QueryIsGranteeRequest{
					Granter: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					Grantee: "band17eplw6tga7wqgruqdtalw3rky4njkx6vngxjlt",
				}
			},
			true,
			func(res *types.QueryIsGranteeResponse) {
				s.Require().False(res.IsGrantee)
			},
		},
		{
			"success",
			func() {
				req = types.QueryIsGranteeRequest{
					Granter: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					Grantee: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
				}
			},
			true,
			func(res *types.QueryIsGranteeResponse) {
				s.Require().True(res.IsGrantee)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := q.IsGrantee(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryDE() {
	ctx, q := s.ctx, s.queryClient

	var req types.QueryDERequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryDEResponse, err error)
	}{
		{
			"success",
			func() {
				req = types.QueryDERequest{
					Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			true,
			func(res *types.QueryDEResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.DEs, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.DE(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res, err)
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryPendingGroups() {
	ctx, q := s.ctx, s.queryClient

	var req types.QueryPendingGroupsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryPendingGroupsResponse, err error)
	}{
		{
			"success",
			func() {
				req = types.QueryPendingGroupsRequest{
					Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			true,
			func(res *types.QueryPendingGroupsResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.PendingGroups, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.PendingGroups(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res, err)
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryPendingSignings() {
	ctx, q := s.ctx, s.queryClient

	var req types.QueryPendingSigningsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryPendingSigningsResponse, err error)
	}{
		{
			"success",
			func() {
				req = types.QueryPendingSigningsRequest{
					Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			true,
			func(res *types.QueryPendingSigningsResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.PendingSignings, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.PendingSignings(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res, err)
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryReplacement() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper

	now := time.Now().UTC()

	// Create a replacement
	replacement := types.Replacement{
		ID:             1,
		SigningID:      1,
		CurrentGroupID: 1,
		NewGroupID:     2,
		CurrentPubKey:  []byte("test_pub_key"),
		NewPubKey:      []byte("test_pub_key"),
		Status:         types.REPLACEMENT_STATUS_WAITING,
		ExecTime:       now,
	}
	k.SetReplacement(ctx, replacement)

	var req types.QueryReplacementRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryReplacementResponse, err error)
	}{
		{
			"success",
			func() {
				req = types.QueryReplacementRequest{
					Id: replacement.ID,
				}
			},
			true,
			func(res *types.QueryReplacementResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Equal(replacement, *res.Replacement)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.Replacement(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res, err)
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryReplacements() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper

	// Create a replacement
	replacement := types.Replacement{
		ID:             1,
		SigningID:      1,
		CurrentGroupID: 1,
		NewGroupID:     2,
		CurrentPubKey:  []byte("test_pub_key"),
		NewPubKey:      []byte("test_pub_key"),
		Status:         types.REPLACEMENT_STATUS_WAITING,
		ExecTime:       time.Now(),
	}
	k.SetReplacement(ctx, replacement)

	var req types.QueryReplacementsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QueryReplacementsResponse, err error)
	}{
		{
			"success",
			func() {
				req = types.QueryReplacementsRequest{
					Status: types.REPLACEMENT_STATUS_WAITING,
				}
			},
			true,
			func(res *types.QueryReplacementsResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.Replacements, 1)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.Replacements(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res, err)
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQuerySigning() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper
	signingID, memberID, groupID := tss.SigningID(1), tss.MemberID(1), tss.GroupID(1)
	signing := types.Signing{
		ID:      signingID,
		GroupID: groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID:      memberID,
				Address:       "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:          []byte("D"),
				PubE:          []byte("E"),
				BindingFactor: []byte("binding_factor"),
				PubNonce:      []byte("public_nonce"),
			},
		},
		Message:       []byte("message"),
		GroupPubNonce: []byte("group_pub_nonce"),
		Signature: testutil.HexDecode(
			"02d447778a1a2cd2a55ceb47d6bd3f01587d079d6ddadbfff5d6956ca9b7ca0e317074433c8adbfb338cd69a343fc1155ce60d4f5e276975ba8bdb8ae8f803ea23",
		),
	}
	sig := []byte("signatures")

	// Add partial signature
	k.AddPartialSignature(ctx, signingID, memberID, []byte("signatures"))

	// Add signing
	k.AddSigning(ctx, signing)

	var req types.QuerySigningRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func(res *types.QuerySigningResponse, err error)
	}{
		{
			"invalid signing id",
			func() {
				req = types.QuerySigningRequest{
					SigningId: 999,
				}
			},
			false,
			func(res *types.QuerySigningResponse, err error) {
				s.Require().Error(err)
				s.Require().Nil(res)
			},
		},
		{
			"success",
			func() {
				req = types.QuerySigningRequest{
					SigningId: 1,
				}
			},
			true,
			func(res *types.QuerySigningResponse, err error) {
				s.Require().NoError(err)
				s.Require().Equal(signing, res.Signing)
				s.Require().
					Equal([]types.PartialSignature{{MemberID: memberID, Signature: sig}}, res.ReceivedPartialSignatures)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := q.Signing(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest(res, err)
		})
	}
}
