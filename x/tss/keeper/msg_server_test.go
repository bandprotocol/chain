package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
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

func (s *KeeperTestSuite) TestSubmitDKGRound1Req() {
	ctx, msgSrvr, _ := s.ctx, s.msgSrvr, s.app.TSSKeeper

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

	// TODO: add more test
	var req types.MsgSubmitDKGRound1
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"Success",
			func() {
				// Key generated from GenerateRound1Data() ref. github.com/bandprotocol/chain/v2/pkg/tss
				req = types.MsgSubmitDKGRound1{
					GroupID:  1,
					MemberID: 0,
					CoefficientsCommit: []tss.Point{
						[]byte{3, 75, 66, 222, 122, 121, 223, 81, 177, 137, 217, 188, 160, 151, 150, 135, 153, 30, 96, 160, 243, 5, 55, 70, 176, 75, 233, 146, 29, 52, 86, 136, 103},
						[]byte{2, 148, 129, 160, 43, 93, 6, 105, 254, 8, 38, 225, 162, 222, 78, 134, 254, 186, 88, 253, 104, 229, 101, 194, 207, 244, 98, 213, 122, 83, 202, 227, 173},
						[]byte{3, 148, 18, 67, 72, 98, 47, 138, 90, 231, 22, 122, 89, 94, 130, 50, 191, 41, 139, 35, 181, 47, 157, 222, 191, 112, 169, 121, 129, 120, 17, 38, 83},
					},
					OneTimePubKey: []byte{2, 74, 147, 225, 29, 62, 171, 171, 174, 0, 205, 215, 222, 51, 150, 123, 78, 39, 114, 61, 138, 148, 104, 130, 118, 246, 127, 217, 114, 144, 27, 137, 123},
					A0Sig:         []byte{83, 123, 146, 36, 179, 169, 176, 247, 134, 245, 167, 120, 246, 25, 36, 36, 139, 25, 148, 88, 235, 108, 43, 119, 195, 175, 226, 135, 183, 123, 183, 19, 88, 24, 84, 16, 8, 245, 137, 149, 237, 152, 229, 66, 184, 162, 118, 120, 185, 240, 193, 116, 82, 41, 200, 212, 173, 178, 181, 214, 182, 207, 146, 181},
					OneTimeSig:    []byte{218, 250, 159, 247, 123, 66, 233, 19, 11, 15, 146, 70, 175, 1, 235, 147, 11, 75, 179, 78, 23, 181, 215, 169, 185, 86, 167, 105, 192, 64, 82, 127, 59, 133, 198, 135, 60, 117, 90, 28, 149, 170, 184, 7, 42, 174, 212, 175, 118, 130, 174, 66, 164, 177, 135, 217, 156, 33, 30, 242, 186, 246, 123, 173},
					Member:        "band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
				}
			},
			true,
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
		})
	}
}
