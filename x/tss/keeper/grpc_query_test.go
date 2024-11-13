package keeper_test

import (
	"encoding/hex"
	"fmt"

	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGRPCQueryCounts() {
	ctx, q, k := s.ctx, s.queryServer, s.keeper

	res, err := q.Counts(ctx, &types.QueryCountsRequest{})
	s.Require().Nil(err)
	s.Require().Equal(k.GetGroupCount(ctx), res.GroupCount)
	s.Require().Equal(k.GetSigningCount(ctx), res.SigningCount)
}

func (s *KeeperTestSuite) TestGRPCQueryGroup() {
	ctx, q, k := s.ctx, s.queryServer, s.keeper

	addrStrs := []string{
		"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
		"band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9",
		"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	}
	memberAddrs := make([]sdk.AccAddress, 0, len(addrStrs))
	members := make([]types.Member, 0, len(addrStrs))

	for i, addr := range addrStrs {
		address := sdk.MustAccAddressFromBech32(addr)
		memberAddrs = append(memberAddrs, address)
		m := types.Member{
			ID:          tss.MemberID(i + 1),
			GroupID:     1,
			Address:     addr,
			PubKey:      nil,
			IsMalicious: false,
			IsActive:    true,
		}
		members = append(members, m)

		err := k.EnqueueDEs(ctx, address, []types.DE{
			{
				PubD: testutil.HexDecode("dddd"),
				PubE: testutil.HexDecode("eeee"),
			},
		})
		s.Require().NoError(err)
	}

	groupID, err := k.CreateGroup(ctx, memberAddrs, 3, "bandtss")
	s.Require().NoError(err)

	// Add round 1 infos
	round1Infos := make([]types.Round1Info, 0, len(members))
	for i := range members {
		round1Info := newMockRound1Info(tss.MemberID(i))
		round1Infos = append(round1Infos, round1Info)

		k.AddRound1Info(ctx, groupID, round1Info)
	}

	// Add round 2 infos
	round2Infos := make([]types.Round2Info, 0, len(members))
	for i := range members {
		round2Info := newMockRound2Info(tss.MemberID(i))
		round2Infos = append(round2Infos, round2Info)

		k.AddRound2Info(ctx, groupID, round2Info)
	}

	// Add complains
	complaintWithStatus1 := newMockComplaintsWithStatus(1, 2)
	complaintWithStatus2 := newMockComplaintsWithStatus(2, 1)
	k.AddComplaintsWithStatus(ctx, groupID, complaintWithStatus1)
	k.AddComplaintsWithStatus(ctx, groupID, complaintWithStatus2)

	// Add confirms
	confirm1 := types.Confirm{MemberID: 3, OwnPubKeySig: []byte("own_pub_key_sig")}
	confirm2 := types.Confirm{MemberID: 4, OwnPubKeySig: []byte("own_pub_key_sig")}
	confirm3 := types.Confirm{MemberID: 5, OwnPubKeySig: []byte("own_pub_key_sig")}
	k.AddConfirm(ctx, groupID, confirm1)
	k.AddConfirm(ctx, groupID, confirm2)
	k.AddConfirm(ctx, groupID, confirm3)

	dkgContextB, _ := hex.DecodeString("B1DA723010D6FE6199670F31390445B0BDD9A5122EB7B30C7764AF112A2A4F78")
	expectedGroup := types.GroupResult{
		Group: types.Group{
			ID:          1,
			Size_:       5,
			Threshold:   3,
			PubKey:      nil,
			Status:      types.GROUP_STATUS_ROUND_1,
			ModuleOwner: bandtsstypes.ModuleName,
		},
		DKGContext:           dkgContextB,
		Members:              members,
		Round1Infos:          round1Infos,
		Round2Infos:          round2Infos,
		ComplaintsWithStatus: []types.ComplaintsWithStatus{complaintWithStatus1, complaintWithStatus2},
		Confirms:             []types.Confirm{confirm1, confirm2, confirm3},
	}

	testCases := []struct {
		name    string
		input   types.QueryGroupRequest
		expPass bool
		expOut  types.QueryGroupResponse
	}{
		{
			name:    "non existing group",
			input:   types.QueryGroupRequest{GroupId: 2},
			expPass: false,
		},
		{
			name:    "success",
			input:   types.QueryGroupRequest{GroupId: 1},
			expPass: true,
			expOut:  types.QueryGroupResponse{GroupResult: expectedGroup},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			res, err := q.Group(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryMembers() {
	ctx, q, k := s.ctx, s.queryServer, s.keeper
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

	testCases := []struct {
		name    string
		input   types.QueryMembersRequest
		expPass bool
		expOut  types.QueryMembersResponse
	}{
		{
			name:    "non existing member",
			input:   types.QueryMembersRequest{GroupId: 2},
			expPass: false,
		},
		{
			name:    "success",
			input:   types.QueryMembersRequest{GroupId: 1},
			expPass: true,
			expOut:  types.QueryMembersResponse{Members: members},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			res, err := q.Members(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryIsGrantee() {
	ctx, q, authzKeeper := s.ctx, s.queryServer, s.authzKeeper

	// Init grantee and grantee address
	grantee := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	granter := sdk.MustAccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	someone := sdk.MustAccAddressFromBech32("band1s3k4330ps8gj3dkw8x77ug0qf50ff6vqdmwax9")

	genericAuthz := authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgSubmitDKGRound1{}))

	authzKeeper.EXPECT().
		GetAuthorization(gomock.Any(), grantee, granter, gomock.Any()).
		Return(genericAuthz, nil).
		AnyTimes()

	authzKeeper.EXPECT().
		GetAuthorization(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()

	testCases := []struct {
		name     string
		input    types.QueryIsGranteeRequest
		expPass  bool
		expOut   types.QueryIsGranteeResponse
		postTest func(res *types.QueryIsGranteeResponse)
	}{
		{
			name: "address is not bech32",
			input: types.QueryIsGranteeRequest{
				Granter: "asdsd1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				Grantee: "padads40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			},
			expPass: false,
		},
		{
			name: "address is empty",
			input: types.QueryIsGranteeRequest{
				Granter: "",
				Grantee: "",
			},
			expPass: false,
		},
		{
			name: "grantee address is not grant by granter",
			input: types.QueryIsGranteeRequest{
				Granter: granter.String(),
				Grantee: someone.String(),
			},
			expPass: true,
			expOut: types.QueryIsGranteeResponse{
				IsGrantee: false,
			},
		},
		{
			name: "success",
			input: types.QueryIsGranteeRequest{
				Granter: granter.String(),
				Grantee: grantee.String(),
			},
			expPass: true,
			expOut: types.QueryIsGranteeResponse{
				IsGrantee: true,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			res, err := q.IsGrantee(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryDE() {
	ctx, q, k := s.ctx, s.queryServer, s.keeper
	acc1 := sdk.MustAccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	err := k.EnqueueDEs(ctx, acc1, []types.DE{
		{PubD: tss.Point("pubD1"), PubE: tss.Point("pubE1")},
		{PubD: tss.Point("pubD2"), PubE: tss.Point("pubE2")},
		{PubD: tss.Point("pubD3"), PubE: tss.Point("pubE3")},
		{PubD: tss.Point("pubD4"), PubE: tss.Point("pubE4")},
		{PubD: tss.Point("pubD5"), PubE: tss.Point("pubE5")},
	})
	s.Require().NoError(err)

	de, err := k.DequeueDE(ctx, acc1)
	s.Require().NoError(err)
	s.Require().Equal(types.DE{PubD: tss.Point("pubD1"), PubE: tss.Point("pubE1")}, de)

	testCases := []struct {
		name    string
		input   types.QueryDERequest
		expPass bool
		expOut  types.QueryDEResponse
	}{
		{
			name: "success",
			input: types.QueryDERequest{
				Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			},
			expPass: true,
			expOut: types.QueryDEResponse{
				DEs:        nil,
				Pagination: &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
		{
			name:    "success - multiple DEs",
			input:   types.QueryDERequest{Address: acc1.String()},
			expPass: true,
			expOut: types.QueryDEResponse{
				DEs: []types.DE{
					{PubD: tss.Point("pubD2"), PubE: tss.Point("pubE2")},
					{PubD: tss.Point("pubD3"), PubE: tss.Point("pubE3")},
					{PubD: tss.Point("pubD4"), PubE: tss.Point("pubE4")},
					{PubD: tss.Point("pubD5"), PubE: tss.Point("pubE5")},
				},
				Pagination: &query.PageResponse{NextKey: nil, Total: 4},
			},
		},
		{
			name: "success - multiple DEs query by key; not set countTotal",
			input: types.QueryDERequest{
				Address: acc1.String(),
				Pagination: &query.PageRequest{
					Key: sdk.Uint64ToBigEndian(uint64(2)),
				},
			},
			expPass: true,
			expOut: types.QueryDEResponse{
				DEs: []types.DE{
					{PubD: tss.Point("pubD3"), PubE: tss.Point("pubE3")},
					{PubD: tss.Point("pubD4"), PubE: tss.Point("pubE4")},
					{PubD: tss.Point("pubD5"), PubE: tss.Point("pubE5")},
				},
				Pagination: &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
		{
			name: "success - pagination query DEs",
			input: types.QueryDERequest{
				Address: acc1.String(),
				Pagination: &query.PageRequest{
					Offset: 1,
					Limit:  2,
				},
			},
			expPass: true,
			expOut: types.QueryDEResponse{
				DEs: []types.DE{
					{PubD: tss.Point("pubD3"), PubE: tss.Point("pubE3")},
					{PubD: tss.Point("pubD4"), PubE: tss.Point("pubE4")},
				},
				Pagination: &query.PageResponse{NextKey: sdk.Uint64ToBigEndian(4), Total: 0},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			res, err := q.DE(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryPendingGroups() {
	ctx, q, k := s.ctx, s.queryServer, s.keeper

	addrStrs := []string{
		"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
	}
	memberAddrs := make([]sdk.AccAddress, 0, len(addrStrs))

	for _, addr := range addrStrs {
		address := sdk.MustAccAddressFromBech32(addr)
		memberAddrs = append(memberAddrs, address)
	}

	groupID, err := k.CreateGroup(ctx, memberAddrs, 2, "bandtss")
	s.Require().NoError(err)

	testCases := []struct {
		name    string
		input   types.QueryPendingGroupsRequest
		expPass bool
		expOut  types.QueryPendingGroupsResponse
	}{
		{
			name: "success",
			input: types.QueryPendingGroupsRequest{
				Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			},
			expPass: true,
			expOut: types.QueryPendingGroupsResponse{
				PendingGroups: nil,
			},
		},
		{
			name: "success - with a pending group",
			input: types.QueryPendingGroupsRequest{
				Address: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
			},
			expPass: true,
			expOut: types.QueryPendingGroupsResponse{
				PendingGroups: []uint64{uint64(groupID)},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			res, err := q.PendingGroups(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryPendingSignings() {
	ctx, q := s.ctx, s.queryServer

	testCases := []struct {
		name    string
		input   types.QueryPendingSigningsRequest
		expPass bool
		expOut  types.QueryPendingSigningsResponse
	}{
		{
			name: "success",
			input: types.QueryPendingSigningsRequest{
				Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			},
			expPass: true,
			expOut: types.QueryPendingSigningsResponse{
				PendingSignings: nil,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			res, err := q.PendingSignings(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQuerySigning() {
	ctx, q, k := s.ctx, s.queryServer, s.keeper
	mockSig := []byte("signatures")

	s.rollingseedKeeper.EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("long_enough_rolling_seed")).
		AnyTimes()

	// Create group and signing
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

	testCases := []struct {
		name       string
		input      types.QuerySigningRequest
		preProcess func(s *KeeperTestSuite) error
		expPass    bool
		expOut     types.QuerySigningResponse
	}{
		{
			name: "invalid signing id",
			input: types.QuerySigningRequest{
				SigningId: 999,
			},
			expPass: false,
		},
		{
			name: "success",
			input: types.QuerySigningRequest{
				SigningId: 1,
			},
			expPass: true,
			expOut: types.QuerySigningResponse{
				SigningResult: types.SigningResult{
					Signing:               signing,
					CurrentSigningAttempt: &sa,
					EVMSignature:          nil,
					ReceivedPartialSignatures: []types.PartialSignature{
						{
							SigningID:      1,
							SigningAttempt: 1,
							MemberID:       2,
							Signature:      mockSig,
						},
					},
				},
			},
		},
		{
			name: "success but signing is completed for a while and some data is already removed",
			preProcess: func(s *KeeperTestSuite) error {
				k.DeleteInterimSigningData(ctx, 1, 1)
				return nil
			},
			input: types.QuerySigningRequest{
				SigningId: 1,
			},
			expPass: true,
			expOut: types.QuerySigningResponse{
				SigningResult: types.SigningResult{
					Signing:                   signing,
					CurrentSigningAttempt:     nil,
					EVMSignature:              nil,
					ReceivedPartialSignatures: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			if tc.preProcess != nil {
				err := tc.preProcess(s)
				s.Require().NoError(err)
			}

			res, err := q.Signing(ctx, &tc.input)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&tc.expOut, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
