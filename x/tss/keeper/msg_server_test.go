package keeper_test

import (
	"fmt"
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TestCase struct {
	Msg      string
	Malleate func()
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

func (s *KeeperTestSuite) TestFailedSubmitDKGRound1Req() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	tc1Group := testutil.TestCases[0].Group

	// Setup group
	s.SetupGroup(types.GROUP_STATUS_ROUND_1)

	// Add failed cases
	var req types.MsgSubmitDKGRound1
	tcs := []TestCase{
		{
			"group not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: 99,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
		},
		{
			"member not found",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           99,
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
		},
		{
			"round 1 already commit",
			func() {
				// Set round 1 info
				k.SetRound1Info(ctx, tc1Group.ID, types.Round1Info{
					MemberID:           tc1Group.Members[0].ID,
					CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
					OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
					A0Sig:              tc1Group.Members[0].A0Sig,
					OneTimeSig:         tc1Group.Members[0].OneTimeSig,
				})

				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
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
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              tc1Group.Members[0].A0Sig,
						OneTimeSig:         []byte("wrong one_time_sig"),
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
		},
		{
			"wrong a0 sig",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Sig:              []byte("wrong a0_sig"),
						OneTimeSig:         tc1Group.Members[0].OneTimeSig,
					},
					Member: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
		},
	}

	// Run test cases
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDKGRound1(ctx, &req)
			s.Require().Error(err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitDKGRound1Req() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	expiration := ctx.BlockHeader().Time.Add(k.CreationPeriod(ctx))

	s.SetupGroup(types.GROUP_STATUS_ROUND_1)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				// Submit DKGRound1 message for each member
				_, err := msgSrvr.SubmitDKGRound1(ctx, &types.MsgSubmitDKGRound1{
					GroupID: tc.Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           m.ID,
						CoefficientCommits: m.CoefficientCommits,
						OneTimePubKey:      m.OneTimePubKey(),
						A0Sig:              m.A0Sig,
						OneTimeSig:         m.OneTimeSig,
					},
					Member: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Verify group status, expiration, and public key after submitting Round 1
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(types.GROUP_STATUS_ROUND_2, got.Status)
			s.Require().Equal(expiration, *got.Expiration)
			s.Require().Equal(tc.Group.PubKey, got.PubKey)

			// Clean up Round1Infos
			k.DeleteRound1Infos(ctx, tc.Group.ID)
		})
	}
}

func (s *KeeperTestSuite) TestFailedSubmitDKGRound2Req() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	tc1Group := testutil.TestCases[0].Group

	// Setup group
	s.SetupGroup(types.GROUP_STATUS_ROUND_2)

	// Add failed cases
	var req types.MsgSubmitDKGRound2
	tcs := []TestCase{
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
			func() {},
		},
	}

	// Run test cases
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDKGRound2(ctx, &req)
			s.Require().Error(err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitDKGRound2Req() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	expiration := ctx.BlockHeader().Time.Add(k.CreationPeriod(ctx))

	// Setup group as round 2
	s.SetupGroup(types.GROUP_STATUS_ROUND_2)

	// Add success test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				// Submit DKGRound2 message for each member
				_, err := msgSrvr.SubmitDKGRound2(ctx, &types.MsgSubmitDKGRound2{
					GroupID: tc.Group.ID,
					Round2Info: types.Round2Info{
						MemberID:              m.ID,
						EncryptedSecretShares: m.EncSecretShares,
					},
					Member: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Verify group status and expiration after submitting Round 2
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(got.Status, types.GROUP_STATUS_ROUND_3)
			s.Require().Equal(*got.Expiration, expiration)

			// Clean up Round1Infos and Round2Infos
			k.DeleteRound1Infos(ctx, tc.Group.ID)
			k.DeleteRound2Infos(ctx, tc.Group.ID)
		})
	}
}

func (s *KeeperTestSuite) TestSuccessComplainReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	complaintID := tss.MemberID(1)

	s.SetupGroup(types.GROUP_STATUS_ROUND_3)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			// Iterate through the group members to handle complaints
			for i, m := range tc.Group.Members {
				// Skip the respondent
				if i == 1 {
					continue
				}
				respondent := tc.Group.Members[complaintID]

				// Get respondent's Round 2 info
				respondentRound2, err := k.GetRound2Info(ctx, tc.Group.ID, respondent.ID)
				s.Require().NoError(err)

				// Determine which slot of encrypted secret shares is for respondent
				respondentSlot := types.FindMemberSlot(complaintID, m.ID)

				// Set fake encrypted secret shares
				respondentRound2.EncryptedSecretShares[respondentSlot] = testutil.FakePrivKey
				k.SetRound2Info(ctx, tc.Group.ID, respondentRound2)

				sig, keySym, err := tss.SignComplaint(
					m.OneTimePubKey(),
					respondent.OneTimePubKey(),
					m.OneTimePrivKey,
				)
				s.Require().NoError(err)

				// Complain the respondent
				_, err = msgSrvr.Complain(ctx, &types.MsgComplain{
					GroupID: tc.Group.ID,
					Complaints: []types.Complaint{
						{
							Complainant: m.ID,
							Respondent:  respondent.ID,
							KeySym:      keySym,
							Signature:   sig,
						},
					},
					Member: sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			respondent := tc.Group.Members[complaintID]

			// Complaint send message confirm
			_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
				GroupID:      tc.Group.ID,
				MemberID:     respondent.ID,
				OwnPubKeySig: respondent.PubKeySig,
				Member:       sdk.AccAddress(respondent.PubKey()).String(),
			})
			s.Require().NoError(err)

			// Check the group's status and expiration time after complain
			var nilTime *time.Time
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(types.GROUP_STATUS_FALLEN, got.Status)
			s.Require().Equal(nilTime, got.Expiration)
		})
	}
}

func (s *KeeperTestSuite) TestSuccessConfirmReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(types.GROUP_STATUS_ROUND_3)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			// Confirm the participation of each member in the group
			for _, m := range tc.Group.Members {
				_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
					GroupID:      tc.Group.ID,
					MemberID:     m.ID,
					OwnPubKeySig: m.PubKeySig,
					Member:       sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Check the group's status and expiration time after confirmation
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(types.GROUP_STATUS_ACTIVE, got.Status)
			s.Require().Nil(got.Expiration)
		})
	}
}

func (s *KeeperTestSuite) TestFailedSubmitDEsReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	var req types.MsgSubmitDEs

	// Add failed case
	tcs := []TestCase{
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
			func() {},
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDEs(ctx, &req)
			s.Require().Error(err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitDEsReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("Case %s", fmt.Sprintf("success %s", tc.Name)), func() {
			for _, m := range tc.Group.Members {
				// Submit DEs for each member in the group
				_, err := msgSrvr.SubmitDEs(ctx, &types.MsgSubmitDEs{
					DEs:    []types.DE{de},
					Member: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Verify that each member has the correct DE
			for _, m := range tc.Group.Members {
				got, err := k.GetDE(ctx, sdk.AccAddress(m.PubKey()), 0)
				s.Require().NoError(err)
				s.Require().Equal(de, got)
			}
		})
	}
}

func (s *KeeperTestSuite) TestFailedRequestSignReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	var req types.MsgRequestSignature

	// Add failed case
	tcs := []TestCase{
		{
			"failure with invalid groupID",
			func() {
				req = types.MsgRequestSignature{
					GroupID: tss.GroupID(999), // non-existent groupID
					Message: []byte("data"),
				}
			},
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
			func() {},
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.RequestSignature(ctx, &req)
			s.Require().Error(err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessRequestSignReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		// Request signature for each member in the group
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				_, err := msgSrvr.RequestSignature(ctx, &types.MsgRequestSignature{
					GroupID: tc.Group.ID,
					Message: tc.Signings[0].Data,
					Sender:  sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestFailedSubmitSignatureReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	expiration := ctx.BlockHeader().Time.Add(k.SigningPeriod(ctx))

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	var req types.MsgSubmitSignature

	// Add test cases
	tc1 := testutil.TestCases[0]
	tcs := []TestCase{
		{
			"failure with invalid signingID",
			func() {
				req = types.MsgSubmitSignature{
					SigningID: tss.SigningID(99), // non-existent signingID
					MemberID:  tc1.Group.Members[0].ID,
					Signature: tc1.Signings[0].Sig,
					Member:    sdk.AccAddress(tc1.Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
		},
		{
			"failure with invalid memberID",
			func() {
				k.SetSigning(ctx, types.Signing{
					SigningID:       tc1.Signings[0].ID,
					GroupID:         tc1.Group.ID,
					AssignedMembers: []types.AssignedMember{},
					Message:         tc1.Signings[0].Data,
					GroupPubNonce:   tc1.Signings[0].PubNonce,
					Commitment:      tc1.Signings[0].Commitment,
					Signature:       nil,
					Expiration:      &expiration,
				})

				req = types.MsgSubmitSignature{
					SigningID: tc1.Signings[0].ID,
					MemberID:  tss.MemberID(99), // non-existent memberID
					Signature: tc1.Signings[0].Sig,
					Member:    sdk.AccAddress(tc1.Group.Members[0].PubKey()).String(),
				}
			},
			func() {
				k.DeleteSigning(ctx, tc1.Signings[0].ID)
			},
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitSignature(ctx, &req)
			s.Require().Error(err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitSignatureReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Iterate through test cases from testutil
	for i, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			// Request signature for the first member in the group
			_, err := msgSrvr.RequestSignature(ctx, &types.MsgRequestSignature{
				GroupID: tc.Group.ID,
				Message: []byte("msg"),
				Sender:  sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
			})
			s.Require().NoError(err)

			// Get the signing information
			signing, err := k.GetSigning(ctx, tss.SigningID(i+1))
			s.Require().NoError(err)

			// Get the group information
			group, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)

			// Process signing for each assigned member
			for _, am := range signing.AssignedMembers {
				// Compute Lagrange coefficient
				var lgc tss.Scalar
				mids := types.AssignedMembers(signing.AssignedMembers).MemberIDs()
				if len(mids) <= 20 {
					// Compute the Lagrange coefficient using the optimized operation
					lgc = tss.ComputeLagrangeCoefficientOp(
						am.MemberID,
						types.AssignedMembers(signing.AssignedMembers).MemberIDs(),
					)
				} else {
					// Compute the Lagrange coefficient using the default implementation
					lgc = tss.ComputeLagrangeCoefficient(
						am.MemberID,
						types.AssignedMembers(signing.AssignedMembers).MemberIDs(),
					)
				}

				// Compute private nonce
				pn, err := tss.ComputeOwnPrivNonce(PrivD, PrivE, am.BindingFactor)
				s.Require().NoError(err)

				// Sign the message
				sig, err := tss.SignSigning(
					signing.GroupPubNonce,
					group.PubKey,
					signing.Message,
					lgc,
					pn,
					tc.Group.GetMember(am.MemberID).PrivKey,
				)
				s.Require().NoError(err)

				// Submit the signature
				_, err = msgSrvr.SubmitSignature(ctx, &types.MsgSubmitSignature{
					SigningID: tss.SigningID(i + 1),
					MemberID:  am.MemberID,
					Signature: sig,
					Member:    sdk.AccAddress(tc.Group.GetMember(am.MemberID).PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Retrieve the signing information after signing
			signing, err = k.GetSigning(ctx, tss.SigningID(i+1))
			s.Require().NoError(err)
			s.Require().NotNil(signing.Signature)
			s.Require().Nil(signing.Expiration)
		})
	}
}
