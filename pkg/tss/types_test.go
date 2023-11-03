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
			scalar := tss.NewScalarFromModNScalar(t.scalar)
			suite.Require().Equal(t.expScalar, scalar)
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
			point := tss.NewPointFromJacobianPoint(&jacobianPoint)
			suite.Require().Equal(t.expPoint, point)
		})
	}
}

func (suite *TSSTestSuite) TestSortingCommitmentIDEs() {
	tests := []struct {
		name    string
		B       tss.CommitmentIDEList
		wantErr string
	}{
		{
			"sort([<5,...>,<1,...>])",
			tss.CommitmentIDEList{
				tss.CommitmentIDE{
					ID: 5,
					D:  testutil.HexDecode("02517a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad45"),
					E:  testutil.HexDecode("035eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b5"),
				},
				tss.CommitmentIDE{
					ID: 1,
					D:  testutil.HexDecode("02117a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad41"),
					E:  testutil.HexDecode("031eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b1"),
				},
			},
			"",
		},
		{
			"sort([<1,...>,<2,...>,<3,...>])",
			tss.CommitmentIDEList{
				tss.CommitmentIDE{
					ID: 1,
					D:  testutil.HexDecode("02117a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad41"),
					E:  testutil.HexDecode("031eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b1"),
				},
				tss.CommitmentIDE{
					ID: 2,
					D:  testutil.HexDecode("02217a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
					E:  testutil.HexDecode("032eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b2"),
				},
				tss.CommitmentIDE{
					ID: 3,
					D:  testutil.HexDecode("02317a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad43"),
					E:  testutil.HexDecode("033eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b3"),
				},
			},
			"",
		},
		{
			"error because there is a repeated element",
			tss.CommitmentIDEList{
				tss.CommitmentIDE{
					ID: 1,
					D:  testutil.HexDecode("02117a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad41"),
					E:  testutil.HexDecode("031eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b1"),
				},
				tss.CommitmentIDE{
					ID: 2,
					D:  testutil.HexDecode("02217a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
					E:  testutil.HexDecode("032eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b2"),
				},
				tss.CommitmentIDE{
					ID: 2,
					D:  testutil.HexDecode("02217a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad42"),
					E:  testutil.HexDecode("032eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b2"),
				},
				tss.CommitmentIDE{
					ID: 3,
					D:  testutil.HexDecode("02317a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad43"),
					E:  testutil.HexDecode("033eb31cfac1aeb0466f4c6fe98804b85bcca87a3a55c50340a04bf378830ed9b3"),
				},
			},
			"CommitmentIDEList: sorting fail because repeated element found at ID = 2",
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Sort: %s", t.name), func() {
			err := t.B.Sort()
			if t.wantErr != "" {
				suite.Require().EqualError(err, t.wantErr)
			} else {
				suite.Require().NoError(err)
				j := int64(-1)
				for _, ide := range t.B {
					suite.Require().Equal(true, j < int64(ide.ID))
					j = int64(ide.ID)
				}
			}
		})
	}
}
