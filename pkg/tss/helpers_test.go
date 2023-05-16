package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestConcatBytes() {
	res := tss.ConcatBytes([]byte("abc"), []byte("de"), []byte("fghi"))
	suite.Require().Equal([]byte("abcdefghi"), res)
}

func (suite *TSSTestSuite) TestGenerateKeyPairs() {
	kps, err := tss.GenerateKeyPairs(3)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(kps))
}

func (suite *TSSTestSuite) TestGenerateKeyPair() {
	_, err := tss.GenerateKeyPair()
	suite.Require().NoError(err)
}
