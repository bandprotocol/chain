package keeper_test

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestCreateGroupReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	s.Run("create group", func() {
		_, err := msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members: []string{
				"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
				"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
				"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
				"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
			},
			Threshold: 3,
			Sender:    "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
		})
		s.Require().NoError(err)
	})
}

func (s *KeeperTestSuite) TestSubmitDKGRound1Req() {
	ctx, msgSrvr, tssKeeper := s.ctx, s.msgSrvr, s.app.TSSKeeper
	cof1B, _ := hex.DecodeString("03373ab7ba39b7fbe5250990da1ef0414f9b8741604335dbae6a25b4f069a68259")
	cof2B, _ := hex.DecodeString("03348c23fa321dd1cf0791df247fb04424403ba244e359d80509bd645ce17f153e")
	cof3B, _ := hex.DecodeString("02979a0ac813d1d32499de36c52c8ffd1eb43846907860aba99e6a3759a04383b5")
	oneTimePubKeyB, _ := hex.DecodeString("039cba3c997f9755a67e7f7c182326a2a69bf1dfffb76eae7247a75d14dce8ee17")
	a0SigB, _ := hex.DecodeString(
		"035670c573f810b76ed14c89f0436b96db7f37e6da4beb6f9242b84f2be7e28b9675314a4e20df5c6395643879ee7a1cfba6758f15b6cb4a0689ddd0e5bb650051",
	)
	oneTimeSigB, _ := hex.DecodeString(
		"037dce67a68e450dbcb3aed3c0e6cd1bfbae73f637e47ecf9aabcb40703e390c1fdfb02476a300faee310dece8ca2a06fbd4e024b96064b1a610c4d0407706930e",
	)

	// create group for submit dkg context
	msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members: []string{
			"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
			"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
		},
		Threshold: 3,
		Sender:    "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	})

	var req types.MsgSubmitDKGRound1
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		{
			"group not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: 0,
					Round1Data: types.Round1Data{
						MemberID:           1,
						CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
						OneTimePubKey:      oneTimePubKeyB,
						A0Sig:              a0SigB,
						OneTimeSig:         oneTimeSigB,
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"member not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: 1,
					Round1Data: types.Round1Data{
						MemberID:           2,
						CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
						OneTimePubKey:      oneTimePubKeyB,
						A0Sig:              a0SigB,
						OneTimeSig:         oneTimeSigB,
					},
					Member: "band1rqjc6czdeu2w2nst9vfvv6yqj6nwqkv48s4jmq",
				}
			},
			false,
			func() {},
		},
		{
			"round 1 already commit",
			func() {
				// Set round 1 data
				tssKeeper.SetRound1Data(ctx, 1, types.Round1Data{
					MemberID:           1,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
				})

				req = types.MsgSubmitDKGRound1{
					GroupID: 1,
					Round1Data: types.Round1Data{
						MemberID:           1,
						CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
						OneTimePubKey:      oneTimePubKeyB,
						A0Sig:              a0SigB,
						OneTimeSig:         oneTimeSigB,
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {
				tssKeeper.DeleteRound1Data(ctx, 1, 1)
			},
		},
		{
			"wrong one time sign",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: 1,
					Round1Data: types.Round1Data{
						MemberID:           1,
						CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
						OneTimePubKey:      oneTimePubKeyB,
						A0Sig:              a0SigB,
						OneTimeSig:         []byte("wrong one_time_sign"),
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"wrong a0 sig",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: 1,
					Round1Data: types.Round1Data{
						MemberID:           1,
						CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
						OneTimePubKey:      oneTimePubKeyB,
						A0Sig:              []byte("wrong a0_sig"),
						OneTimeSig:         oneTimeSigB,
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"success",
			func() {
				// Key generated from GenerateRound1Data() ref. github.com/bandprotocol/chain/v2/pkg/tss
				req = types.MsgSubmitDKGRound1{
					GroupID: 1,
					Round1Data: types.Round1Data{
						MemberID:           1,
						CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
						OneTimePubKey:      oneTimePubKeyB,
						A0Sig:              a0SigB,
						OneTimeSig:         oneTimeSigB,
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.SubmitDKGRound1(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func (s *KeeperTestSuite) TestSubmitDKGRound2Req() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// create group for submit dkg context
	msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members: []string{
			"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
			"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
		},
		Threshold: 3,
		Sender:    "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	})
	k.SetGroup(ctx, types.Group{
		GroupID:   1,
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_2,
	})

	var req types.MsgSubmitDKGRound2
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		{
			"group not found",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 0,
					Round2Data: types.Round2Data{
						MemberID: 1,
						EncryptedSecretShares: tss.Scalars{
							[]byte("e_12"),
							[]byte("e_13"),
							[]byte("e_14"),
							[]byte("e_15"),
						},
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"member not authorized",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Data: types.Round2Data{
						MemberID: 10,
						EncryptedSecretShares: tss.Scalars{
							[]byte("e_12"),
							[]byte("e_13"),
							[]byte("e_14"),
							[]byte("e_15"),
						},
					},
					Member: "band1rqjc6czdeu2w2nst9vfvv6yqj6nwqkv48s4jmq",
				}
			},
			false,
			func() {},
		},
		{
			"round 2 already submit",
			func() {
				// Set round 2 data
				k.SetRound2Data(ctx, 1, types.Round2Data{
					MemberID: 1,
					EncryptedSecretShares: tss.Scalars{
						[]byte("e_12"),
						[]byte("e_13"),
						[]byte("e_14"),
						[]byte("e_15"),
					}})

				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Data: types.Round2Data{
						MemberID: 1,
						EncryptedSecretShares: tss.Scalars{
							[]byte("e_12"),
							[]byte("e_13"),
							[]byte("e_14"),
							[]byte("e_15"),
						},
					},
					Member: "band1rqjc6czdeu2w2nst9vfvv6yqj6nwqkv48s4jmq",
				}
			},
			false,
			func() {
				k.DeleteRound2Data(ctx, 1, 1)
			},
		},
		{
			"round 2 data is not correct length n-1",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Data: types.Round2Data{
						MemberID: 1,
						EncryptedSecretShares: tss.Scalars{
							[]byte("e_12"),
							[]byte("e_13"),
							[]byte("e_14"),
						},
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"success",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Data: types.Round2Data{
						MemberID: 1,
						EncryptedSecretShares: tss.Scalars{
							[]byte("e_12"),
							[]byte("e_13"),
							[]byte("e_14"),
							[]byte("e_15"),
						},
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.SubmitDKGRound2(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func (s *KeeperTestSuite) TestComplain() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	privKeyI, _ := hex.DecodeString("7fc4175e7eb9661496cc38526f0eb4abccfd89d15f3371c3729e11c3ba1d6a14")
	pubKeyI, _ := hex.DecodeString("03936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa")
	pubKeyJ, _ := hex.DecodeString("03f70e80bac0b32b2599fa54d83b5471e90fac27bb09528f0337b49d464d64426f")
	member1 := types.Member{
		Address:     "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		PubKey:      pubKeyI,
		IsMalicious: false,
	}
	member2 := types.Member{
		Address:     "band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		PubKey:      pubKeyJ,
		IsMalicious: false,
	}

	// Create group for submit dkg context
	msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members: []string{
			"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
			"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
		},
		Threshold: 3,
		Sender:    "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
	})

	// Update member public key
	k.SetMember(ctx, groupID, memberID1, member1)
	k.SetMember(ctx, groupID, memberID2, member2)

	// Update group to round 3
	k.SetGroup(ctx, types.Group{
		GroupID:   1,
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_3,
	})

	// Sign
	sig, keySym, err := tss.SignComplain(pubKeyI, pubKeyJ, privKeyI)
	s.Require().NoError(err)

	var req types.MsgComplain
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		// TODO: add more test case
		{
			"success",
			func() {
				req = types.MsgComplain{
					GroupID: groupID,
					Complains: []types.Complain{
						{
							I:      memberID1,
							J:      memberID2,
							KeySym: keySym,
							Sig:    sig,
						},
					},
					Member: "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.Complain(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func (s *KeeperTestSuite) TestConfirm() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	groupID, memberID1, memberID2 := tss.GroupID(1), tss.MemberID(1), tss.MemberID(2)
	dkgContext, _ := hex.DecodeString("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c")
	member1p1, _ := hex.DecodeString("039cff182d3b5653c215207c5b141983d6e784e51cc8088b7edfef6cba504573e3")
	member1p2, _ := hex.DecodeString("035b8a99ebc56c07b88404407046e6f0d5e5318a87b888ea25d6d12d8175b2d70c")
	member2p1, _ := hex.DecodeString("02786741d28ca0a66b628d6401d975da448fc08c15a1228eb7b65203c6bac5cedb")
	member2p2, _ := hex.DecodeString("023d61b24c8785efe8c7459dc706d95b197c0acb31697feb49fec2d3446dc36de4")
	groupPubKey, _ := hex.DecodeString("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044")
	sig, _ := hex.DecodeString(
		"02bf7d39a54f6d468ce71317e2d5cc87c34c4ef11ee2b6638f57b435dadd7a976520e65c8e296ff1570ad0bb4a5f18557126642e76cbda0f6ffd4a546ea4651ef8",
	)

	// Create group for submit dkg context
	msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members: []string{
			"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
			"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
			"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
		},
		Threshold: 3,
		Sender:    "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
	})

	// Set dkg context
	k.SetDKGContext(ctx, groupID, dkgContext)

	// Set group to round 3
	k.SetGroup(ctx, types.Group{
		GroupID:   1,
		Size_:     5,
		Threshold: 2,
		PubKey:    groupPubKey,
		Status:    types.GROUP_STATUS_ROUND_3,
	})

	// Set round 1 data
	k.SetRound1Data(ctx, groupID, types.Round1Data{
		MemberID:           memberID1,
		CoefficientsCommit: tss.Points{member1p1, member1p2},
	})
	k.SetRound1Data(ctx, groupID, types.Round1Data{
		MemberID:           memberID2,
		CoefficientsCommit: tss.Points{member2p1, member2p2},
	})

	m1, _ := k.GetMember(ctx, 1, 1)
	m1.PubKey, _ = hex.DecodeString("0268c34a74f75ea26f3eba73a44afdaaa5e4704baa6f58d6e1ab831a5608e4dae4")
	k.SetMember(ctx, groupID, 1, m1)

	m2, _ := k.GetMember(ctx, 1, 2)
	m2.PubKey, _ = hex.DecodeString("034c0386dff08b142f356c0c7ae610c9cba27239a5447cde69c7c953b7b65f89c7")
	k.SetMember(ctx, groupID, 2, m2)

	var req types.MsgConfirm
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		// TODO: add more test case
		{
			"success",
			func() {
				req = types.MsgConfirm{
					GroupID:      groupID,
					MemberID:     memberID1,
					OwnPubKeySig: sig,
					Member:       "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.Confirm(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func (s *KeeperTestSuite) TestSubmitDEs() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberAddress := "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad"
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}
	group := types.Group{
		GroupID:   groupID,
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_3,
	}

	// Set group
	k.SetGroup(ctx, group)

	var req types.MsgSubmitDEs
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		{
			"failure with invalid member address",
			func() {
				req = types.MsgSubmitDEs{
					DEs:    des,
					Member: "invalidMemberAddress", //invalid address
				}
			},
			false,
			func() {},
		},
		{
			"success",
			func() {
				req = types.MsgSubmitDEs{
					DEs:    des,
					Member: memberAddress,
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.SubmitDEs(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func (s *KeeperTestSuite) TestRequestSign() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	accMember, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	member := types.Member{
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      tss.PublicKey(nil),
		IsMalicious: false,
	}
	message := []byte("test message")
	de := types.DE{
		PubD: hexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044"),
		PubE: hexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044"),
	}

	// Create an active group
	group := types.Group{
		GroupID:   groupID,
		Size_:     1,
		Threshold: 1,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ACTIVE,
	}
	k.SetGroup(ctx, group)
	k.SetMember(ctx, groupID, memberID, member)
	k.SetDE(ctx, accMember, 0, de)

	var req types.MsgRequestSign
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		{
			"failure with invalid groupID",
			func() {
				req = types.MsgRequestSign{
					GroupID: tss.GroupID(999), // non-existent groupID
					Message: message,
				}
			},
			false,
			func() {},
		},
		{
			"failure with inactive group",
			func() {
				inactiveGroup := types.Group{
					GroupID:   2,
					Size_:     5,
					Threshold: 3,
					PubKey:    nil,
					Status:    types.GROUP_STATUS_FALLEN,
				}
				k.SetGroup(ctx, inactiveGroup)
				req = types.MsgRequestSign{
					GroupID: tss.GroupID(2), // inactive groupID
					Message: message,
				}
			},
			false,
			func() {},
		},
		{
			"success",
			func() {
				req = types.MsgRequestSign{
					GroupID: groupID,
					Message: message,
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.RequestSign(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func (s *KeeperTestSuite) TestSign() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	groupID, signingID, member1, member2 := tss.GroupID(1), tss.SigningID(1), tss.MemberID(1), tss.MemberID(2)
	member := types.Member{
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      hexDecode("0268c34a74f75ea26f3eba73a44afdaaa5e4704baa6f58d6e1ab831a5608e4dae4"),
		IsMalicious: false,
	}
	signature := hexDecode(
		"03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67e60572162975eedcd2c605e480ebc5293a7b11472e911ab37e9139a2fb60eddd",
	)
	message := []byte("data")

	// Create an active group and a signing in pending state
	group := types.Group{
		GroupID:   groupID,
		Size_:     2,
		Threshold: 2,
		PubKey:    hexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044"),
		Status:    types.GROUP_STATUS_ACTIVE,
	}
	k.SetGroup(ctx, group)
	k.SetMember(ctx, groupID, member1, member)

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
			{
				MemberID: member2,
				Member:   "band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("02aacc8be43d6af147efc41f41754acc7764f31b9d0be33a5acbf9bd46bd3bb4bc"),
			},
		},
		Message:       message,
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
		Sig:           nil,
	}
	k.AddSigning(ctx, signing)

	var req types.MsgSign
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		postTest func()
	}{
		{
			"failure with invalid signingID",
			func() {
				req = types.MsgSign{
					SigningID: tss.SigningID(999), // non-existent signingID
					MemberID:  member1,
					Sig:       signature,
					Member:    "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			false,
			func() {},
		},
		{
			"failure with invalid memberID",
			func() {
				req = types.MsgSign{
					SigningID: signingID,
					MemberID:  tss.MemberID(999), // non-existent memberID
					Sig:       signature,
					Member:    "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			false,
			func() {},
		},
		{
			"success",
			func() {
				req = types.MsgSign{
					SigningID: signingID,
					MemberID:  member1,
					Sig:       signature,
					Member:    "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			_, err := msgSrvr.Sign(ctx, &req)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.postTest()
		})
	}
}

func hexDecode(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}
