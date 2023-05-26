package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestComputeOwnPublicKey() {
	point1, err := tss.SumPoints(suite.member1.CoefficientsCommit[0], suite.member2.CoefficientsCommit[0])
	suite.Require().NoError(err)
	point2, err := tss.SumPoints(suite.member1.CoefficientsCommit[1], suite.member2.CoefficientsCommit[1])
	suite.Require().NoError(err)

	pubKey, err := tss.ComputeOwnPublicKey(
		tss.Points{point1, point2},
		suite.member1.mid,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.ownKey.PublicKey, pubKey)
}

func (suite *TSSTestSuite) TestComputeGroupPublicKey() {
	pubKey, err := tss.ComputeGroupPublicKey(
		suite.member1.CoefficientsCommit[0], suite.member2.CoefficientsCommit[0],
	)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.groupPubKey, pubKey)
}

func (suite *TSSTestSuite) TestComputeOwnPrivateKey() {
	ownSecret, err := tss.ComputeSecretShare(suite.member1.Coefficients, 1)
	privKey, err := tss.ComputeOwnPrivateKey(ownSecret, suite.member2.secretShares[0])
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.ownKey.PrivateKey, privKey)
}

func (suite *TSSTestSuite) TestVerifySecretShare() {
	secret, err := tss.ComputeSecretShare(suite.member1.Coefficients, suite.member1.mid)
	suite.Require().NoError(err)

	err = tss.VerifySecretShare(suite.member1.mid, secret, suite.member1.CoefficientsCommit)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestComputeSecretShareCommit() {
	secretCommit, err := tss.ComputeSecretShareCommit(suite.member1.CoefficientsCommit, suite.member1.mid)
	suite.Require().NoError(err)
	suite.Require().
		Equal("033c55180665bd1ec4467a3872da2f574f9eafe2aa28a605315b49860a5215f849", hex.EncodeToString(secretCommit))
}

func (suite *TSSTestSuite) TestDecryptSecretShares() {
	expectedSecrets := tss.Scalars{
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(1)),
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(2)),
	}

	pubKeys := tss.PublicKeys{suite.member1.OneTimePubKey, suite.member2.OneTimePubKey}

	encSecrets, err := tss.EncryptSecretShares(
		expectedSecrets,
		pubKeys,
	)
	suite.Require().NoError(err)

	secrets, err := tss.DecryptSecretShares(encSecrets, pubKeys)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedSecrets, secrets)
}

func (suite *TSSTestSuite) TestSignAndVerifyOwnPubKey() {
	// Sign
	sig, err := tss.SignOwnPublickey(
		suite.member1.mid,
		suite.groupDKGContext,
		suite.member1.ownKey.PublicKey,
		suite.member1.ownKey.PrivateKey,
	)
	suite.Require().NoError(err)

	// Success case
	err = tss.VerifyOwnPubKeySig(suite.member1.mid, suite.groupDKGContext, sig, suite.member1.ownKey.PublicKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestGenerateMessageOwnPublicKey() {
	challenge := tss.GenerateMessageOwnPublicKey(
		suite.member1.mid,
		suite.groupDKGContext,
		suite.member1.ownKey.PublicKey,
	)
	suite.Require().Equal(
		"726f756e64334f776e5075624b65790000000000000001a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c0268c34a74f75ea26f3eba73a44afdaaa5e4704baa6f58d6e1ab831a5608e4dae4",
		hex.EncodeToString(challenge),
	)
}
func (suite *TSSTestSuite) TestSignAndVerifyComplain() {
	// Sign
	sig, keySym, nonceSym, err := tss.SignComplain(
		suite.member1.OneTimePubKey,
		suite.member2.OneTimePubKey,
		suite.member1.OneTimePrivKey,
	)
	suite.Require().NoError(err)

	// Success case
	err = tss.VerifyComplainSig(suite.member1.OneTimePubKey, suite.member2.OneTimePubKey, keySym, nonceSym, sig)
	suite.Require().NoError(err)

	// Wrong public key I case
	err = tss.VerifyComplainSig(suite.fakeKey.PublicKey, suite.member2.OneTimePubKey, keySym, nonceSym, sig)
	suite.Require().Error(err)

	// Wrong public key J case
	err = tss.VerifyComplainSig(suite.member1.OneTimePubKey, suite.fakeKey.PublicKey, keySym, nonceSym, sig)
	suite.Require().Error(err)

	// Wrong key sym case
	err = tss.VerifyComplainSig(
		suite.member1.OneTimePubKey,
		suite.member2.OneTimePubKey,
		suite.fakeKey.PublicKey,
		nonceSym,
		sig,
	)
	suite.Require().Error(err)

	// Wrong nonce sym case
	err = tss.VerifyComplainSig(
		suite.member1.OneTimePubKey,
		suite.member2.OneTimePubKey,
		keySym,
		suite.fakeKey.PublicKey,
		sig,
	)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestGenerateMessageComplain() {
	keySym, err := tss.ComputeKeySym(suite.member1.OneTimePrivKey, suite.member2.OneTimePubKey)
	suite.Require().NoError(err)

	challenge := tss.GenerateMessageComplain(suite.member1.OneTimePubKey, suite.member2.OneTimePubKey, keySym)
	suite.Require().Equal(
		"726f756e6433436f6d706c61696e02e20b4d6bd3f10e7c3a9098c5832180b809a826ae49a972d5348758529c5015c502e20b4d6bd3f10e7c3a9098c5832180b809a826ae49a972d5348758529c5015c5035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26",
		hex.EncodeToString(challenge),
	)
}
