package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestComputeLagrangeCoefficient() {
	expValue := tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(1140))
	value := tss.ComputeLagrangeCoefficient(3, 20)

	suite.Require().Equal(expValue, value)
}
