package keeper_test

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TestCase struct {
	Msg      string
	Malleate func()
	ExpPass  bool
	PostTest func()
}

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
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})

		// Set dkg context
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)
	}

	var tcs []TestCase
	var req types.MsgSubmitDKGRound1

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		tcGroup := tc.Group
		tcs = append(tcs, TestCase{
			Msg: fmt.Sprintf("success %s", tc.Name),
			Malleate: func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tcGroup.ID,
					Round1Info: types.Round1Info{
						MemberID:           tcGroup.Members[0].ID,
						CoefficientsCommit: tcGroup.Members[0].CoefficientsCommit,
						OneTimePubKey:      tcGroup.Members[0].OneTimePubKey(),
						A0Sig:              tcGroup.Members[0].A0Sig,
						OneTimeSig:         tcGroup.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tcGroup.Members[0].PubKey()).String(),
				}
			},
			ExpPass: true,
			PostTest: func() {
				k.DeleteRound1Info(ctx, tcGroup.ID, tcGroup.Members[0].ID)
			},
		})
	}

	// Add failed cases
	tc1Group := testutil.TestCases[0].Group
	failedCases := []TestCase{
		{
			"group not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: 99,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientsCommit: tc1Group.Members[0].CoefficientsCommit,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
		{
			"member not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           99,
						CoefficientsCommit: tc1Group.Members[0].CoefficientsCommit,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
		{
			"round 1 already commit",
			func() {
				// Set round 1 info
				k.SetRound1Info(ctx, tc1Group.ID, types.Round1Info{
					MemberID:           tc1Group.Members[0].ID,
					CoefficientsCommit: tc1Group.Members[0].CoefficientsCommit,
					OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
					A0Sig:              tc1Group.Members[0].A0Sig,
					OneTimeSig:         tc1Group.Members[0].OneTimeSig,
				})

				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientsCommit: tc1Group.Members[0].CoefficientsCommit,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {
				k.DeleteRound1Info(ctx, tc1Group.ID, tc1Group.Members[0].ID)
			},
		},
		{
			"wrong one time sign",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientsCommit: tc1Group.Members[0].CoefficientsCommit,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         []byte("wrong one_time_sig"),
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
		{
			"wrong a0 sig",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientsCommit: tc1Group.Members[0].CoefficientsCommit,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              []byte("wrong a0_sig"),
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
	}
	tcs = append(tcs, failedCases...)

	// Run test cases
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDKGRound1(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSubmitDKGRound2Req() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})

		// Update group status
		group, err := k.GetGroup(ctx, tc.Group.ID)
		s.Require().NoError(err)
		group.Status = types.GROUP_STATUS_ROUND_2
		k.SetGroup(ctx, group)

		// Set dkg context
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)
	}

	var tcs []TestCase
	var req types.MsgSubmitDKGRound2

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		tcGroup := tc.Group
		tcs = append(tcs, TestCase{
			Msg: fmt.Sprintf("success %s", tc.Name),
			Malleate: func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: tcGroup.ID,
					Round2Info: types.Round2Info{
						MemberID:              tcGroup.Members[0].ID,
						EncryptedSecretShares: tcGroup.Members[0].EncSecretShares,
					},
					Member: sdk.AccAddress(tcGroup.Members[0].PubKey()).String(),
				}
			},
			ExpPass: true,
			PostTest: func() {
				k.DeleteRound2Info(ctx, tcGroup.ID, tcGroup.Members[0].ID)
			},
		})
	}

	// Add failed case
	tc1Group := testutil.TestCases[0].Group
	failedCases := []TestCase{
		{
			"group not found",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: 99,
					Round2Info: types.Round2Info{
						MemberID:              tc1Group.Members[0].ID,
						EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
		{
			"member not authorized",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: tc1Group.ID,
					Round2Info: types.Round2Info{
						MemberID:              99,
						EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
		{
			"round 2 already submit",
			func() {
				// Set round 2 info
				k.SetRound2Info(ctx, tc1Group.ID, types.Round2Info{
					MemberID:              tc1Group.Members[0].ID,
					EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
				})

				req = types.MsgSubmitDKGRound2{
					GroupID: tc1Group.ID,
					Round2Info: types.Round2Info{
						MemberID:              tc1Group.Members[0].ID,
						EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {
				k.DeleteRound2Info(ctx, tc1Group.ID, tc1Group.Members[0].ID)
			},
		},
		{
			"number of encrypted secret shares is not correct",
			func() {
				inValidEncSecretShares := append(tc1Group.Members[0].EncSecretShares, []byte("enc"))
				req = types.MsgSubmitDKGRound2{
					GroupID: tc1Group.ID,
					Round2Info: types.Round2Info{
						MemberID:              tc1Group.Members[0].ID,
						EncryptedSecretShares: inValidEncSecretShares,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			false,
			func() {},
		},
	}
	tcs = append(tcs, failedCases...)

	// Run test cases
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDKGRound2(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestComplain() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})

		// Update group status
		group, err := k.GetGroup(ctx, tc.Group.ID)
		s.Require().NoError(err)
		group.Status = types.GROUP_STATUS_ROUND_3
		k.SetGroup(ctx, group)

		// Update member public key
		for i, m := range tc.Group.Members {
			member := types.Member{
				MemberID:    tss.MemberID(i + 1),
				Address:     sdk.AccAddress(m.PubKey()).String(),
				PubKey:      m.PubKey(),
				IsMalicious: false,
			}
			k.SetMember(ctx, tc.Group.ID, member)
		}
	}

	var tcs []TestCase
	var req types.MsgComplain

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		tcGroup := tc.Group
		tcs = append(tcs, TestCase{
			Msg: fmt.Sprintf("success %s", tc.Name),
			Malleate: func() {
				sig, keySym, err := tss.SignComplaint(
					tcGroup.Members[0].OneTimePubKey(),
					tcGroup.Members[1].OneTimePubKey(),
					tcGroup.Members[0].OneTimePrivKey,
				)
				s.Require().NoError(err)

				req = types.MsgComplain{
					GroupID: tcGroup.ID,
					Complaints: []types.Complaint{
						{
							Complainer:  tcGroup.Members[0].ID,
							Complainant: tcGroup.Members[1].ID,
							KeySym:      keySym,
							Signature:   sig,
						},
					},
					Member: sdk.AccAddress(tcGroup.Members[0].PubKey()).String(),
				}
			},
			ExpPass:  true,
			PostTest: func() {},
		})
	}

	// TODO: add failed test cases

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.Complain(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestConfirm() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})

		// Update group status
		group, err := k.GetGroup(ctx, tc.Group.ID)
		s.Require().NoError(err)
		group.Status = types.GROUP_STATUS_ROUND_3
		k.SetGroup(ctx, group)

		// Set dkg context
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)

		// Update member public key
		for i, m := range tc.Group.Members {
			member := types.Member{
				MemberID:    tss.MemberID(i + 1),
				Address:     sdk.AccAddress(m.PubKey()).String(),
				PubKey:      m.PubKey(),
				IsMalicious: false,
			}
			k.SetMember(ctx, tc.Group.ID, member)
		}
	}

	var tcs []TestCase
	var req types.MsgConfirm

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		tcGroup := tc.Group
		tcs = append(tcs, TestCase{
			Msg: fmt.Sprintf("success %s", tc.Name),
			Malleate: func() {
				req = types.MsgConfirm{
					GroupID:      tcGroup.ID,
					MemberID:     tcGroup.Members[0].ID,
					OwnPubKeySig: tcGroup.Members[0].PubKeySig,
					Member:       sdk.AccAddress(tcGroup.Members[0].PubKey()).String(),
				}
			},
			ExpPass:  true,
			PostTest: func() {},
		})
	}

	// TODO: add failed test cases

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.Confirm(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSubmitDEs() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	var req types.MsgSubmitDEs

	// Add failed case
	tcs := []TestCase{
		{
			"success with 1 DE",
			func() {
				req = types.MsgSubmitDEs{
					DEs:    []types.DE{de},
					Member: "band197gn3gpq4f8ylnufwxjsafznxykglgacf4t384",
				}
			},
			true,
			func() {},
		},
		{
			"success with 99 DEs",
			func() {
				var deList []types.DE
				for i := 0; i < 99; i++ {
					deList = append(deList, de)
				}

				req = types.MsgSubmitDEs{
					DEs:    deList,
					Member: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			true,
			func() {},
		},
		{
			"failure with number of DE more than max",
			func() {
				var deList []types.DE
				for i := 0; i < 100; i++ {
					deList = append(deList, de)
				}

				req = types.MsgSubmitDEs{
					DEs:    deList,
					Member: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
				}
			},
			false,
			func() {},
		},
		{
			"failure with invalid member address",
			func() {
				req = types.MsgSubmitDEs{
					DEs:    []types.DE{de},
					Member: "invalidMemberAddress", //invalid address
				}
			},
			false,
			func() {},
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDEs(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestRequestSign() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})

		// Update group status
		group, err := k.GetGroup(ctx, tc.Group.ID)
		s.Require().NoError(err)
		group.Status = types.GROUP_STATUS_ACTIVE
		k.SetGroup(ctx, group)

		// Add DEs
		for _, signing := range tc.Signings {
			for _, am := range signing.AssignedMembers {
				pubD := am.PrivD.Point()
				pubE := am.PrivE.Point()

				member, err := k.GetMember(ctx, group.GroupID, am.ID)
				s.Require().NoError(err)
				address, err := sdk.AccAddressFromBech32(member.Address)
				s.Require().NoError(err)

				k.HandleSetDEs(ctx, address, []types.DE{
					{
						PubD: pubD,
						PubE: pubE,
					},
				})
			}
		}
	}

	var tcs []TestCase
	var req types.MsgRequestSignature

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		tcGroup := tc.Group
		tcs = append(tcs, TestCase{
			Msg: fmt.Sprintf("success %s", tc.Name),
			Malleate: func() {
				req = types.MsgRequestSignature{
					GroupID: tcGroup.ID,
					Message: tc.Signings[0].Data,
					Sender:  sdk.AccAddress(tcGroup.Members[0].PubKey()).String(),
				}
			},
			ExpPass:  true,
			PostTest: func() {},
		})
	}

	// Add failed case
	failedCases := []TestCase{
		{
			"failure with invalid groupID",
			func() {
				req = types.MsgRequestSignature{
					GroupID: tss.GroupID(999), // non-existent groupID
					Message: []byte("data"),
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
				req = types.MsgRequestSignature{
					GroupID: tss.GroupID(2), // inactive groupID
					Message: []byte("data"),
				}
			},
			false,
			func() {},
		},
	}
	tcs = append(tcs, failedCases...)

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.RequestSignature(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSign() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	expiryTime := ctx.BlockHeader().Time.Add(k.SigningPeriod(ctx))

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})

		// Get group
		group, err := k.GetGroup(ctx, tc.Group.ID)
		s.Require().NoError(err)

		// Update group status and public key
		group.Status = types.GROUP_STATUS_ACTIVE
		group.PubKey = tc.Group.PubKey
		k.SetGroup(ctx, group)

		// Update member public key
		for i, m := range tc.Group.Members {
			member := types.Member{
				MemberID:    tss.MemberID(i + 1),
				Address:     sdk.AccAddress(m.PubKey()).String(),
				PubKey:      m.PubKey(),
				IsMalicious: false,
			}
			k.SetMember(ctx, tc.Group.ID, member)
		}
	}

	var tcs []TestCase
	var req types.MsgSign

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		tcSignings := tc.Signings
		tcGroup := tc.Group

		for _, si := range tcSignings {
			signing := si
			tcs = append(tcs, TestCase{
				Msg: fmt.Sprintf("success %s signing ID: %d", tc.Name, signing.ID),
				Malleate: func() {
					// Combine assigned member
					var ams []types.AssignedMember
					for _, am := range signing.AssignedMembers {
						member, err := k.GetMember(ctx, tcGroup.ID, am.ID)
						s.Require().NoError(err)

						pubD := am.PrivD.Point()
						pubE := am.PrivE.Point()

						ams = append(ams, types.AssignedMember{
							MemberID: am.ID,
							Member:   member.Address,
							PubD:     pubD,
							PubE:     pubE,
							PubNonce: am.PubNonce(),
						})
					}

					// Set signing
					k.SetSigning(ctx, types.Signing{
						SigningID:       signing.ID,
						GroupID:         tcGroup.ID,
						AssignedMembers: ams,
						Message:         signing.Data,
						GroupPubNonce:   signing.PubNonce,
						Commitment:      signing.Commitment,
						Signature:       nil,
						ExpiryTime:      &expiryTime,
					})

					req = types.MsgSign{
						SigningID: signing.ID,
						MemberID:  signing.AssignedMembers[0].ID,
						Signature: signing.AssignedMembers[0].Sig,
						Member:    sdk.AccAddress(tcGroup.Members[0].PubKey()).String(),
					}
				},
				ExpPass: true,
				PostTest: func() {
					for _, signing := range tc.Signings {
						k.DeleteSigning(ctx, signing.ID)
					}
				},
			})
		}
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.Sign(ctx, &req)
			if tc.ExpPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			tc.PostTest()
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
