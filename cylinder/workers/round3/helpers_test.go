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
		name         string
		modify       func(*store.Group, *client.GroupResponse, tss.MemberID)
		expPrivKey   bool
		expComplains bool
		expErr       bool
	}{
		{
			"success - private key",
			func(group *store.Group, groupRes *client.GroupResponse, mid tss.MemberID) {},
			true, false, false,
		},
		{
			"success - complain",
			func(group *store.Group, groupRes *client.GroupResponse, mid tss.MemberID) {
				for _, round2Data := range groupRes.AllRound2Data {
					if round2Data.MemberID != mid {
						round2Data.EncryptedSecretShares[testutil.GetSlot(round2Data.MemberID, mid)] = testutil.HexDecode(
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

					privKey, complains, err := getOwnPrivKey(group, &groupRes)

					if test.expPrivKey {
						assert.Equal(t, member.PrivKey, privKey)
					} else {
						assert.Nil(t, privKey)
					}

					if test.expComplains {
						var expComplains []types.Complain

						for _, m := range tc.Group.Members {
							if m.ID == member.ID {
								continue
							}

							slot := testutil.GetSlot(member.ID, m.ID)
							expComplains = append(expComplains, types.Complain{
								I:      member.ID,
								J:      m.ID,
								KeySym: member.KeySyms[slot],
								Sig:    member.ComplainSigs[slot],
							})
						}

						assert.Equal(t, expComplains, complains)
					} else {
						assert.Nil(t, complains)
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
		expComplain    bool
		expErr         bool
	}{
		{
			"success - secret share",
			func(group *store.Group, groupRes *client.GroupResponse, i tss.MemberID, j tss.MemberID) {},
			true, false, false,
		},
		{
			"success - complain",
			func(group *store.Group, groupRes *client.GroupResponse, i tss.MemberID, j tss.MemberID) {
				for _, round2Data := range groupRes.AllRound2Data {
					if round2Data.MemberID == j {
						round2Data.EncryptedSecretShares[testutil.GetSlot(j, i)] = testutil.HexDecode(
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

							secretShare, complain, err := getSecretShare(
								memberI.ID,
								memberJ.ID,
								group.OneTimePrivKey,
								&groupRes,
							)

							if test.expSecretShare {
								assert.Nil(t, complain)
								assert.Nil(t, err)
								assert.Equal(
									t,
									memberJ.SecretShares[testutil.GetSlot(memberJ.ID, memberI.ID)],
									secretShare,
								)
							} else {
								assert.Nil(t, secretShare)
							}

							if test.expComplain {
								slot := testutil.GetSlot(memberI.ID, memberJ.ID)
								expComplain := &types.Complain{
									I:      memberI.ID,
									J:      memberJ.ID,
									KeySym: memberI.KeySyms[slot],
									Sig:    memberI.ComplainSigs[slot],
								}

								assert.Equal(t, expComplain, complain)
							} else {
								assert.Nil(t, complain)
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
			AllRound1Data: []types.Round1Data{},
			AllRound2Data: []types.Round2Data{},
		},
	}

	for _, m := range tc.Group.Members {
		round1Data := types.Round1Data{
			MemberID:           m.ID,
			CoefficientsCommit: m.CoefficientsCommit,
			OneTimePubKey:      m.OneTimePubKey(),
		}
		groupRes.AllRound1Data = append(groupRes.AllRound1Data, round1Data)

		round2Data := types.Round2Data{
			MemberID:              m.ID,
			EncryptedSecretShares: m.EncSecretShares,
		}
		groupRes.AllRound2Data = append(groupRes.AllRound2Data, round2Data)
	}

	return group, groupRes
}
