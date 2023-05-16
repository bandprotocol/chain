package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestComputeEncryptedSecretShares() {
	// n = 2
	// t = 2
	scalars, err := tss.ComputeEncryptedSecretShares(
		1,
		suite.kpI.PrivateKey,
		tss.PublicKeys{suite.kpI.PublicKey, suite.kpJ.PublicKey},
		suite.scalars,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(scalars))
}

func (suite *TSSTestSuite) TestEncryptSecretShares() {
	secret, err := hex.DecodeString("765613e9f908c8379b2b9e29ac0fda7b8bcc5e3fe34e010dc7a1491bfc0f96a1")
	suite.Require().NoError(err)

	keySym, err := hex.DecodeString("03bc213e4251592d29c070e4c31b980d150e755ec848afa4c06730ec1dcd09c482")
	suite.Require().NoError(err)

	encSecretShares, err := tss.EncryptSecretShares(tss.Scalars{secret}, tss.PublicKeys{keySym})
	suite.Require().NoError(err)
	suite.Require().
		Equal("0fd56e0e97d585edf49e82e3a7abe4ac9fc2bdd1b9681c840341849eb0eb852d", hex.EncodeToString(encSecretShares[0]))
}

func (suite *TSSTestSuite) TestComputeSecretShare() {
	secret := tss.ComputeSecretShare(suite.scalars, 2)
	suite.Require().
		Equal("765613e9f908c8379b2b9e29ac0fda7b8bcc5e3fe34e010dc7a1491bfc0f96a1", hex.EncodeToString(secret))
}
