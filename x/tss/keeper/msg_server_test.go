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
	cof1B, _ := hex.DecodeString("034b42de7a79df51b189d9bca0979687991e60a0f3053746b04be9921d34568867")
	cof2B, _ := hex.DecodeString("02212ca02b5d0669fe0826e1a2de4e86feba58fd68e565c2cff462d57a53cae3ad")
	cof3B, _ := hex.DecodeString("03941243486f2f8a5ae7167a595e8232bf298b23b52f9ddebf70a9798178112653")
	oneTimePubKeyB, _ := hex.DecodeString("024a93e11d3eababae00cdd7de33967b4e27723d8a94688276f67fd972901b897b")
	a0SigB, _ := hex.DecodeString("537b9224b3a9b0f786f5a778f61924248b199458eb6c2b77c3afe287b77bb7135818541008f58995ed98e542b8a27678b9f0c1745229c8d4adb2b5d6b6cf92b5")
	oneTimeSigB, _ := hex.DecodeString("dafa9ff77b42e9130b0f9246af01eb930b4bb34e17b5d7a9b956a769c040527f3b85c6873c755a1c95aab8072aaed4af7682ae42a4b187d99c211ef2baf67bad")

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
					GroupID:            0,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
					Member:             "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"member not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID:            1,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
					Member:             "band1rqjc6czdeu2w2nst9vfvv6yqj6nwqkv48s4jmq",
				}
			},
			false,
			func() {},
		},
		{
			"round 1 already commit",
			func() {
				// Set round 1 commitments
				tssKeeper.SetRound1Commitments(ctx, 1, 1, types.Round1Commitments{
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
				})

				req = types.MsgSubmitDKGRound1{
					GroupID:            1,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
					Member:             "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {
				tssKeeper.DeleteRound1Commitments(ctx, 1, 1)
			},
		},
		{
			"wrong one_time_sign",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID:            1,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         []byte("wrong one_time_sign"),
					Member:             "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			false,
			func() {},
		},
		{
			"wrong a0_sig",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID:            1,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              []byte("wrong a0_sig"),
					OneTimeSig:         oneTimeSigB,
					Member:             "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
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
					GroupID:            1,
					CoefficientsCommit: []tss.Point{cof1B, cof2B, cof3B},
					OneTimePubKey:      oneTimePubKeyB,
					A0Sig:              a0SigB,
					OneTimeSig:         oneTimeSigB,
					Member:             "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
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
