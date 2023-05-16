package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func (suite *TSSTestSuite) TestGenerateRound1Data() {
	data, err := tss.GenerateRound1Data(suite.mid, suite.threshold, suite.dkgContext)
	suite.Require().NoError(err)

	err = tss.VerifyOneTimeSig(suite.mid, suite.dkgContext, data.OneTimeSig, data.OneTimePubKey)
	suite.Require().NoError(err)

	err = tss.VerifyA0Sig(suite.mid, suite.dkgContext, data.A0Sig, data.A0PubKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestVerifyOneTimeSig() {
	// Sign
	sig, err := tss.SignOneTime(suite.mid, suite.dkgContext, suite.kpI.PublicKey, suite.kpI.PrivateKey)
	suite.Require().NoError(err)

	// Success case
	err = tss.VerifyOneTimeSig(suite.mid, suite.dkgContext, sig, suite.kpI.PublicKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestGenerateChallengeOneTime() {
	challenge := tss.GenerateChallengeOneTime(suite.mid, suite.dkgContext, suite.kpI.PublicKey)
	assert.Equal(
		suite.T(),
		"726f756e64314f6e6554696d650000000000000001646b67436f6e7465787403936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa",
		hex.EncodeToString(challenge),
	)
}

func (suite *TSSTestSuite) TestVerifyA0Sig() {
	// Sign
	sig, err := tss.SignA0(suite.mid, suite.dkgContext, suite.kpI.PublicKey, suite.kpI.PrivateKey)
	suite.Require().NoError(err)

	// Success case
	err = tss.VerifyA0Sig(suite.mid, suite.dkgContext, sig, suite.kpI.PublicKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestGenerateChallengeA0() {
	challenge := tss.GenerateChallengeA0(suite.mid, suite.dkgContext, suite.kpI.PublicKey)
	assert.Equal(
		suite.T(),
		"726f756e643141300000000000000001646b67436f6e7465787403936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa",
		hex.EncodeToString(challenge),
	)
}
