package keeper_test

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *AppTestSuite) TestGRPCQueryCounts() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper

	res, err := q.Counts(ctx, &types.QueryCountsRequest{})
	s.Require().Nil(err)
	s.Require().Equal(k.GetGroupCount(ctx), res.GroupCount)
	s.Require().Equal(k.GetSigningCount(ctx), res.SigningCount)
}

func (s *AppTestSuite) TestGRPCQueryGroup() {
	ctx, q, k, bandtssKeeper := s.ctx, s.queryClient, s.app.TSSKeeper, s.app.BandtssKeeper
	bandtssMsgSrvr := bandtsskeeper.NewMsgServerImpl(s.app.BandtssKeeper)
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
		err := k.HandleSetDEs(ctx, address, []types.DE{
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

	_, err := bandtssMsgSrvr.TransitionGroup(ctx, &bandtsstypes.MsgTransitionGroup{
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

	bandtssKeeper.SetCurrentGroupID(ctx, groupID)
	for _, m := range members {
		address := sdk.MustAccAddressFromBech32(m)
		err := bandtssKeeper.AddMember(ctx, address, groupID)
		s.Require().NoError(err)
	}

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
				dkgContextB, _ := hex.DecodeString("4C46F109ABC25187430BA6D2726210D3D81933952F6C9B900564F71D44D146A1")

				expectedMembers := []bandtsstypes.Member{
					{
						Address:  "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
						GroupID:  tss.GroupID(1),
						IsActive: true,
						Since:    ctx.BlockTime(),
					},
					{
						Address:  "band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
						GroupID:  tss.GroupID(1),
						IsActive: true,
						Since:    ctx.BlockTime(),
					},
					{
						Address:  "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
						GroupID:  tss.GroupID(1),
						IsActive: true,
						Since:    ctx.BlockTime(),
					},
					{
						Address:  "band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9",
						GroupID:  tss.GroupID(1),
						IsActive: true,
						Since:    ctx.BlockTime(),
					},
					{
						Address:  "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
						GroupID:  tss.GroupID(1),
						IsActive: true,
						Since:    ctx.BlockTime(),
					},
				}

				for _, expectedMember := range expectedMembers {
					member, err := bandtssKeeper.GetMember(
						ctx,
						sdk.MustAccAddressFromBech32(expectedMember.Address),
						tss.GroupID(1),
					)
					s.Require().NoError(err)
					s.Require().Equal(expectedMember, member)
				}

				groupResult := types.GroupResult{
					Group: types.Group{
						ID:          1,
						Size_:       5,
						Threshold:   3,
						PubKey:      nil,
						Status:      types.GROUP_STATUS_ROUND_1,
						ModuleOwner: bandtsstypes.ModuleName,
					},
					DKGContext: dkgContextB,
					Members: []types.Member{
						{
							ID:          1,
							GroupID:     1,
							Address:     "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
							PubKey:      nil,
							IsMalicious: false,
							IsActive:    true,
						},
						{
							ID:          2,
							GroupID:     1,
							Address:     "band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
							PubKey:      nil,
							IsMalicious: false,
							IsActive:    true,
						},
						{
							ID:          3,
							GroupID:     1,
							Address:     "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
							PubKey:      nil,
							IsMalicious: false,
							IsActive:    true,
						},
						{
							ID:          4,
							GroupID:     1,
							Address:     "band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9",
							PubKey:      nil,
							IsMalicious: false,
							IsActive:    true,
						},
						{
							ID:          5,
							GroupID:     1,
							Address:     "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
							PubKey:      nil,
							IsMalicious: false,
							IsActive:    true,
						},
					},
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
				}

				s.Require().Equal(&types.QueryGroupResponse{
					GroupResult: groupResult,
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

func (s *AppTestSuite) TestGRPCQueryMembers() {
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

func (s *AppTestSuite) TestGRPCQueryIsGrantee() {
	ctx, q, authzKeeper := s.ctx, s.queryClient, s.app.AuthzKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.GetGrantMsgTypes() {
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

func (s *AppTestSuite) TestGRPCQueryDE() {
	ctx, q := s.ctx, s.queryClient
	acc1 := sdk.MustAccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	err := s.app.TSSKeeper.HandleSetDEs(ctx, acc1, []types.DE{
		{PubD: tss.Point("pubD1"), PubE: tss.Point("pubE1")},
		{PubD: tss.Point("pubD2"), PubE: tss.Point("pubE2")},
		{PubD: tss.Point("pubD3"), PubE: tss.Point("pubE3")},
		{PubD: tss.Point("pubD4"), PubE: tss.Point("pubE4")},
		{PubD: tss.Point("pubD5"), PubE: tss.Point("pubE5")},
	})
	s.Require().NoError(err)

	de, err := s.app.TSSKeeper.PollDE(ctx, acc1)
	s.Require().NoError(err)
	s.Require().Equal(types.DE{PubD: tss.Point("pubD1"), PubE: tss.Point("pubE1")}, de)

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
		{
			"success - multiple DEs",
			func() {
				req = types.QueryDERequest{Address: acc1.String()}
			},
			true,
			func(res *types.QueryDEResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.DEs, 4)
			},
		},
		{
			"success - multiple DEs query by key",
			func() {
				req = types.QueryDERequest{
					Address: acc1.String(),
					Pagination: &query.PageRequest{
						Key: sdk.Uint64ToBigEndian(uint64(2)),
					},
				}
			},
			true,
			func(res *types.QueryDEResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.DEs, 3)
			},
		},
		{
			"success - pagination query DEs",
			func() {
				req = types.QueryDERequest{
					Address: acc1.String(),
					Pagination: &query.PageRequest{
						Offset: 1,
						Limit:  2,
					},
				}
			},
			true,
			func(res *types.QueryDEResponse, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Len(res.DEs, 2)
				s.Require().Equal(
					types.DE{PubD: tss.Point("pubD3"), PubE: tss.Point("pubE3")},
					res.DEs[0],
				)
				s.Require().Equal(
					types.DE{PubD: tss.Point("pubD4"), PubE: tss.Point("pubE4")},
					res.DEs[1],
				)
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

func (s *AppTestSuite) TestGRPCQueryPendingGroups() {
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

func (s *AppTestSuite) TestGRPCQueryPendingSignings() {
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

func (s *AppTestSuite) TestGRPCQuerySigning() {
	ctx, q, k := s.ctx, s.queryClient, s.app.TSSKeeper
	mockSig := []byte("signatures")

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	group, err := k.GetGroup(ctx, groupCtx.GroupID)
	s.Require().NoError(err)
	s.Require().Equal(types.GROUP_STATUS_ACTIVE, group.Status)

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	s.Require().NoError(err)
	err = k.InitiateNewSigningRound(ctx, signingID)
	s.Require().NoError(err)

	sa := k.MustGetSigningAttempt(ctx, signingID, 1)
	signing := k.MustGetSigning(ctx, signingID)
	memberID := sa.AssignedMembers[0].MemberID

	// Add partial signature
	k.AddPartialSignature(ctx, signingID, 1, memberID, mockSig)

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
				s.Require().Equal(signing, res.SigningResult.Signing)
				s.Require().Equal(&sa, res.SigningResult.CurrentSigningAttempt)
				s.Require().
					Equal([]types.PartialSignature{
						{
							SigningID:      1,
							SigningAttempt: 1,
							MemberID:       memberID,
							Signature:      mockSig,
						},
					}, res.SigningResult.ReceivedPartialSignatures)
			},
		},
		{
			"success but signing is completed for a while and some data is already removed",
			func() {
				k.DeleteInterimSigningData(ctx, 1, 1)
				req = types.QuerySigningRequest{
					SigningId: 1,
				}
			},
			true,
			func(res *types.QuerySigningResponse, err error) {
				s.Require().NoError(err)
				s.Require().Equal(signing, res.SigningResult.Signing)
				s.Require().Nil(res.SigningResult.CurrentSigningAttempt)
				s.Require().Nil(res.SigningResult.ReceivedPartialSignatures)

				// set it back
				k.SetSigningAttempt(ctx, sa)
				k.AddPartialSignature(ctx, signingID, 1, memberID, mockSig)
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
