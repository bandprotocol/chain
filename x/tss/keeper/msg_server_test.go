package keeper_test

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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
			"round1 already commit",
			func() {
				// Set round1 data
				tssKeeper.SetRound1Data(ctx, 1, 1, types.Round1Data{
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
				})

				req = types.MsgSubmitDKGRound1{
					GroupID: 1,
					Round1Data: types.Round1Data{
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
	k.UpdateGroup(ctx, 1, types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.ROUND_2,
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
					Round2Share: types.Round2Share{
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
			"member not found",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Share: types.Round2Share{
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
			"round2already submit",
			func() {
				// Set round 2
				k.SetRound2Share(ctx, 1, 1, types.Round2Share{EncryptedSecretShares: tss.Scalars{
					[]byte("e_12"),
					[]byte("e_13"),
					[]byte("e_14"),
					[]byte("e_15"),
				}})

				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Share: types.Round2Share{
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
				k.DeleteRound2share(ctx, 1, 1)
			},
		},
		{
			"round2share is not correct length n-1",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 1,
					Round2Share: types.Round2Share{
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
					Round2Share: types.Round2Share{
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
