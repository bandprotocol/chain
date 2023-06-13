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
			"value 1",
			new(secp256k1.ModNScalar).SetInt(1),
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			"value 255",
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
			"invalid length - less length",
			testutil.HexDecode("00"),
			nil,
			tss.ErrInvalidLength,
		},
		{
			"invalid length - more length",
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
			"no element",
			tss.Scalars{},
			nil,
			nil,
		},
		{
			"one element",
			tss.Scalars{
				testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
			},
			[]*secp256k1.ModNScalar{
				new(secp256k1.ModNScalar).SetInt(1),
			},
			nil,
		},
		{
			"two element",
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
			"parse error - less length",
			tss.Scalars{
				testutil.HexDecode("00"),
			},
			nil,
			tss.ErrInvalidLength,
		},
		{
			"parse error - more length",
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

func (suite *TSSTestSuite) TestParsePrivateKey() {
	tests := []struct {
		name       string
		scalar     *secp256k1.ModNScalar
		expPrivKey tss.PrivateKey
	}{
		{
			"value 1",
			new(secp256k1.ModNScalar).SetInt(1),
			testutil.HexDecode("0000000000000000000000000000000000000000000000000000000000000001"),
		},
		{
			"value 255",
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
