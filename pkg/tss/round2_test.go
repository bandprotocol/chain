package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestComputeEncryptedSecretShares() {
	encSecretShares, err := tss.ComputeEncryptedSecretShares(
		1,
		suite.member1.OneTimePrivKey,
		tss.PublicKeys{suite.member1.OneTimePubKey, suite.member2.OneTimePubKey},
		suite.member1.Coefficients,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.encSecretShares, encSecretShares)
}

func (suite *TSSTestSuite) TestEncryptSecretShares() {
	secret := suite.member1.secretShares[0]
	keySym := suite.member1.keySyms[0]

	encSecretShares, err := tss.EncryptSecretShares(tss.Scalars{secret}, tss.PublicKeys{keySym})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.encSecretShares, encSecretShares)
}

func (suite *TSSTestSuite) TestComputeSecretShare() {
	secret, err := tss.ComputeSecretShare(suite.member1.Coefficients, 2)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.secretShares[0], secret)
}
