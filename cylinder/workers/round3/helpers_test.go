package round3

import (
	"fmt"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/stretchr/testify/assert"
)

func TestGetOwnPrivKey(t *testing.T) {
	tests := []struct {
		name          string
		modify        func(*store.Group, *client.GroupResponse, tss.MemberID)
		expPrivKey    bool
		expComplaints bool
		expErr        bool
	}{
		{
			"success - private key",
			func(group *store.Group, groupRes *client.GroupResponse, mid tss.MemberID) {},
			true, false, false,
		},
		{
			"success - complaint",
			func(group *store.Group, groupRes *client.GroupResponse, mid tss.MemberID) {
				for _, round2Info := range groupRes.Round2Infos {
					if round2Info.MemberID != mid {
						round2Info.EncryptedSecretShares[testutil.GetSlot(round2Info.MemberID, mid)] = testutil.HexDecode(
							"0000000000000000000000000000000000000000000000000000000000000000",
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
					group, groupRes := getTestData(tc, member)
					test.modify(&group, &groupRes, member.ID)

					privKey, complaints, err := getOwnPrivKey(group, &groupRes)

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
								Complainer:  member.ID,
								Complainant: m.ID,
								KeySym:      member.KeySyms[slot],
								Signature:   member.ComplaintSigs[slot],
							}

							assert.Equal(t, expComplaint.Complainer, complaints[slot].Complainer)
							assert.Equal(t, expComplaint.Complainant, complaints[slot].Complainant)
							assert.Equal(t, expComplaint.KeySym, complaints[slot].KeySym)

							// Can't compare signature as the nonce will be randomly generated
							err := tss.VerifyComplaintSig(
								member.OneTimePubKey(),
								m.OneTimePubKey(),
								expComplaint.KeySym,
								complaints[slot].Signature,
							)
							assert.Nil(t, err)
						}
					} else {
						assert.Nil(t, complaints)
					}

					if test.expErr {
						assert.Error(t, err)
					} else {
						assert.Nil(t, err)
					}
				})
			}
		}
	}
}

func TestGetSecretShare(t *testing.T) {
	tests := []struct {
		name           string
		modify         func(*store.Group, *client.GroupResponse, tss.MemberID, tss.MemberID)
		expSecretShare bool
		expComplaint   bool
		expErr         bool
	}{
		{
			"success - secret share",
			func(group *store.Group, groupRes *client.GroupResponse, i tss.MemberID, j tss.MemberID) {},
			true, false, false,
		},
		{
			"success - complaint",
			func(group *store.Group, groupRes *client.GroupResponse, i tss.MemberID, j tss.MemberID) {
				for _, round2Info := range groupRes.Round2Infos {
					if round2Info.MemberID == j {
						round2Info.EncryptedSecretShares[testutil.GetSlot(j, i)] = testutil.HexDecode(
							"0000000000000000000000000000000000000000000000000000000000000000",
						)
					}
				}
			},
			false, true, false,
		},
	}

	for _, test := range tests {
		for _, tc := range testutil.TestCases {
			for _, memberI := range tc.Group.Members {
				for _, memberJ := range tc.Group.Members {
					if memberI.ID == memberJ.ID {
						continue
					}

					t.Run(
						fmt.Sprintf(
							"%s, Test: (%s), MemberI: %d, MemberJ: %d",
							test.name,
							tc.Name,
							memberI.ID,
							memberJ.ID,
						),
						func(t *testing.T) {
							group, groupRes := getTestData(tc, memberI)
							test.modify(&group, &groupRes, memberI.ID, memberJ.ID)

							secretShare, complaint, err := getSecretShare(
								memberI.ID,
								memberJ.ID,
								group.OneTimePrivKey,
								&groupRes,
							)

							if test.expSecretShare {
								assert.Nil(t, complaint)
								assert.Nil(t, err)
								assert.Equal(
									t,
									memberJ.SecretShares[testutil.GetSlot(memberJ.ID, memberI.ID)],
									secretShare,
								)
							} else {
								assert.Nil(t, secretShare)
							}

							if test.expComplaint {
								slot := testutil.GetSlot(memberI.ID, memberJ.ID)
								expComplaint := &types.Complaint{
									Complainer:  memberI.ID,
									Complainant: memberJ.ID,
									KeySym:      memberI.KeySyms[slot],
									Signature:   memberI.ComplaintSigs[slot],
								}

								assert.Equal(t, expComplaint.Complainer, complaint.Complainer)
								assert.Equal(t, expComplaint.Complainant, complaint.Complainant)
								assert.Equal(t, expComplaint.KeySym, complaint.KeySym)

								// Can't compare signature as the nonce will be randomly generated
								err := tss.VerifyComplaintSig(
									memberI.OneTimePubKey(),
									memberJ.OneTimePubKey(),
									expComplaint.KeySym,
									complaint.Signature,
								)
								assert.Nil(t, err)
							} else {
								assert.Nil(t, complaint)
							}

							if test.expErr {
								assert.Error(t, err)
							} else {
								assert.Nil(t, err)
							}
						},
					)
				}
			}
		}
	}
}

func getTestData(testCase testutil.TestCase, member testutil.Member) (store.Group, client.GroupResponse) {
	tc := testutil.CopyTestCase(testCase)

	group := store.Group{
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
		round1Info := types.Round1Info{
			MemberID:           m.ID,
			CoefficientCommits: m.CoefficientCommits,
			OneTimePubKey:      m.OneTimePubKey(),
		}
		groupRes.Round1Infos = append(groupRes.Round1Infos, round1Info)

		round2Info := types.Round2Info{
			MemberID:              m.ID,
			EncryptedSecretShares: m.EncSecretShares,
		}
		groupRes.Round2Infos = append(groupRes.Round2Infos, round2Info)
	}

	return group, groupRes
}
