package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestComputeOwnPublicKey() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		accCommits, err := tc.Group.GetAccumulatedCommits()
		suite.Require().NoError(err)

		pubKey, err := tss.ComputeOwnPublicKey(
			accCommits,
			member.ID,
		)
		suite.Require().NoError(err)
		suite.Require().Equal(member.PubKey(), pubKey)
	})
}

func (suite *TSSTestSuite) TestComputeGroupPublicKey() {
	for _, tc := range suite.testCases {
		suite.Run(tc.Name, func() {
			pubKey, err := tss.ComputeGroupPublicKey(tc.Group.GetCommits(0)...)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.Group.PubKey, pubKey)
		})
	}
}

func (suite *TSSTestSuite) TestComputeOwnPrivateKey() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		var allSecrets []tss.Scalar
		ownSecret, err := tss.ComputeSecretShare(member.Coefficients, member.ID)
		allSecrets = append(allSecrets, ownSecret)

		for _, m := range tc.Group.Members {
			if m.ID == member.ID {
				continue
			}
			allSecrets = append(allSecrets, m.SecretShares[testutil.GetSlot(m.ID, member.ID)])
		}

		privKey, err := tss.ComputeOwnPrivateKey(allSecrets...)
		suite.Require().NoError(err)
		suite.Require().Equal(member.PrivKey, privKey)
	})
}

func (suite *TSSTestSuite) TestVerifySecretShare() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			err := tss.VerifySecretShare(
				memberJ.ID,
				memberI.SecretShares[testutil.GetSlot(memberI.ID, memberJ.ID)],
				memberI.CoefficientCommits,
			)
			suite.Require().NoError(err)
		},
	)
}

func (suite *TSSTestSuite) TestComputeSecretShareCommit() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			secretCommit, err := tss.ComputeSecretShareCommit(memberI.CoefficientCommits, memberJ.ID)
			suite.Require().NoError(err)

			expSecretCommit := memberI.SecretShares[testutil.GetSlot(memberI.ID, memberJ.ID)].Point()
			suite.Require().Equal(tss.Point(expSecretCommit), secretCommit)
		},
	)
}

func (suite *TSSTestSuite) TestDecryptSecretShares() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		secretShares, err := tss.DecryptSecretShares(member.EncSecretShares, member.KeySyms)
		suite.Require().NoError(err)
		suite.Require().Equal(member.SecretShares, secretShares)
	})
}

func (suite *TSSTestSuite) TestDecryptSecretShare() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		for i, encSecretShare := range member.EncSecretShares {
			secretShare, err := tss.DecryptSecretShare(encSecretShare, member.KeySyms[i])
			suite.Require().NoError(err)
			suite.Require().Equal(member.SecretShares[i], secretShare)
		}
	})
}

func (suite *TSSTestSuite) TestSignOwnPubKey() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		signature, err := tss.SignOwnPubkey(
			member.ID,
			tc.Group.DKGContext,
			member.PubKey(),
			member.PrivKey,
		)
		suite.Require().NoError(err)

		err = tss.VerifyOwnPubKeySignature(member.ID, tc.Group.DKGContext, signature, member.PubKey())
		suite.Require().NoError(err)
	})
}

func (suite *TSSTestSuite) TestVerifyOwnPubKeySignature() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Success case
		err := tss.VerifyOwnPubKeySignature(member.ID, tc.Group.DKGContext, member.PubKeySignature, member.PubKey())
		suite.Require().NoError(err)

		// Wrong ID case
		err = tss.VerifyOwnPubKeySignature(0, tc.Group.DKGContext, member.PubKeySignature, member.PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong DKGContext case
		err = tss.VerifyOwnPubKeySignature(
			member.ID,
			[]byte("false DKGContext"),
			member.PubKeySignature,
			member.PubKey(),
		)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyOwnPubKeySignature(member.ID, tc.Group.DKGContext, testutil.FalseSignature, member.PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong public key case
		err = tss.VerifyOwnPubKeySignature(
			member.ID,
			tc.Group.DKGContext,
			member.PubKeySignature,
			testutil.FalsePubKey,
		)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}

func (suite *TSSTestSuite) TestSignComplaint() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			signature, keySym, err := tss.SignComplaint(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.OneTimePrivKey,
			)
			suite.Require().NoError(err)

			err = tss.VerifyComplaintSignature(memberI.OneTimePubKey(), memberJ.OneTimePubKey(), keySym, signature)
			suite.Require().NoError(err)

			suite.Require().
				Equal(memberI.KeySyms[testutil.GetSlot(memberI.ID, memberJ.ID)], keySym)
		})
}

func (suite *TSSTestSuite) TestVerifyComplaint() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			iSlot := testutil.GetSlot(memberI.ID, memberJ.ID)
			jSlot := testutil.GetSlot(memberJ.ID, memberI.ID)
			// Success case - wrong encrypted secret share
			falseEncSecretShare := testutil.HexDecode(
				"9a4b4aff91200c9c8604fc218f67c35796d0aba5a2e277c46a01140dc4ff24b600939f506aa1550df8a0cf08db8f00d3",
			)
			err := tss.VerifyComplaint(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[iSlot],
				memberI.ComplaintSignatures[iSlot],
				falseEncSecretShare,
				memberI.ID,
				memberJ.CoefficientCommits,
			)
			suite.Require().NoError(err)

			// Failed case - correct encrypted secret share
			err = tss.VerifyComplaint(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[iSlot],
				memberI.ComplaintSignatures[iSlot],
				memberJ.EncSecretShares[jSlot],
				memberI.ID,
				memberJ.CoefficientCommits,
			)
			suite.Require().ErrorIs(err, tss.ErrValidSecretShare)
		})
}

func (suite *TSSTestSuite) TestVerifyComplaintSignature() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			slot := testutil.GetSlot(memberI.ID, memberJ.ID)
			// Success case
			err := tss.VerifyComplaintSignature(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[slot],
				memberI.ComplaintSignatures[slot],
			)
			suite.Require().NoError(err)

			// Wrong public key I case
			err = tss.VerifyComplaintSignature(
				testutil.FalsePubKey,
				memberJ.OneTimePubKey(),
				memberI.KeySyms[slot],
				memberI.ComplaintSignatures[slot],
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong public key J case
			err = tss.VerifyComplaintSignature(
				memberI.OneTimePubKey(),
				testutil.FalsePubKey,
				memberI.KeySyms[slot],
				memberI.ComplaintSignatures[slot],
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong key sym case
			err = tss.VerifyComplaintSignature(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				testutil.FalsePubKey,
				memberI.ComplaintSignatures[slot],
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong signature case
			err = tss.VerifyComplaintSignature(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[slot],
				testutil.FalseComplaintSignature,
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
		})
}
