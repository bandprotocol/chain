package tss_test

import (
	"github.com/bandprotocol/chain/v3/pkg/tss"
)

func (suite *TSSTestSuite) TestConcatBytes() {
	res := tss.ConcatBytes([]byte("abc"), []byte("de"), []byte("fghi"))
	suite.Require().Equal([]byte("abcdefghi"), res)
}

func (suite *TSSTestSuite) TestGenerateKeyPairs() {
	kps, err := tss.GenerateKeyPairs(3)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(kps))

	for _, kp := range kps {
		pubKey := kp.PrivKey.Point()
		suite.Require().Equal(kp.PubKey, pubKey)
	}
}

func (suite *TSSTestSuite) TestGenerateKeyPair() {
	kp, err := tss.GenerateKeyPair()
	suite.Require().NoError(err)

	pubKey := kp.PrivKey.Point()
	suite.Require().Equal(kp.PubKey, pubKey)
}

func (suite *TSSTestSuite) TestPaddingBytes() {
	// Test padding with 0
	padded := tss.PaddingBytes([]byte{5, 1, 2, 3}, 8)
	suite.Require().Equal([]byte{0, 0, 0, 0, 5, 1, 2, 3}, padded)

	// Test padding with 1
	padded = tss.PaddingBytes([]byte{3, 1, 2}, 1)
	suite.Require().Equal([]byte{3, 1, 2}, padded)
}
