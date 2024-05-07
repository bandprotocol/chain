package tss_test

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
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
					suite.Require().Less(j, int64(ide.ID))
					j = int64(ide.ID)
				}
			}
		})
	}
}

func (suite *TSSTestSuite) TestCloneEncSecretShares() {
	tests := []struct {
		name      string
		encShares tss.EncSecretShares
		wantErr   string
	}{
		{
			"size = 0",
			tss.EncSecretShares{},
			"",
		},
		{
			"size = 1",
			tss.EncSecretShares{
				testutil.HexDecode(
					"00bf89d839d9b4cbfea51435c7e49ac8696e6c1faf1715e1b343e62f90027d4b7ba8fb095282c02a43d59cd8e1a0708b",
				),
			},
			"",
		},
		{
			"size = 2",
			tss.EncSecretShares{
				testutil.HexDecode(
					"00bf89d839d9b4cbfea51435c7e49ac8696e6c1faf1715e1b343e62f90027d4b7ba8fb095282c02a43d59cd8e1a0708b",
				),
				testutil.HexDecode(
					"129636d592a2a9a90c96ae19c838b85d8a67cbcb29933f7189fe3518021df5cc6760ae358e712495017c254a28c236e1",
				),
			},
			"",
		},
		{
			"size = 3",
			tss.EncSecretShares{
				testutil.HexDecode(
					"129636d592a2a9a90c96ae19c838b85d8a67cbcb29933f7189fe3518021df5cc6760ae358e712495017c254a28c236e1",
				),
				testutil.HexDecode(
					"e51531c2d3458a9c439fed1838e82c27a94b17e5f9ef19a9deb47dbd4e10a18f312d3fa4380ac8be9b630a28dab9083d",
				),
				testutil.HexDecode(
					"d3ef4b34705961d66f2554f0007bc46942633793500be061db6076623d84bdb0a8457035649d8f85aff3632206f26a11",
				),
			},
			"",
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Clone: %s", t.name), func() {
			suite.Require().Equal(t.encShares, t.encShares.Clone())
		})
	}
}

func (suite *TSSTestSuite) TestValidateEncSecretShares() {
	tests := []struct {
		name      string
		encShares tss.EncSecretShares
		wantErr   string
	}{
		{
			"valid, size = 0",
			tss.EncSecretShares{},
			"",
		},
		{
			"valid, size = 1",
			tss.EncSecretShares{
				testutil.HexDecode(
					"00bf89d839d9b4cbfea51435c7e49ac8696e6c1faf1715e1b343e62f90027d4b7ba8fb095282c02a43d59cd8e1a0708b",
				),
			},
			"",
		},
		{
			"valid, size = 2",
			tss.EncSecretShares{
				testutil.HexDecode(
					"00bf89d839d9b4cbfea51435c7e49ac8696e6c1faf1715e1b343e62f90027d4b7ba8fb095282c02a43d59cd8e1a0708b",
				),
				testutil.HexDecode(
					"bebd586f83ed9038d8f6526e954c8e38ce70b9a4a012d4d3020edfbdd62b77966760ae358e712495017c254a28c236e1",
				),
			},
			"",
		},
		{
			"invalid because len(value) != 48",
			tss.EncSecretShares{
				testutil.HexDecode("6760ae358e712495017c254a28c236e16760ae358e712495017c254a28c236e1"),
			},
			"index 0 error: EncSecretShare: invalid size",
		},
	}

	for _, t := range tests {
		suite.Run(fmt.Sprintf("Clone: %s", t.name), func() {
			err := t.encShares.Validate()
			if t.wantErr != "" {
				suite.Require().EqualError(err, t.wantErr)
			}
		})
	}
}
