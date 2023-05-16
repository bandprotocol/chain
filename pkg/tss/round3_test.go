package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestComputeOwnPublicKey() {
	pubKey, err := tss.ComputeOwnPublicKey(suite.points, uint32(suite.mid))
	suite.Require().NoError(err)
	suite.Require().
		Equal("023704dcdb774ed4fd0841ded5757211fe5a6f7637c4f9a1346b5b20e2524d12e5", hex.EncodeToString(pubKey))
}

func (suite *TSSTestSuite) TestComputeGroupPublicKey() {
	pubKey, err := tss.ComputeGroupPublicKey(suite.points)
	suite.Require().NoError(err)
	suite.Require().
		Equal("023704dcdb774ed4fd0841ded5757211fe5a6f7637c4f9a1346b5b20e2524d12e5", hex.EncodeToString(pubKey))
}

func (suite *TSSTestSuite) TestComputeOwnPrivateKey() {
	privKey := tss.ComputeOwnPrivateKey(suite.scalars)
	suite.Require().
		Equal("a537d3a6166c8efdf89d76bda2392e228f567ef12fe0632c0445e66cfdf53f02", hex.EncodeToString(privKey))
}

func (suite *TSSTestSuite) TestVerifySecretShare() {
	secret := tss.ComputeSecretShare(suite.scalars, uint32(suite.mid))
	err := tss.VerifySecretShare(suite.mid, secret, suite.points)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestComputeSecretShareCommit() {
	secretCommit, err := tss.ComputeSecretShareCommit(suite.points, uint32(suite.mid))
	suite.Require().NoError(err)
	suite.Require().
		Equal("023704dcdb774ed4fd0841ded5757211fe5a6f7637c4f9a1346b5b20e2524d12e5", hex.EncodeToString(secretCommit))
}

func (suite *TSSTestSuite) TestDecryptSecretShares() {
	expectedSecrets := tss.Scalars{
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(1)),
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(2)),
	}

	encSecrets, err := tss.EncryptSecretShares(
		expectedSecrets,
		tss.PublicKeys{suite.kpI.PublicKey, suite.kpJ.PublicKey},
	)
	suite.Require().NoError(err)

	secrets, err := tss.DecryptSecretShares(encSecrets, tss.PublicKeys{suite.kpI.PublicKey, suite.kpJ.PublicKey})
	suite.Require().NoError(err)

	suite.Require().Equal(expectedSecrets, secrets)
}

func (suite *TSSTestSuite) TestVerifyOwnPubKeySig() {
	// Sign
	sig, err := tss.SignOwnPublickey(suite.mid, suite.dkgContext, suite.kpI.PublicKey, suite.kpI.PrivateKey)
	suite.Require().NoError(err)

	// Success case
	err = tss.VerifyOwnPubKeySig(suite.mid, suite.dkgContext, sig, suite.kpI.PublicKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestVerifyComplainSig() {
	// Sign
	sig, keySym, nonceSym, err := tss.SignComplain(suite.kpI.PublicKey, suite.kpJ.PublicKey, suite.kpI.PrivateKey)
	suite.Require().NoError(err)

	// Success case
	err = tss.VerifyComplainSig(suite.kpI.PublicKey, suite.kpJ.PublicKey, keySym, nonceSym, sig)
	suite.Require().NoError(err)

	// Wrong public key I case
	err = tss.VerifyComplainSig(suite.fakeKp.PublicKey, suite.kpJ.PublicKey, keySym, nonceSym, sig)
	suite.Require().Error(err)

	// Wrong public key J case
	err = tss.VerifyComplainSig(suite.kpI.PublicKey, suite.fakeKp.PublicKey, keySym, nonceSym, sig)
	suite.Require().Error(err)

	// Wrong key sym case
	err = tss.VerifyComplainSig(suite.kpI.PublicKey, suite.kpJ.PublicKey, suite.fakeKp.PublicKey, nonceSym, sig)
	suite.Require().Error(err)

	// Wrong nonce sym case
	err = tss.VerifyComplainSig(suite.kpI.PublicKey, suite.kpJ.PublicKey, keySym, suite.fakeKp.PublicKey, sig)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestGenerateChallengeOwnPublicKey() {
	challenge := tss.GenerateChallengeOwnPublicKey(suite.mid, suite.dkgContext, suite.kpI.PublicKey)
	suite.Require().Equal(
		"726f756e64334f776e5075624b65790000000000000001646b67436f6e7465787403936f4b0644c78245124c19c9378e307cd955b227ee59c9ba16f4c7426c6418aa",
		hex.EncodeToString(challenge),
	)
}

func (suite *TSSTestSuite) TestGenerateChallengeComplain() {
	keySym, err := tss.ComputeKeySym(suite.kpI.PrivateKey, suite.kpJ.PublicKey)
	suite.Require().NoError(err)

	nonceSym, err := tss.ComputeNonceSym(tss.Scalar(suite.kpI.PrivateKey), suite.kpJ.PublicKey)
	suite.Require().NoError(err)

	challenge := tss.GenerateChallengeComplain(suite.kpI.PublicKey, suite.kpJ.PublicKey, keySym, nonceSym)
	suite.Require().Equal(
		"726f756e6433436f6d706c61696e03f70e80bac0b32b2599fa54d83b5471e90fac27bb09528f0337b49d464d64426f03f70e80bac0b32b2599fa54d83b5471e90fac27bb09528f0337b49d464d64426f03bc213e4251592d29c070e4c31b980d150e755ec848afa4c06730ec1dcd09c48202bc213e4251592d29c070e4c31b980d150e755ec848afa4c06730ec1dcd09c482",
		hex.EncodeToString(challenge),
	)
}
