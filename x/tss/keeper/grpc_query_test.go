package keeper_test

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

func (s *KeeperTestSuite) TestGRPCQueryGroup() {
	ctx, msgSrvr, q, k := s.ctx, s.msgSrvr, s.querier, s.app.TSSKeeper

	groupID := tss.GroupID(1)
	members := []string{
		"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
		"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	}
	round1DataMember1 := types.Round1Data{
		MemberID: 1,
		CoefficientsCommit: []tss.Point{
			[]byte("point1"),
			[]byte("point2"),
			[]byte("point3"),
		},
		OneTimePubKey: []byte("OneTimePubKeySample"),
		A0Sig:         []byte("A0SigSample"),
		OneTimeSig:    []byte("OneTimeSigSample"),
	}
	round1DataMember2 := types.Round1Data{
		MemberID: 2,
		CoefficientsCommit: []tss.Point{
			[]byte("point1"),
			[]byte("point2"),
			[]byte("point3"),
		},
		OneTimePubKey: []byte("OneTimePubKeySample"),
		A0Sig:         []byte("A0SigSample"),
		OneTimeSig:    []byte("OneTimeSigSample"),
	}
	round2DataMember1 := types.Round2Data{
		MemberID: 1,
		EncryptedSecretShares: tss.Scalars{
			[]byte("scalar1"),
			[]byte("scalar2"),
		},
	}
	round2DataMember2 := types.Round2Data{
		MemberID: 2,
		EncryptedSecretShares: tss.Scalars{
			[]byte("scalar1"),
			[]byte("scalar2"),
		},
	}

	msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members:   members,
		Threshold: 3,
		Sender:    members[0],
	})

	// set round1 data
	k.SetRound1Data(ctx, groupID, round1DataMember1)
	k.SetRound1Data(ctx, groupID, round1DataMember2)

	// set round 2 data
	k.SetRound2Data(ctx, groupID, round2DataMember1)
	k.SetRound2Data(ctx, groupID, round2DataMember2)

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
				dkgContextB, _ := hex.DecodeString("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c")

				s.Require().Equal(&types.QueryGroupResponse{
					Group: types.Group{
						Size_:     5,
						Threshold: 3,
						PubKey:    nil,
						Status:    types.ROUND_1,
					},
					DKGContext: dkgContextB,
					Members: []types.Member{
						{
							Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
							PubKey: tss.PublicKey(nil),
						},
						{
							Member: "band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
							PubKey: tss.PublicKey(nil),
						},
						{
							Member: "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
							PubKey: tss.PublicKey(nil),
						},
						{
							Member: "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
							PubKey: tss.PublicKey(nil),
						},
						{
							Member: "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
							PubKey: tss.PublicKey(nil),
						},
					},
					AllRound1Data: []types.Round1Data{
						round1DataMember1,
						round1DataMember2,
					},
					AllRound2Data: []types.Round2Data{
						round2DataMember1,
						round2DataMember2,
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
	ctx, q, k := s.ctx, s.querier, s.app.TSSKeeper
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
		k.SetMember(ctx, tss.GroupID(1), tss.MemberID(i+1), m)
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
	ctx, q, authzKeeper := s.ctx, s.querier, s.app.AuthzKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.MsgGrants {
		authzKeeper.SaveGrant(s.ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
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
