package tss_test

import (
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestComputeLagrangeCoefficient() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			value, err := tss.ComputeLagrangeCoefficient(assignedMember.ID, signing.GetAllIDs())
			suite.Require().NoError(err)
			suite.Require().Equal(assignedMember.Lagrange, value)
		})
}

func (suite *TSSTestSuite) TestComputeCommitment() {
	suite.RunOnSigning(suite.testCases, func(tc testutil.TestCase, signing testutil.Signing) {
		commitment, err := tss.ComputeCommitment(
			signing.GetAllIDs(),
			signing.GetAllPubDs(),
			signing.GetAllPubEs(),
		)
		suite.Require().NoError(err)
		suite.Require().Equal(signing.Commitment, commitment)
	})
}

func (suite *TSSTestSuite) TestComputeOwnBindingFactor() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			bindingFactor, err := tss.ComputeOwnBindingFactor(assignedMember.ID, signing.Data, signing.Commitment)
			suite.Require().NoError(err)
			suite.Require().Equal(assignedMember.BindingFactor, bindingFactor)
		},
	)
}

func (suite *TSSTestSuite) TestComputeOwnPublicNonce() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			pubNonce, err := tss.ComputeOwnPubNonce(
				assignedMember.PubD(),
				assignedMember.PubE(),
				assignedMember.BindingFactor,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(assignedMember.PubNonce(), pubNonce)
		},
	)
}

func (suite *TSSTestSuite) TestComputeOwnPrivateNonce() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			privNonce, err := tss.ComputeOwnPrivNonce(
				assignedMember.PrivD,
				assignedMember.PrivE,
				assignedMember.BindingFactor,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(assignedMember.PrivNonce, privNonce)
		},
	)
}

func (suite *TSSTestSuite) TestComputeGroupPublicNonce() {
	suite.RunOnSigning(suite.testCases, func(tc testutil.TestCase, signing testutil.Signing) {
		groupPubNonce, err := tss.ComputeGroupPublicNonce(signing.GetAllOwnPubNonces()...)
		suite.Require().NoError(err)
		suite.Require().Equal(signing.PubNonce, groupPubNonce)
	})
}

func (suite *TSSTestSuite) TestCombineSignatures() {
	suite.RunOnSigning(suite.testCases, func(tc testutil.TestCase, signing testutil.Signing) {
		signature, err := tss.CombineSignatures(signing.GetAllSignatures()...)
		suite.Require().NoError(err)
		suite.Require().Equal(signing.Signature, signature)
	})
}

func (suite *TSSTestSuite) TestSignSigning() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			signature, err := tss.SignSigning(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.PrivNonce,
				tc.Group.GetMember(assignedMember.ID).PrivKey,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(assignedMember.Signature, signature)
		})
}

func (suite *TSSTestSuite) TestVerifySigningSignature() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			// Success case
			err := tss.VerifySignature(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Signature,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().NoError(err)

			// Wrong public nonce case
			err = tss.VerifySignature(
				testutil.FalsePubKey,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Signature,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong group public key case
			err = tss.VerifySignature(
				signing.PubNonce,
				testutil.FalsePubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Signature,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong data case
			err = tss.VerifySignature(
				signing.PubNonce,
				tc.Group.PubKey,
				[]byte("fake data"),
				assignedMember.Lagrange,
				assignedMember.Signature,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong lagrange case
			err = tss.VerifySignature(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				testutil.FalseLagrange,
				assignedMember.Signature,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong signature case
			err = tss.VerifySignature(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				testutil.FalseSignature,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong own public key case
			err = tss.VerifySignature(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Signature,
				testutil.FalsePubKey,
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
		})
}

func (suite *TSSTestSuite) TestVerifyGroupSigningSignature() {
	suite.RunOnSigning(suite.testCases, func(tc testutil.TestCase, signing testutil.Signing) {
		// Success case
		err := tss.VerifyGroupSignature(tc.Group.PubKey, signing.Data, signing.Signature)
		suite.Require().NoError(err)

		// Wrong group public key case
		err = tss.VerifyGroupSignature(testutil.FalsePubKey, signing.Data, signing.Signature)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong data case
		err = tss.VerifyGroupSignature(tc.Group.PubKey, []byte("fake data"), signing.Signature)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyGroupSignature(tc.Group.PubKey, signing.Data, testutil.FalseSignature)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}
