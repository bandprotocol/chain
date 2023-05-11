package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestCreateGroupReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	s.Run("Create group", func() {
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

// TODO:
// func (s *KeeperTestSuite) TestSubmitDKGRound1Req() {
// 	ctx, msgSrvr, _ := s.ctx, s.msgSrvr, s.app.TSSKeeper

// 	// create group for submit dkg context
// 	msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
// 		Members: []string{
// 			"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
// 			"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
// 			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
// 			"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
// 			"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
// 		},
// 		Threshold: 3,
// 		Sender:    "band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
// 	})

// 	// TODO: add more test
// 	var req types.MsgSubmitDKGRound1
// 	testCases := []struct {
// 		msg      string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"Success",
// 			func() {
// 				req = types.MsgSubmitDKGRound1{
// 					GroupID:  1,
// 					MemberID: 1,
// 					CoefficientsCommit: []types.Point{
// 						[]byte("point1"),
// 						[]byte("point2"),
// 					},
// 					OneTimePubKey: []byte("OneTimePubKeySimple"),
// 					A0Sig:         []byte("A0SigSimple"),
// 					OneTimeSig:    []byte("OneTimeSigSimple"),
// 					Member:        "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
// 				}
// 			},
// 			true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			tc.malleate()

// 			_, err := msgSrvr.SubmitDKGRound1(ctx, &req)
// 			if tc.expPass {
// 				s.Require().NoError(err)
// 			} else {
// 				s.Require().Error(err)
// 			}
// 		})
// 	}
// }
