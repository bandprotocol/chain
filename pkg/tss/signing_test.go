package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestComputeLagrangeCoefficient() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			value := tss.ComputeLagrangeCoefficient(assignedMember.ID, signing.GetAllIDs())
			suite.Require().Equal(assignedMember.Lagrange, value)
		})
}

func (suite *TSSTestSuite) TestComputeBytes() {
	suite.RunOnSigning(suite.testCases, func(tc testutil.TestCase, signing testutil.Signing) {
		bytes, err := tss.ComputeBytes(
			signing.GetAllIDs(),
			signing.GetAllPubDs(),
			signing.GetAllPubEs(),
		)
		suite.Require().NoError(err)
		suite.Require().Equal(signing.Bytes, bytes)
	})
}

func (suite *TSSTestSuite) TestComputeOwnLo() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			lo := tss.ComputeOwnLo(assignedMember.ID, signing.Data, signing.Bytes)
			suite.Require().Equal(assignedMember.Lo, lo)
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
				assignedMember.Lo,
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
				assignedMember.Lo,
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
		sig, err := tss.CombineSignatures(signing.GetAllSigs()...)
		suite.Require().NoError(err)
		suite.Require().Equal(signing.Sig, sig)
	})
}

func (suite *TSSTestSuite) TestSignSigning() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			sig, err := tss.SignSigning(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.PrivNonce,
				tc.Group.GetMember(assignedMember.ID).PrivKey,
			)
			suite.Require().NoError(err)
			suite.Require().Equal(assignedMember.Sig, sig)
		})
}

func (suite *TSSTestSuite) TestVerifySigningSig() {
	suite.RunOnAssignedMember(
		suite.testCases,
		func(tc testutil.TestCase, signing testutil.Signing, assignedMember testutil.AssignedMember) {
			// Success case
			err := tss.VerifySigningSig(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Sig,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().NoError(err)

			// Wrong public nonce case
			err = tss.VerifySigningSig(
				testutil.FakePubKey,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Sig,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().Error(err)

			// Wrong group public key case
			err = tss.VerifySigningSig(
				signing.PubNonce,
				testutil.FakePubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Sig,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().Error(err)

			// Wrong data case
			err = tss.VerifySigningSig(
				signing.PubNonce,
				tc.Group.PubKey,
				[]byte("fake data"),
				assignedMember.Lagrange,
				assignedMember.Sig,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().Error(err)

			// Wrong lagrange case
			err = tss.VerifySigningSig(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				testutil.FakeLagrange,
				assignedMember.Sig,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().Error(err)

			// Wrong signature case
			err = tss.VerifySigningSig(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				testutil.FakeSig,
				tc.Group.GetMember(assignedMember.ID).PubKey(),
			)
			suite.Require().Error(err)

			// Wrong own public key case
			err = tss.VerifySigningSig(
				signing.PubNonce,
				tc.Group.PubKey,
				signing.Data,
				assignedMember.Lagrange,
				assignedMember.Sig,
				testutil.FakePubKey,
			)
			suite.Require().Error(err)
		})
}

func (suite *TSSTestSuite) TestVerifyGroupSigningSig() {
	suite.RunOnSigning(suite.testCases, func(tc testutil.TestCase, signing testutil.Signing) {
		// Success case
		err := tss.VerifyGroupSigningSig(tc.Group.PubKey, signing.Data, signing.Sig)
		suite.Require().NoError(err)

		// Wrong group public key case
		err = tss.VerifyGroupSigningSig(testutil.FakePubKey, signing.Data, signing.Sig)
		suite.Require().Error(err)

		// Wrong data case
		err = tss.VerifyGroupSigningSig(tc.Group.PubKey, []byte("fake data"), signing.Sig)
		suite.Require().Error(err)

		// Wrong signature case
		err = tss.VerifyGroupSigningSig(tc.Group.PubKey, signing.Data, testutil.FakeSig)
		suite.Require().Error(err)
	})
}
