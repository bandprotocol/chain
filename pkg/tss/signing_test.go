package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestComputeLagrangeCoefficient() {
	expValue := new(secp256k1.ModNScalar).SetInt(1140)

	value := tss.ComputeLagrangeCoefficient(3, 20).Bytes()
	scalarValue := new(secp256k1.ModNScalar)
	scalarValue.SetByteSlice(value)

	suite.Require().Equal(expValue, scalarValue)
}
