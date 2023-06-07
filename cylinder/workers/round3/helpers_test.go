package round3

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/stretchr/testify/assert"
)

func TestGetOwnPrivKey(t *testing.T) {
	tests := []struct {
		name         string
		modify       func(*store.Group, *client.GroupResponse)
		expPrivKey   tss.PrivateKey
		expComplains []types.Complain
		expError     error
	}{
		{
			"success - private key",
			func(group *store.Group, groupRes *client.GroupResponse) {},
			hexDecode("b248a8a2f6f1644b196402de4026d3b63db36529b2b365995f5b21eebf20acea"),
			nil,
			nil,
		},
		{
			"success - complain",
			func(group *store.Group, groupRes *client.GroupResponse) {
				groupRes.AllRound2Data[1].EncryptedSecretShares[0] = hexDecode(
					"c3acf1cd68a4e9b00b0487fe9d6c44560487c6d463d410d1b9d81242f33c0e7a",
				)
			},
			nil,
			[]types.Complain{
				{
					I:        1,
					J:        2,
					KeySym:   hexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
					NonceSym: hexDecode("034946dba60574e576aef1c252e48db7c2c40f828efdb374ec8bd48ea36af06ac8"),
					Signature: hexDecode(
						"02a55f7d417d1b51d91e6097473f00f528291aaa0dd11733e83eb85680ed5d4e369fe3b8aef036713c547118f5a0adb8108dfe19b4067081f26a2fe27a87f60c0b",
					),
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			group, groupRes := getTestData()
			test.modify(&group, &groupRes)

			privKey, complains, err := getOwnPrivKey(group, &groupRes)
			assert.Equal(t, test.expPrivKey, privKey)
			assert.Equal(t, test.expComplains, complains)
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestGetSecretShare(t *testing.T) {
	tests := []struct {
		name           string
		modify         func(*store.Group, *client.GroupResponse)
		expSecretShare tss.Scalar
		expComplain    *types.Complain
		expError       error
	}{
		{
			"success - secret share",
			func(group *store.Group, groupRes *client.GroupResponse) {},
			hexDecode("dbc69d7d8fb753f3143e050a4d3fe01c35de8c5fe8937490dd9c5ccbf29567be"),
			nil,
			nil,
		},
		{
			"success - complain",
			func(group *store.Group, groupRes *client.GroupResponse) {
				groupRes.AllRound2Data[1].EncryptedSecretShares[0] = hexDecode(
					"c3acf1cd68a4e9b00b0487fe9d6c44560487c6d463d410d1b9d81242f33c0e7a",
				)
			},
			nil,
			&types.Complain{
				I:        1,
				J:        2,
				KeySym:   hexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
				NonceSym: hexDecode("034946dba60574e576aef1c252e48db7c2c40f828efdb374ec8bd48ea36af06ac8"),
				Signature: hexDecode(
					"02a55f7d417d1b51d91e6097473f00f528291aaa0dd11733e83eb85680ed5d4e369fe3b8aef036713c547118f5a0adb8108dfe19b4067081f26a2fe27a87f60c0b",
				),
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			group, groupRes := getTestData()
			test.modify(&group, &groupRes)

			secretShare, complain, err := getSecretShare(
				1,
				2,
				group.OneTimePrivKey,
				&groupRes,
			)
			assert.Equal(t, test.expSecretShare, secretShare)
			assert.Equal(t, test.expComplain, complain)
			assert.Equal(t, test.expError, err)
		})
	}
}

func getTestData() (store.Group, client.GroupResponse) {
	group := store.Group{
		MemberID:       1,
		OneTimePrivKey: hexDecode("83127264737dd61b4b7f8058a8418874f0e0e52ada48b39a497712a487096304"),
		Coefficients: tss.Scalars{
			hexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
			hexDecode("2611e629e7043dbd32682e91c73292b087bb9f0747cd4bdd94e92093b6716504"),
		},
	}

	groupRes := client.GroupResponse{
		QueryGroupResponse: types.QueryGroupResponse{
			Group: types.Group{
				Size_: 2,
			},
			AllRound1Data: []types.Round1Data{
				{
					MemberID: 1,
					CoefficientsCommit: tss.Points{
						hexDecode("039cff182d3b5653c215207c5b141983d6e784e51cc8088b7edfef6cba504573e3"),
						hexDecode("035b8a99ebc56c07b88404407046e6f0d5e5318a87b888ea25d6d12d8175b2d70c"),
					},
					OneTimePubKey: hexDecode(
						"0383764b806848430ed195ef8017fb4e768893ea07782e679c31e5ff1b8b453973",
					),
				},
				{
					MemberID: 2,
					CoefficientsCommit: tss.Points{
						hexDecode("02786741d28ca0a66b628d6401d975da448fc08c15a1228eb7b65203c6bac5cedb"),
						hexDecode("023d61b24c8785efe8c7459dc706d95b197c0acb31697feb49fec2d3446dc36de4"),
					},
					OneTimePubKey: hexDecode(
						"02e20b4d6bd3f10e7c3a9098c5832180b809a826ae49a972d5348758529c5015c5",
					),
				},
			},
			AllRound2Data: []types.Round2Data{
				{
					MemberID: 1,
					EncryptedSecretShares: tss.Scalars{
						hexDecode("d47a459f272be3d22e54af5a0a45ea8318e88f2c3c767962b2b5f9ba53d9922d")},
				},
				{
					MemberID: 2,
					EncryptedSecretShares: tss.Scalars{
						hexDecode("b3acf1cd68a4e9b00b0487fe9d6c44560487c6d463d410d1b9d81242f33c0e7a")},
				},
			}},
	}

	return group, groupRes
}

func hexDecode(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}
