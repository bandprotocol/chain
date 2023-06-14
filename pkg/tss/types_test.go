package tss_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func (suite *TSSTestSuite) TestParseScalar() {
	tests := []struct {
		name      string
		scalar    *secp256k1.ModNScalar
		expScalar tss.Scalar
	}{
		{
			"case 1",
			new(secp256k1.ModNScalar).SetInt(1),
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			"case 2",
			new(secp256k1.ModNScalar).SetInt(255),
			testutil.HexDecode("00000000000000000000000000000000000000000000000000000000000000ff"),
		},
	}

	for _, t := range tests {
		suite.Run(t.name, func() {
			scalar := tss.ParseScalar(t.scalar)
			suite.Require().Equal(t.expScalar, scalar)
		})
	}
}

func (suite *TSSTestSuite) TestScalar() {
	tests := []struct {
		name          string
		scalar        tss.Scalar
		expParse      *secp256k1.ModNScalar
		expParseError error
	}{
		{
			"success",
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			new(secp256k1.ModNScalar).SetInt(1),
			nil,
		},
		{
			"failed - invalid length - less length",
			testutil.HexDecode("00"),
			nil,
			tss.ErrInvalidLength,
		},
		{
			"failed - invalid length - more length",
			testutil.HexDecode("000000000000000000000000000000000000000000000000000000000000000001"),
			nil,
			tss.ErrInvalidLength,
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Parse: %s", t.name), func() {
			scalars, err := t.scalar.Parse()
			suite.Require().ErrorIs(err, t.expParseError)
			suite.Require().Equal(t.expParse, scalars)
		})
	}
}

func (suite *TSSTestSuite) TestScalars() {
	tests := []struct {
		name          string
		scalars       tss.Scalars
		expParse      []*secp256k1.ModNScalar
		expParseError error
	}{
		{
			"success - no element",
			tss.Scalars{},
			nil,
			nil,
		},
		{
			"success - one element",
			tss.Scalars{
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			},
			[]*secp256k1.ModNScalar{
				new(secp256k1.ModNScalar).SetInt(1),
			},
			nil,
		},
		{
			"success - two element",
			tss.Scalars{
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000002"),
			},
			[]*secp256k1.ModNScalar{
				new(secp256k1.ModNScalar).SetInt(1),
				new(secp256k1.ModNScalar).SetInt(2),
			},
			nil,
		},
		{
			"failed - less length",
			tss.Scalars{
				testutil.HexDecode("00"),
			},
			nil,
			tss.ErrInvalidLength,
		},
		{
			"failed - more length",
			tss.Scalars{
				testutil.HexDecode("000000000000000000000000000000000000000000000000000000000000000001"),
			},
			nil,
			tss.ErrInvalidLength,
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Parse: %s", t.name), func() {
			scalars, err := t.scalars.Parse()
			suite.Require().ErrorIs(err, t.expParseError)
			suite.Require().Equal(t.expParse, scalars)
		})
	}
}

func (suite *TSSTestSuite) TestParsePoint() {
	tests := []struct {
		name     string
		x        []byte
		y        []byte
		z        []byte
		expPoint tss.Point
	}{
		{
			"case 1 - z != 1",
			testutil.HexDecode("1a85b4aeb536706399da14007a8360c3253c3b90a4c151a9805816abc90f7055"),
			testutil.HexDecode("739a528f7105cd60f6e5c4b1bab2eaaa41aec6de84d62a2bee5461c4200c65d7"),
			testutil.HexDecode("bf538f07e087fb8ded6e09672b57e73fea3505b5055693ac3651ab30460db568"),
			testutil.HexDecode("03a50a76f243836311dd2fbaaf8b5185f5f7f34bd4cb99ac7309af18f89703960b"),
		},
		{
			"case 2 - z = 1",
			testutil.HexDecode("517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			testutil.HexDecode("4f2694a0f7f145aad7b9f722ce319d3a8145259c390e5596f79f0c17e0c8859a"),
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			testutil.HexDecode("02517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
		},
	}

	for _, t := range tests {
		suite.Run(t.name, func() {
			var x, y, z secp256k1.FieldVal
			x.SetByteSlice(t.x)
			y.SetByteSlice(t.y)
			z.SetByteSlice(t.z)
			jacobianPoint := secp256k1.MakeJacobianPoint(&x, &y, &z)
			point := tss.ParsePoint(&jacobianPoint)
			suite.Require().Equal(t.expPoint, point)
		})
	}
}

func (suite *TSSTestSuite) TestPoint() {
	tests := []struct {
		name     string
		point    tss.Point
		expX     []byte
		expY     []byte
		expZ     []byte
		expError error
	}{
		{
			"success",
			testutil.HexDecode("02517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			testutil.HexDecode("517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			testutil.HexDecode("4f2694a0f7f145aad7b9f722ce319d3a8145259c390e5596f79f0c17e0c8859a"),
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			nil,
		},
		{
			"failed - invalid length",
			testutil.HexDecode("517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			nil,
			nil,
			nil,
			tss.ErrParseError,
		},
		{
			"failed - wrong parity",
			testutil.HexDecode("01517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			nil,
			nil,
			nil,
			tss.ErrParseError,
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Parse: %s", t.name), func() {
			point, err := t.point.Parse()

			if t.expError != nil {
				suite.Require().ErrorIs(err, t.expError)
			} else {
				suite.Require().Equal(t.expX, point.X.Bytes()[:])
				suite.Require().Equal(t.expY, point.Y.Bytes()[:])
				suite.Require().Equal(t.expZ, point.Z.Bytes()[:])
			}
		})
	}
}

func (suite *TSSTestSuite) TestPoints() {
	tests := []struct {
		name     string
		points   tss.Points
		expXs    [][]byte
		expYs    [][]byte
		expZs    [][]byte
		expError error
	}{
		{
			"success - one element",
			tss.Points{
				testutil.HexDecode("02517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			},
			[][]byte{
				testutil.HexDecode("517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			},
			[][]byte{
				testutil.HexDecode("4f2694a0f7f145aad7b9f722ce319d3a8145259c390e5596f79f0c17e0c8859a"),
			},
			[][]byte{
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			},
			nil,
		},
		{
			"success - two elements",
			tss.Points{
				testutil.HexDecode("02517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
				testutil.HexDecode("036eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b1"),
			},
			[][]byte{
				testutil.HexDecode("517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
				testutil.HexDecode("6eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b1"),
			},
			[][]byte{
				testutil.HexDecode("4f2694a0f7f145aad7b9f722ce319d3a8145259c390e5596f79f0c17e0c8859a"),
				testutil.HexDecode("883b00bf3bf9136d6a6daf63915b2a66cb0fe8fc5f4d8129ab5834a61916500b"),
			},
			[][]byte{
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			},
			nil,
		},
		{
			"failed - invalid length",
			tss.Points{
				testutil.HexDecode("517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			},
			nil,
			nil,
			nil,
			tss.ErrParseError,
		},
		{
			"failed - wrong parity",
			tss.Points{
				testutil.HexDecode("01517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
			},
			nil,
			nil,
			nil,
			tss.ErrParseError,
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Parse: %s", t.name), func() {
			points, err := t.points.Parse()

			if t.expError != nil {
				suite.Require().ErrorIs(err, t.expError)
			} else {
				for i, p := range points {
					suite.Require().Equal(t.expXs[i], p.X.Bytes()[:])
					suite.Require().Equal(t.expYs[i], p.Y.Bytes()[:])
					suite.Require().Equal(t.expZs[i], p.Z.Bytes()[:])
				}
			}
		})
	}
}

func (suite *TSSTestSuite) TestParsePrivateKey() {
	tests := []struct {
		name       string
		scalar     *secp256k1.ModNScalar
		expPrivKey tss.PrivateKey
	}{
		{
			"case 1",
			new(secp256k1.ModNScalar).SetInt(1),
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			"case 255",
			new(secp256k1.ModNScalar).SetInt(255),
			testutil.HexDecode("00000000000000000000000000000000000000000000000000000000000000ff"),
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("from secp256k1.ModNScalar: %s", t.name), func() {
			privKey := tss.ParsePrivateKeyFromScalar(t.scalar)
			suite.Require().Equal(t.expPrivKey, privKey)
		})

		suite.Run(fmt.Sprintf("from secp256k1.PrivateKey: %s", t.name), func() {
			privKey := tss.ParsePrivateKey(secp256k1.NewPrivateKey(t.scalar))
			suite.Require().Equal(t.expPrivKey, privKey)
		})
	}
}
