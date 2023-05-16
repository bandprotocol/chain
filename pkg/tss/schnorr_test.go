package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestVerify() {
	// Sign
	sig, err := tss.Sign(suite.kpI.PrivateKey, suite.challenge, nil)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(sig, suite.challenge, suite.kpI.PublicKey, nil, nil)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(sig, append(suite.challenge, []byte("a")...), suite.kpI.PublicKey, nil, nil)
	suite.Require().Error(err)

	// Wrong public key case
	err = tss.Verify(sig, suite.challenge, suite.fakeKp.PublicKey, nil, nil)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestVerifyWithCustomNonce() {
	// Sign
	sig, err := tss.Sign(suite.kpI.PrivateKey, suite.challenge, suite.nonce)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(sig, suite.challenge, suite.kpI.PublicKey, nil, nil)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(sig, append(suite.challenge, []byte("a")...), suite.kpI.PublicKey, nil, nil)
	suite.Require().Error(err)

	// Wrong public key case
	err = tss.Verify(sig, suite.challenge, suite.fakeKp.PublicKey, nil, nil)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestVerifyWithCustomNonceAndGenerator() {
	// Prepare
	generator := []byte(suite.kpJ.PublicKey)
	fakeGenerator := []byte(suite.fakeKp.PublicKey)

	// Sign
	sig, err := tss.Sign(suite.kpI.PrivateKey, suite.challenge, suite.nonce)
	suite.Require().NoError(err)

	keySym, err := tss.ComputeKeySym(suite.kpI.PrivateKey, generator)
	suite.Require().NoError(err)

	nonceSym, err := tss.ComputeNonceSym(suite.nonce, generator)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(sig, suite.challenge, keySym, generator, nonceSym)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(sig, append(suite.challenge, []byte("a")...), keySym, generator, nonceSym)
	suite.Require().Error(err)

	// Wrong key sym case
	err = tss.Verify(sig, suite.challenge, suite.fakeKp.PublicKey, generator, nonceSym)
	suite.Require().Error(err)

	// Wrong generator case
	err = tss.Verify(sig, suite.challenge, keySym, fakeGenerator, nonceSym)
	suite.Require().Error(err)

	// Wrong nonce sym case
	err = tss.Verify(sig, suite.challenge, keySym, fakeGenerator, suite.fakeKp.PublicKey)
	suite.Require().Error(err)
}
