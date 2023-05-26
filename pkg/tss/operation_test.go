package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestComputeKeySym() {
	keySym, err := tss.ComputeKeySym(suite.member1.OneTimePrivKey, suite.member2.OneTimePubKey)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.keySyms[0], keySym)

	keySym, err = tss.ComputeKeySym(suite.member2.OneTimePrivKey, suite.member1.OneTimePubKey)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member2.keySyms[0], keySym)
}

func (suite *TSSTestSuite) TestComputeNonceSym() {
	nonceSym, err := tss.ComputeNonceSym(suite.nonce, suite.member1.OneTimePubKey)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.nonceSym, nonceSym)

	nonceSym, err = tss.ComputeNonceSym(suite.nonce, suite.member2.OneTimePubKey)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member2.nonceSym, nonceSym)
}

func (suite *TSSTestSuite) TestSumPoints() {
	// Prepare
	var p1, p2, expectedPoint secp256k1.JacobianPoint

	s1 := new(secp256k1.ModNScalar).SetInt(1)
	secp256k1.ScalarBaseMultNonConst(s1, &p1)

	s2 := new(secp256k1.ModNScalar).SetInt(2)
	secp256k1.ScalarBaseMultNonConst(s2, &p2)

	secp256k1.ScalarBaseMultNonConst(s1.Add(s2), &expectedPoint)

	// Try sum with function
	total, err := tss.SumPoints(tss.ParsePoint(&p1), tss.ParsePoint(&p2))
	suite.Require().NoError(err)
	suite.Require().Equal(tss.ParsePoint(&expectedPoint), total)
}

func (suite *TSSTestSuite) TestSumScalars() {
	total, err := tss.SumScalars(
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(1)),
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(2)),
		tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(3)),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(tss.ParseScalar(new(secp256k1.ModNScalar).SetInt(6)), total)
}
