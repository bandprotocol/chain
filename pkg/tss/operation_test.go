package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestComputeKeySym() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			keySym, err := tss.ComputeKeySym(
				memberI.OneTimePrivKey,
				memberJ.OneTimePubKey(),
			)

			suite.Require().NoError(err)
			suite.Require().Equal(memberI.KeySyms[testutil.GetSlot(memberI.ID, memberJ.ID)], keySym)
		},
	)
}

func (suite *TSSTestSuite) TestComputeNonceSym() {
	nonce := testutil.HexDecode("1111111111111111111111111111111111111111111111111111111111111111")
	pubKey := testutil.HexDecode("03c820245f18671206e752122953397786af3444f3a8da8098e594a8f612d94059")

	nonceSym, err := tss.ComputeNonceSym(nonce, pubKey)
	suite.Require().NoError(err)
	suite.Require().
		Equal("0360bf3f69810cc3472702c1f76ec76cdbefd85b1537db870e91b382a2e6e2bf6c", hex.EncodeToString(nonceSym))
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
