package keeper_test

import (
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type TestCase struct {
	Msg         string
	Malleate    func()
	PostTest    func()
	ExpectedErr error
}

func (s *KeeperTestSuite) TestCreateGroupReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	members := []string{
		"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
		"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	}

	for _, m := range members {
		msgSrvr.Activate(ctx, &types.MsgActivate{
			Address: m,
		})
		msgSrvr.SubmitDEs(ctx, &types.MsgSubmitDEs{
			DEs: []types.DE{
				{
					PubD: testutil.HexDecode("dddd"),
					PubE: testutil.HexDecode("eeee"),
				},
			},
			Address: m,
		})
	}

	s.Run("create group", func() {
		_, err := msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: 3,
			Fee:       sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
			Authority: s.authority.String(),
		})
		s.Require().NoError(err)
	})
}

func (s *KeeperTestSuite) TestFailedReplaceGroup() {
	// Set up the test context and message server.
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create an authority address.
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	// Define fromGroupID and toGroupID.
	fromGroupID := tss.GroupID(2)
	toGroupID := tss.GroupID(1)

	// Create a replace group message.
	var req types.MsgReplaceGroup

	// Set up the test by creating an active group.
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)
	group := k.MustGetGroup(ctx, fromGroupID)

	// Define test cases.
	tcs := []TestCase{
		{
			"failure due to incorrect authority",
			func() {
				req = types.MsgReplaceGroup{
					Authority:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					FromGroupID: fromGroupID,
					ToGroupID:   toGroupID,
					ExecTime:    time.Now().UTC(),
				}
			},
			func() {
			},
			govtypes.ErrInvalidSigner,
		},
		{
			"failure due to group is not active",
			func() {
				req = types.MsgReplaceGroup{
					Authority:   authority.String(),
					FromGroupID: fromGroupID,
					ToGroupID:   toGroupID,
					ExecTime:    time.Now().UTC(),
				}
				group.Status = types.GROUP_STATUS_FALLEN
				k.SetGroup(ctx, group)
			},
			func() {
				group.Status = types.GROUP_STATUS_ACTIVE
				k.SetGroup(ctx, group)
			},
			types.ErrGroupIsNotActive,
		},
	}

	// Loop through each test case.
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			// Modify the request based on the test case.
			tc.Malleate()

			// Execute the ReplaceGroup method and check for expected errors.
			_, err := msgSrvr.ReplaceGroup(ctx, &req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			// Perform post-test actions.
			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessReplaceGroup() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	toGroupID, replacementID := tss.GroupID(1), uint64(1)

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	now := time.Now()

	_, err := msgSrvr.ReplaceGroup(ctx, &types.MsgReplaceGroup{
		FromGroupID: 2,
		ToGroupID:   1,
		ExecTime:    now,
		Authority:   s.authority.String(),
	})
	s.Require().NoError(err)

	replacement, err := k.GetReplacement(ctx, replacementID)
	s.Require().NoError(err)

	replacementIterator := k.ReplacementQueueIterator(ctx, now)
	s.Require().True(replacementIterator.Valid())

	gotReplacementID, _ := types.SplitReplacementQueueKey(replacementIterator.Key())
	s.Require().Equal(replacement.ID, gotReplacementID)

	replacementIterator.Close()

	toGroup := k.MustGetGroup(ctx, toGroupID)
	s.Require().Equal(gotReplacementID, toGroup.LatestReplacementID)
}

func (s *KeeperTestSuite) TestFailedUpdateGroupFee() {
	// Set up the test context and message server.
	ctx, msgSrvr := s.ctx, s.msgSrvr

	// Define GroupID
	groupID := tss.GroupID(1)

	// Create a update group fee message.
	var req types.MsgUpdateGroupFee

	// Define test cases.
	tcs := []TestCase{
		{
			"failure due to incorrect authority",
			func() {
				req = types.MsgUpdateGroupFee{
					GroupID:   groupID,
					Fee:       sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Authority: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				}
			},
			func() {
			},
			govtypes.ErrInvalidSigner,
		},
	}

	// Loop through each test case.
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			// Modify the request based on the test case.
			tc.Malleate()

			// Execute the UpdateGroupFee method and check for expected errors.
			_, err := msgSrvr.UpdateGroupFee(ctx, &req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			// Perform post-test actions.
			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessUpdateGroupFee() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Define GroupID
	groupID := tss.GroupID(1)

	// Create an authority address.
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	// Define the group fee.
	fee := sdk.NewCoins(sdk.NewInt64Coin("uband", 10))

	// Set up the test by creating an active group.
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Create the update group fee message.
	msg := types.MsgUpdateGroupFee{
		GroupID:   groupID,
		Fee:       fee,
		Authority: authority.String(),
	}

	// Execute the ReplaceGroup method.
	_, err := msgSrvr.UpdateGroupFee(ctx, &msg)
	s.Require().NoError(err)

	// Check if the pending replace group matches the expected result.
	got := k.MustGetGroup(ctx, groupID)
	s.Require().Equal(fee, got.Fee)
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
						A0Signature:        tc1Group.Members[0].A0Signature,
						OneTimeSignature:   tc1Group.Members[0].OneTimeSignature,
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrGroupNotFound,
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
						A0Signature:        tc1Group.Members[0].A0Signature,
						OneTimeSignature:   tc1Group.Members[0].OneTimeSignature,
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrMemberNotFound,
		},
		{
			"round 1 already commit",
			func() {
				// Add round 1 info
				k.AddRound1Info(ctx, tc1Group.ID, types.Round1Info{
					MemberID:           tc1Group.Members[0].ID,
					CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
					OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
					A0Signature:        tc1Group.Members[0].A0Signature,
					OneTimeSignature:   tc1Group.Members[0].OneTimeSignature,
				})

				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Signature:        tc1Group.Members[0].A0Signature,
						OneTimeSignature:   tc1Group.Members[0].OneTimeSignature,
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {
				k.DeleteRound1Info(ctx, tc1Group.ID, tc1Group.Members[0].ID)
			},
			types.ErrMemberAlreadySubmit,
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
						A0Signature:        tc1Group.Members[0].A0Signature,
						OneTimeSignature:   []byte("wrong one_time_sig"),
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrVerifyOneTimeSignatureFailed,
		},
		{
			"wrong a0 signature",
			func() {
				req = types.MsgSubmitDKGRound1{
					GroupID: tc1Group.ID,
					Round1Info: types.Round1Info{
						MemberID:           tc1Group.Members[0].ID,
						CoefficientCommits: tc1Group.Members[0].CoefficientCommits,
						OneTimePubKey:      tc1Group.Members[0].OneTimePubKey(),
						A0Signature:        []byte("wrong a0_sig"),
						OneTimeSignature:   tc1Group.Members[0].OneTimeSignature,
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrVerifyA0SignatureFailed,
		},
	}

	// Run test cases
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDKGRound1(ctx, &req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitDKGRound1Req() {
	ctx, app, msgSrvr, k := s.ctx, s.app, s.msgSrvr, s.app.TSSKeeper

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
						A0Signature:        m.A0Signature,
						OneTimeSignature:   m.OneTimeSignature,
					},
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Execute the EndBlocker to process groups
			app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})

			// Verify group status, expiration, and public key after submitting Round 1
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(types.GROUP_STATUS_ROUND_2, got.Status)
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
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrGroupNotFound,
		},
		{
			"member not found",
			func() {
				req = types.MsgSubmitDKGRound2{
					GroupID: tc1Group.ID,
					Round2Info: types.Round2Info{
						MemberID:              99,
						EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrMemberNotFound,
		},
		{
			"round 2 already submit",
			func() {
				// Add round 2 info
				k.AddRound2Info(ctx, tc1Group.ID, types.Round2Info{
					MemberID:              tc1Group.Members[0].ID,
					EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
				})

				req = types.MsgSubmitDKGRound2{
					GroupID: tc1Group.ID,
					Round2Info: types.Round2Info{
						MemberID:              tc1Group.Members[0].ID,
						EncryptedSecretShares: tc1Group.Members[0].EncSecretShares,
					},
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {
				k.DeleteRound2Info(ctx, tc1Group.ID, tc1Group.Members[0].ID)
			},
			types.ErrMemberAlreadySubmit,
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
					Address: sdk.AccAddress(tc1Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrInvalidLengthEncryptedSecretShares,
		},
	}

	// Run test cases
	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDKGRound2(ctx, &req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitDKGRound2Req() {
	ctx, app, msgSrvr, k := s.ctx, s.app, s.msgSrvr, s.app.TSSKeeper

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
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Execute the EndBlocker to process groups
			app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})

			// Verify group status and expiration after submitting Round 2
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(got.Status, types.GROUP_STATUS_ROUND_3)

			// Clean up Round1Infos and Round2Infos
			k.DeleteRound1Infos(ctx, tc.Group.ID)
			k.DeleteRound2Infos(ctx, tc.Group.ID)
		})
	}
}

func (s *KeeperTestSuite) TestSuccessComplainReq() {
	ctx, app, msgSrvr, k := s.ctx, s.app, s.msgSrvr, s.app.TSSKeeper
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
				respondentRound2.EncryptedSecretShares[respondentSlot] = testutil.FalsePrivKey
				k.AddRound2Info(ctx, tc.Group.ID, respondentRound2)

				signature, keySym, err := tss.SignComplaint(
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
							Signature:   signature,
						},
					},
					Address: sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			respondent := tc.Group.Members[complaintID]

			// Complaint send message confirm
			_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
				GroupID:      tc.Group.ID,
				MemberID:     respondent.ID,
				OwnPubKeySig: respondent.PubKeySignature,
				Address:      sdk.AccAddress(respondent.PubKey()).String(),
			})
			s.Require().NoError(err)

			// Execute the EndBlocker to process groups
			app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})

			// Check the group's status and expiration time after complain
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(types.GROUP_STATUS_FALLEN, got.Status)
		})
	}
}

func (s *KeeperTestSuite) TestSuccessConfirmReq() {
	ctx, app, msgSrvr, k := s.ctx, s.app, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(types.GROUP_STATUS_ROUND_3)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			// Confirm the participation of each member in the group
			for _, m := range tc.Group.Members {
				_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
					GroupID:      tc.Group.ID,
					MemberID:     m.ID,
					OwnPubKeySig: m.PubKeySignature,
					Address:      sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			// Execute the EndBlocker to process groups
			app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})

			// Check the group's status and expiration time after confirmation
			got, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)
			s.Require().Equal(types.GROUP_STATUS_ACTIVE, got.Status)
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
					DEs:     deList,
					Address: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
				}
			},
			func() {},
			types.ErrDEQueueFull,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitDEs(ctx, &req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

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
					DEs:     []types.DE{de},
					Address: sdk.AccAddress(m.PubKey()).String(),
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

func (s *KeeperTestSuite) TestFailedRequestSignatureReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	var req *types.MsgRequestSignature
	var err error

	// Add failed case
	tcs := []TestCase{
		{
			"failure with invalid groupID",
			func() {
				req, err = types.NewMsgRequestSignature(
					tss.GroupID(999), // non-existent groupID
					types.NewTextRequestingSignature([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					testapp.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {},
			types.ErrGroupNotFound,
		},
		{
			"failure with inactive group",
			func() {
				inactiveGroup := types.Group{
					ID:        2,
					Size_:     5,
					Threshold: 3,
					PubKey:    nil,
					Fee:       sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Status:    types.GROUP_STATUS_FALLEN,
				}
				k.SetGroup(ctx, inactiveGroup)
				req, err = types.NewMsgRequestSignature(
					tss.GroupID(2), // inactive groupID
					types.NewTextRequestingSignature([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					testapp.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {},
			types.ErrGroupIsNotActive,
		},
		{
			"failure with not enough fee",
			func() {
				req, err = types.NewMsgRequestSignature(
					tss.GroupID(1),
					types.NewTextRequestingSignature([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					testapp.FeePayer.Address,
				)
			},
			func() {},
			types.ErrNotEnoughFee,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
			balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
				ctx,
				s.app.TSSKeeper.GetTSSAccount(ctx).GetAddress(),
			)

			_, err := msgSrvr.RequestSignature(ctx, req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
			balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
				ctx,
				s.app.TSSKeeper.GetTSSAccount(ctx).GetAddress(),
			)

			// Check if the balances of payer and module account doesn't change
			s.Require().Equal(balancesAfter, balancesBefore)
			s.Require().Equal(balancesModuleAfter, balancesModuleBefore)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessRequestSignatureReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		// Request signature for each member in the group
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, signing := range tc.Signings {
				balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
				balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
					ctx,
					s.app.TSSKeeper.GetTSSAccount(ctx).GetAddress(),
				)

				msg, err := types.NewMsgRequestSignature(
					tc.Group.ID,
					types.NewTextRequestingSignature(signing.Data),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					testapp.FeePayer.Address,
				)
				s.Require().NoError(err)

				_, err = msgSrvr.RequestSignature(ctx, msg)
				s.Require().NoError(err)

				// Fee should be paid after requesting signature
				balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
				balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
					ctx,
					s.app.TSSKeeper.GetTSSAccount(ctx).GetAddress(),
				)

				diff := sdk.NewCoins(sdk.NewInt64Coin("uband", int64(10*len(signing.AssignedMembers))))
				s.Require().Equal(diff, balancesBefore.Sub(balancesAfter...))
				s.Require().Equal(diff, balancesModuleAfter.Sub(balancesModuleBefore...))
			}
		})
	}
}

func (s *KeeperTestSuite) TestFailedSubmitSignatureReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

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
					Signature: tc1.Signings[0].Signature,
					Address:   sdk.AccAddress(tc1.Group.Members[0].PubKey()).String(),
				}
			},
			func() {},
			types.ErrSigningNotFound,
		},
		{
			"failure with invalid memberID",
			func() {
				k.SetSigning(ctx, types.Signing{
					ID:              tc1.Signings[0].ID,
					GroupID:         tc1.Group.ID,
					AssignedMembers: []types.AssignedMember{},
					Message:         tc1.Signings[0].Data,
					GroupPubNonce:   tc1.Signings[0].PubNonce,
					Status:          types.SIGNING_STATUS_WAITING,
					Signature:       nil,
				})

				req = types.MsgSubmitSignature{
					SigningID: tc1.Signings[0].ID,
					MemberID:  tss.MemberID(99), // non-existent memberID
					Signature: tc1.Signings[0].Signature,
					Address:   sdk.AccAddress(tc1.Group.Members[0].PubKey()).String(),
				}
			},
			func() {
				k.DeleteSigning(ctx, tc1.Signings[0].ID)
			},
			types.ErrMemberNotAssigned,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.SubmitSignature(ctx, &req)
			s.Require().ErrorIs(err, tc.ExpectedErr)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessSubmitSignatureReq() {
	ctx, app, msgSrvr, k := s.ctx, s.app, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Iterate through test cases from testutil
	for i, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			// Request signature for the first member in the group
			msg, err := types.NewMsgRequestSignature(
				tc.Group.ID,
				types.NewTextRequestingSignature([]byte("msg")),
				sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
				testapp.FeePayer.Address,
			)
			s.Require().NoError(err)
			_, err = msgSrvr.RequestSignature(ctx, msg)
			s.Require().NoError(err)

			// Get the signing information
			signing, err := k.GetSigning(ctx, tss.SigningID(i+1))
			s.Require().NoError(err)

			// Get the group information
			group, err := k.GetGroup(ctx, tc.Group.ID)
			s.Require().NoError(err)

			// Process signing for each assigned member
			var balancesBefores []sdk.Coins
			for _, am := range signing.AssignedMembers {
				// Compute Lagrange coefficient
				var lgc tss.Scalar
				lgc, err = tss.ComputeLagrangeCoefficient(
					am.MemberID,
					signing.AssignedMembers.MemberIDs(),
				)
				s.Require().NoError(err)

				// Compute private nonce
				pn, err := tss.ComputeOwnPrivNonce(PrivD, PrivE, am.BindingFactor)
				s.Require().NoError(err)

				// Sign the message
				signature, err := tss.SignSigning(
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
					Signature: signature,
					Address:   sdk.AccAddress(tc.Group.GetMember(am.MemberID).PubKey()).String(),
				})
				s.Require().NoError(err)

				balancesBefores = append(balancesBefores, s.app.BankKeeper.GetAllBalances(
					ctx,
					sdk.AccAddress(tc.Group.GetMember(am.MemberID).PubKey()),
				))
			}

			// Execute the EndBlocker to process signings
			app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})

			// Each assigned member should receive fee for the signature
			for i, am := range signing.AssignedMembers {
				balancesAfter := s.app.BankKeeper.GetAllBalances(
					ctx,
					sdk.AccAddress(tc.Group.GetMember(am.MemberID).PubKey()),
				)
				s.Require().Equal(signing.Fee, balancesAfter.Sub(balancesBefores[i]...))
			}

			// Retrieve the signing information after signing
			signing, err = k.GetSigning(ctx, tss.SigningID(i+1))
			s.Require().NoError(err)
			s.Require().NotNil(signing.Signature)
		})
	}
}

func (s *KeeperTestSuite) TestActivateReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestHealthCheckReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Iterate through test cases from testutil
	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				_, err := msgSrvr.HealthCheck(ctx, &types.MsgHealthCheck{
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParams() {
	k, msgSrvr := s.app.TSSKeeper, s.msgSrvr

	testCases := []struct {
		name         string
		request      *types.MsgUpdateParams
		expectErr    bool
		expectErrStr string
	}{
		{
			name: "set invalid authority",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr:    true,
			expectErrStr: "invalid authority;",
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params: types.Params{
					MaxGroupSize:            types.DefaultMaxGroupSize,
					MaxDESize:               types.DefaultMaxDESize,
					CreatingPeriod:          types.DefaultCreatingPeriod,
					SigningPeriod:           types.DefaultSigningPeriod,
					ActiveDuration:          types.DefaultActiveDuration,
					InactivePenaltyDuration: types.DefaultInactivePenaltyDuration,
					JailPenaltyDuration:     types.DefaultJailPenaltyDuration,
					RewardPercentage:        types.DefaultRewardPercentage,
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := msgSrvr.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErrStr)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
