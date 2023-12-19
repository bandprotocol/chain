package group

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetOwnPrivKey(t *testing.T) {
	tests := []struct {
		name          string
		modify        func(*store.DKG, *client.GroupResponse, tss.MemberID)
		expPrivKey    bool
		expComplaints bool
		expErr        bool
	}{
		{
			"success - without malicious member",
			func(group *store.DKG, groupRes *client.GroupResponse, mid tss.MemberID) {},
			true, false, false,
		},
		{
			"success - with malicious member",
			func(group *store.DKG, groupRes *client.GroupResponse, mid tss.MemberID) {
				for _, r2Info := range groupRes.Round2Infos {
					if r2Info.MemberID != mid {
						r2Info.EncryptedSecretShares[testutil.GetSlot(r2Info.MemberID, mid)] = testutil.HexDecode(
							"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
						)
					}
				}
			},
			false, true, false,
		},
	}

	for _, test := range tests {
		for _, tc := range testutil.TestCases {
			for _, member := range tc.Group.Members {
				t.Run(fmt.Sprintf("%s, Test: %s, Member: %d", test.name, tc.Name, member.ID), func(t *testing.T) {
					dkg, groupRes := getTestData(tc, member)
					test.modify(&dkg, &groupRes, member.ID)

					privKey, complaints, err := getOwnPrivKey(dkg, &groupRes)

					if test.expPrivKey {
						assert.Equal(t, member.PrivKey, privKey)
					} else {
						assert.Nil(t, privKey)
					}

					if test.expComplaints {
						for _, m := range tc.Group.Members {
							if m.ID == member.ID {
								continue
							}

							slot := testutil.GetSlot(member.ID, m.ID)
							expComplaint := types.Complaint{
								Complainant: member.ID,
								Respondent:  m.ID,
								KeySym:      member.KeySyms[slot],
								Signature:   member.ComplaintSignatures[slot],
							}

							assert.Equal(t, expComplaint.Complainant, complaints[slot].Complainant)
							assert.Equal(t, expComplaint.Respondent, complaints[slot].Respondent)
							assert.Equal(t, expComplaint.KeySym, complaints[slot].KeySym)

							// Can't compare signature as the nonce will be randomly generated
							err := tss.VerifyComplaintSignature(
								member.OneTimePubKey(),
								m.OneTimePubKey(),
								expComplaint.KeySym,
								complaints[slot].Signature,
							)
							assert.NoError(t, err)
						}
					} else {
						assert.Nil(t, complaints)
					}

					if test.expErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		}
	}
}

func TestGetSecretShare(t *testing.T) {
	tests := []struct {
		name           string
		modify         func(*store.DKG, *client.GroupResponse, tss.MemberID, tss.MemberID)
		expSecretShare bool
		expComplaint   bool
		expErr         bool
	}{
		{
			"success - without malicious member",
			func(dkg *store.DKG, groupRes *client.GroupResponse, i tss.MemberID, j tss.MemberID) {},
			true, false, false,
		},
		{
			"success - with malicious member",
			func(dkg *store.DKG, groupRes *client.GroupResponse, i tss.MemberID, j tss.MemberID) {
				for _, r2Info := range groupRes.Round2Infos {
					if r2Info.MemberID == j {
						r2Info.EncryptedSecretShares[testutil.GetSlot(j, i)] = testutil.HexDecode(
							"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
						)
					}
				}
			},
			false, true, false,
		},
	}

	for _, test := range tests {
		for _, tc := range testutil.TestCases {
			for _, sender := range tc.Group.Members {
				for _, receiver := range tc.Group.Members {
					if sender.ID == receiver.ID {
						continue
					}

					t.Run(
						fmt.Sprintf(
							"%s, Test: (%s), Receiver: %d, Sender: %d",
							test.name,
							tc.Name,
							receiver,
							sender,
						),
						func(t *testing.T) {
							dkg, groupRes := getTestData(tc, receiver)
							test.modify(&dkg, &groupRes, receiver.ID, sender.ID)

							secretShare, complaint, err := getSecretShare(
								receiver.ID,
								sender.ID,
								dkg.OneTimePrivKey,
								&groupRes,
							)

							if test.expSecretShare {
								assert.Nil(t, complaint)
								assert.NoError(t, err)
								assert.Equal(
									t,
									sender.SecretShares[testutil.GetSlot(sender.ID, receiver.ID)],
									secretShare,
								)
							} else {
								assert.Nil(t, secretShare)
							}

							if test.expComplaint {
								slot := testutil.GetSlot(receiver.ID, sender.ID)
								expComplaint := &types.Complaint{
									Complainant: receiver.ID,
									Respondent:  sender.ID,
									KeySym:      receiver.KeySyms[slot],
									Signature:   receiver.ComplaintSignatures[slot],
								}

								assert.Equal(t, expComplaint.Complainant, complaint.Complainant)
								assert.Equal(t, expComplaint.Respondent, complaint.Respondent)
								assert.Equal(t, expComplaint.KeySym, complaint.KeySym)

								// Can't compare signature as the nonce will be randomly generated
								err := tss.VerifyComplaintSignature(
									receiver.OneTimePubKey(),
									sender.OneTimePubKey(),
									expComplaint.KeySym,
									complaint.Signature,
								)
								assert.NoError(t, err)
							} else {
								assert.Nil(t, complaint)
							}

							if test.expErr {
								assert.Error(t, err)
							} else {
								assert.NoError(t, err)
							}
						},
					)
				}
			}
		}
	}
}

func getTestData(testCase testutil.TestCase, member testutil.Member) (store.DKG, client.GroupResponse) {
	tc := testutil.CopyTestCase(testCase)

	dkg := store.DKG{
		GroupID:        tc.Group.ID,
		MemberID:       member.ID,
		OneTimePrivKey: member.OneTimePrivKey,
		Coefficients:   member.Coefficients,
	}

	groupRes := client.GroupResponse{
		QueryGroupResponse: types.QueryGroupResponse{
			Group: types.Group{
				Size_: uint64(tc.Group.GetSize()),
			},
			Round1Infos: []types.Round1Info{},
			Round2Infos: []types.Round2Info{},
		},
	}

	for _, m := range tc.Group.Members {
		r1Info := types.Round1Info{
			MemberID:           m.ID,
			CoefficientCommits: m.CoefficientCommits,
			OneTimePubKey:      m.OneTimePubKey(),
		}
		groupRes.Round1Infos = append(groupRes.Round1Infos, r1Info)

		r2Info := types.Round2Info{
			MemberID:              m.ID,
			EncryptedSecretShares: m.EncSecretShares,
		}
		groupRes.Round2Infos = append(groupRes.Round2Infos, r2Info)
	}

	return dkg, groupRes
}
