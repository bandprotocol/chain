package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestComputeEncryptedSecretShares() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		var pubKeys tss.Points
		for _, m := range tc.Group.Members {
			pubKeys = append(pubKeys, m.OneTimePubKey())
		}

		var encSecretShares tss.EncSecretShares

		counter := 0
		for idx := range pubKeys {
			if idx+1 == int(member.ID) {
				continue
			}
			encSecretShare, err := tss.ComputeEncryptedSecretShares(
				member.ID,
				member.OneTimePrivKey,
				pubKeys,
				member.Coefficients,
				testutil.MockNonce16Generator{
					MockGenerateFunc: func() ([]byte, error) {
						return member.EncSecretShares[counter].Nonce(), nil
					},
				},
			)
			suite.Require().NoError(err)

			encSecretShares = append(encSecretShares, encSecretShare[counter])
			counter++
		}

		suite.Require().Equal(member.EncSecretShares, encSecretShares)
	})
}

func (suite *TSSTestSuite) TestEncryptSecretShares() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		var encSecretShares tss.EncSecretShares
		counter := 0
		for idx := range tc.Group.Members {
			if idx+1 == int(member.ID) {
				continue
			}

			encSecretShare, err := tss.EncryptSecretShares(
				member.SecretShares,
				member.KeySyms,
				testutil.MockNonce16Generator{
					MockGenerateFunc: func() ([]byte, error) {
						return member.EncSecretShares[counter].Nonce(), nil
					},
				},
			)
			suite.Require().NoError(err)

			encSecretShares = append(encSecretShares, encSecretShare[counter])
			counter++
		}

		suite.Require().Equal(member.EncSecretShares, encSecretShares)
	})
}

func (suite *TSSTestSuite) TestComputeSecretShare() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			secret, err := tss.ComputeSecretShare(memberI.Coefficients, memberJ.ID)
			suite.Require().NoError(err)
			suite.Require().Equal(memberI.SecretShares[testutil.GetSlot(memberI.ID, memberJ.ID)], secret)
		},
	)
}
