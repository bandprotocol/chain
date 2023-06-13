package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
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
	tests := []struct {
		name     string
		points   tss.Points
		expTotal tss.Point
		expError bool
	}{
		{
			"zero element",
			tss.Points{},
			testutil.HexDecode("020000000000000000000000000000000000000000000000000000000000000000"),
			false,
		},
		{
			"one element",
			tss.Points{testutil.HexDecode("0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")},
			testutil.HexDecode("0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"),
			false,
		},
		{
			"three element",
			tss.Points{
				testutil.HexDecode("0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"),
				testutil.HexDecode("02c6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5"),
			},
			testutil.HexDecode("02f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9"),
			false,
		},
		{
			"value is too big",
			tss.Points{
				testutil.HexDecode("02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
				testutil.HexDecode("02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
			},
			nil,
			true,
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			total, err := tss.SumPoints(t.points...)

			if t.expError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			suite.Require().Equal(t.expTotal, total)
		})
	}
}

func (suite *TSSTestSuite) TestSumScalars() {
	tests := []struct {
		name     string
		scalars  tss.Scalars
		expTotal tss.Scalar
		expError bool
	}{
		{
			"zero element",
			tss.Scalars{},
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000000"),
			false,
		},
		{
			"one element",
			tss.Scalars{testutil.HexDecode("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")},
			testutil.HexDecode("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"),
			false,
		},
		{
			"three element",
			tss.Scalars{
				testutil.HexDecode("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"),
				testutil.HexDecode("c6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5"),
			},
			testutil.HexDecode("3fc2e6133bca391985e5a304644787e0a464ae400b74c54545cc2c87a332753c"),
			false,
		},
		{
			"big values",
			tss.Scalars{
				testutil.HexDecode("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
				testutil.HexDecode("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
			},
			testutil.HexDecode("000000000000000000000000000000028aa24632a16ebf88805b42e45f9375de"),
			false,
		},
		{
			"length is too short",
			tss.Scalars{
				testutil.HexDecode("fffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
				testutil.HexDecode("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
			},
			nil,
			true,
		},
		{
			"length is too short",
			tss.Scalars{
				testutil.HexDecode("02fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
				testutil.HexDecode("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30"),
			},
			nil,
			true,
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			total, err := tss.SumScalars(t.scalars...)

			if t.expError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			suite.Require().Equal(t.expTotal, total)
		})
	}
}
